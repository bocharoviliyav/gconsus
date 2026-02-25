package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	// UserContextKey is the context key for storing user claims
	UserContextKey contextKey = "user"
	// RolesContextKey is the context key for storing user roles
	RolesContextKey contextKey = "roles"
)

// KeycloakConfig holds configuration for Keycloak middleware
type KeycloakConfig struct {
	RealmURL  string // e.g., http://keycloak:8080/realms/gconsus
	ClientID  string // e.g., gconsus-backend
	CertsURL  string // optional, defaults to RealmURL/protocol/openid-connect/certs
	CacheTTL  time.Duration
	SkipPaths []string // paths that don't require authentication
}

// KeycloakMiddleware validates JWT tokens from Keycloak
type KeycloakMiddleware struct {
	config     KeycloakConfig
	keysCache  *keysCache
	httpClient *http.Client
}

// keysCache caches public keys from Keycloak
type keysCache struct {
	mu         sync.RWMutex
	keys       map[string]*rsa.PublicKey
	lastUpdate time.Time
	ttl        time.Duration
}

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKSet represents a set of JWKs
type JWKSet struct {
	Keys []JWK `json:"keys"`
}

// UserClaims represents the claims extracted from JWT token
type UserClaims struct {
	Sub               string                 `json:"sub"`
	Email             string                 `json:"email"`
	PreferredUsername string                 `json:"preferred_username"`
	Name              string                 `json:"name"`
	GivenName         string                 `json:"given_name"`
	FamilyName        string                 `json:"family_name"`
	RealmRoles        []string               `json:"realm_access.roles"`
	ResourceAccess    map[string]interface{} `json:"resource_access"`
}

// NewKeycloakMiddleware creates a new Keycloak JWT middleware
func NewKeycloakMiddleware(config KeycloakConfig) *KeycloakMiddleware {
	if config.CertsURL == "" {
		config.CertsURL = config.RealmURL + "/protocol/openid-connect/certs"
	}
	if config.CacheTTL == 0 {
		config.CacheTTL = 5 * time.Minute
	}

	return &KeycloakMiddleware{
		config: config,
		keysCache: &keysCache{
			keys: make(map[string]*rsa.PublicKey),
			ttl:  config.CacheTTL,
		},
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Middleware returns an HTTP middleware function
func (km *KeycloakMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path should skip authentication
		for _, skipPath := range km.config.SkipPaths {
			if strings.HasPrefix(r.URL.Path, skipPath) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			km.unauthorized(w, "missing authorization header")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			km.unauthorized(w, "invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		claims, err := km.validateToken(tokenString)
		if err != nil {
			slog.Error("Token validation failed", "error", err)
			km.unauthorized(w, "invalid token")
			return
		}

		// Extract roles
		roles := km.extractRoles(claims)

		// Add claims and roles to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		ctx = context.WithValue(ctx, RolesContextKey, roles)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateToken validates JWT token and returns claims
func (km *KeycloakMiddleware) validateToken(tokenString string) (*UserClaims, error) {
	// Parse token without validation first to get kid
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("missing kid in token header")
	}

	// Get public key
	publicKey, err := km.getPublicKey(kid)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse and validate token
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !parsedToken.Valid {
		return nil, errors.New("token is invalid")
	}

	// Extract claims
	mapClaims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("failed to extract claims")
	}

	claims := &UserClaims{}

	if sub, ok := mapClaims["sub"].(string); ok {
		claims.Sub = sub
	}
	if email, ok := mapClaims["email"].(string); ok {
		claims.Email = email
	}
	if username, ok := mapClaims["preferred_username"].(string); ok {
		claims.PreferredUsername = username
	}
	if name, ok := mapClaims["name"].(string); ok {
		claims.Name = name
	}
	if givenName, ok := mapClaims["given_name"].(string); ok {
		claims.GivenName = givenName
	}
	if familyName, ok := mapClaims["family_name"].(string); ok {
		claims.FamilyName = familyName
	}

	// Extract realm roles
	if realmAccess, ok := mapClaims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, role := range roles {
				if roleStr, ok := role.(string); ok {
					claims.RealmRoles = append(claims.RealmRoles, roleStr)
				}
			}
		}
	}

	// Extract resource access
	if resourceAccess, ok := mapClaims["resource_access"].(map[string]interface{}); ok {
		claims.ResourceAccess = resourceAccess
	}

	return claims, nil
}

// getPublicKey retrieves the public key for the given kid
func (km *KeycloakMiddleware) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache
	km.keysCache.mu.RLock()
	if key, ok := km.keysCache.keys[kid]; ok && time.Since(km.keysCache.lastUpdate) < km.keysCache.ttl {
		km.keysCache.mu.RUnlock()
		return key, nil
	}
	km.keysCache.mu.RUnlock()

	// Fetch keys from Keycloak
	if err := km.refreshKeys(); err != nil {
		return nil, err
	}

	// Try again from cache
	km.keysCache.mu.RLock()
	defer km.keysCache.mu.RUnlock()

	key, ok := km.keysCache.keys[kid]
	if !ok {
		return nil, fmt.Errorf("key with kid %s not found", kid)
	}

	return key, nil
}

