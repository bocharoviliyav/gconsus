package middleware

import (
	"encoding/json"
	"net/http"
)

// Role constants
const (
	RoleAdmin   = "admin"
	RoleManager = "manager"
	RoleUser    = "user"
)

// RequireRole creates a middleware that requires the user to have a specific role
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !HasRole(r.Context(), role) {
				forbidden(w, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole creates a middleware that requires the user to have at least one of the specified roles
func RequireAnyRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !HasAnyRole(r.Context(), roles...) {
				forbidden(w, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllRoles creates a middleware that requires the user to have all of the specified roles
func RequireAllRoles(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, role := range roles {
				if !HasRole(r.Context(), role) {
					forbidden(w, "insufficient permissions")
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin is a convenience middleware for requiring admin role
func RequireAdmin(next http.Handler) http.Handler {
	return RequireRole(RoleAdmin)(next)
}

// RequireManagerOrAdmin is a convenience middleware for requiring manager or admin role
func RequireManagerOrAdmin(next http.Handler) http.Handler {
	return RequireAnyRole(RoleManager, RoleAdmin)(next)
}

// RequireAuthentication is a middleware that just checks if user is authenticated (has any role)
func RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		roles, ok := GetUserRoles(r.Context())
		if !ok || len(roles) == 0 {
			forbidden(w, "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// forbidden sends a forbidden response
func forbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}