// refreshKeys fetches public keys from Keycloak
func (km *KeycloakMiddleware) refreshKeys() error {
	km.keysCache.mu.Lock()
	defer km.keysCache.mu.Unlock()

	// Double-check if another goroutine already refreshed
	if time.Since(km.keysCache.lastUpdate) < km.keysCache.ttl {
		return nil
	}

	resp, err := km.httpClient.Get(km.config.CertsURL)
	if err != nil {
		return fmt.Errorf("failed to fetch keys: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch keys: status %d", resp.StatusCode)
	}

	var jwks JWKSet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode keys: %w", err)
	}

	// Convert JWKs to RSA public keys
	newKeys := make(map[string]*rsa.PublicKey)
	for _, jwk := range jwks.Keys {
		if jwk.Kty != "RSA" || jwk.Use != "sig" {
			continue
		}

		pubKey, err := km.jwkToRSAPublicKey(jwk)
		if err != nil {
			slog.Warn("Failed to convert JWK to RSA public key", "kid", jwk.Kid, "error", err)
			continue
		}

		newKeys[jwk.Kid] = pubKey
	}

	km.keysCache.keys = newKeys
	km.keysCache.lastUpdate = time.Now()

	slog.Info("Refreshed Keycloak public keys", "count", len(newKeys))

	return nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func (km *KeycloakMiddleware) jwkToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

// extractRoles extracts roles from claims
func (km *KeycloakMiddleware) extractRoles(claims *UserClaims) []string {
	roles := make([]string, 0)

	// Add realm roles
	roles = append(roles, claims.RealmRoles...)

	// Add client-specific roles
	if claims.ResourceAccess != nil {
		if clientAccess, ok := claims.ResourceAccess[km.config.ClientID].(map[string]interface{}); ok {
			if clientRoles, ok := clientAccess["roles"].([]interface{}); ok {
				for _, role := range clientRoles {
					if roleStr, ok := role.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		}
	}

	return roles
}

// unauthorized sends an unauthorized response
func (km *KeycloakMiddleware) unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

// GetUserClaims retrieves user claims from context
func GetUserClaims(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*UserClaims)
	return claims, ok
}

// GetUserRoles retrieves user roles from context
func GetUserRoles(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(RolesContextKey).([]string)
	return roles, ok
}

// HasRole checks if user has a specific role
func HasRole(ctx context.Context, role string) bool {
	roles, ok := GetUserRoles(ctx)
	if !ok {
		return false
	}

	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func HasAnyRole(ctx context.Context, requiredRoles ...string) bool {
	for _, role := range requiredRoles {
		if HasRole(ctx, role) {
			return true
		}
	}
	return false
}
