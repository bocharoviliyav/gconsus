# Keycloak Configuration

This directory contains Keycloak realm configuration for the Git Info application.

## Realm: git-info

The realm includes:

### Roles
- **admin** - Full system access (can manage teams, users, configurations)
- **manager** - Team and analytics management (can create teams, view analytics)
- **user** - Read-only access (can view own metrics and dashboards)

### Clients

#### git-info-backend
- **Type**: Confidential client (Bearer only)
- **Purpose**: Backend API authentication
- **Authentication**: Client secret
- **Token lifespan**: 1 hour

#### git-info-frontend
- **Type**: Public client
- **Purpose**: Frontend React application
- **Authentication**: PKCE flow
- **Redirect URIs**: 
  - `http://localhost:5173/*` (Vite dev)
  - `http://localhost:3000/*` (alternative port)
  - `http://frontend:80/*` (Docker)

### Test Users

| Username | Password | Roles | Description |
|----------|----------|-------|-------------|
| admin | admin123 | admin, manager, user | Full access |
| manager | manager123 | manager, user | Team management |
| testuser | user123 | user | Read-only access |

⚠️ **Security Note**: Change these passwords in production!

## Import Instructions

### Option 1: Docker Compose (Automatic)

The realm will be automatically imported when starting Keycloak via docker-compose:

```bash
make up
```

The docker-compose.yml mounts the realm export file and Keycloak imports it on first start.

### Option 2: Manual Import via Admin Console

1. Access Keycloak Admin Console:
   ```
   http://localhost:8090
   ```

2. Login with admin credentials (from docker-compose.yml):
   - Username: `admin`
   - Password: `admin`

3. Navigate to realm selection dropdown (top-left)

4. Click "Create Realm"

5. Click "Browse" and select `realm-export.json`

6. Click "Create"

### Option 3: CLI Import

Using Keycloak CLI:

```bash
docker exec -it keycloak /opt/keycloak/bin/kc.sh import \
  --file /opt/keycloak/data/import/realm-export.json
```

## Configuration

### Environment Variables for Backend

Add these to your backend `.env` or docker-compose:

```env
KEYCLOAK_REALM_URL=http://keycloak:8080/realms/git-info
KEYCLOAK_CLIENT_ID=git-info-backend
KEYCLOAK_CLIENT_SECRET=<your-client-secret>
```

To get the client secret:
1. Go to Keycloak Admin Console
2. Select "git-info" realm
3. Go to Clients → git-info-backend
4. Navigate to "Credentials" tab
5. Copy the client secret

### Frontend Configuration

Add these to your frontend `.env`:

```env
VITE_KEYCLOAK_URL=http://localhost:8090
VITE_KEYCLOAK_REALM=git-info
VITE_KEYCLOAK_CLIENT_ID=git-info-frontend
```

## Token Structure

### Access Token Claims

```json
{
  "sub": "user-uuid",
  "email": "user@example.com",
  "preferred_username": "username",
  "name": "First Last",
  "given_name": "First",
  "family_name": "Last",
  "realm_access": {
    "roles": ["admin", "manager", "user"]
  }
}
```

## API Endpoints

Backend middleware validates tokens from:
```
Authorization: Bearer <access_token>
```

The middleware extracts:
- User ID from `sub` claim
- Username from `preferred_username`
- Roles from `realm_access.roles`

## Role-Based Access Control

### Endpoint Access Matrix

| Endpoint | Admin | Manager | User |
|----------|-------|---------|------|
| GET /api/v1/analytics/* | ✓ | ✓ | ✓ |
| GET /api/v1/teams | ✓ | ✓ | ✓ |
| POST /api/v1/teams | ✓ | ✓ | ✗ |
| PUT /api/v1/teams/{id} | ✓ | ✓ (own teams) | ✗ |
| DELETE /api/v1/teams/{id} | ✓ | ✗ | ✗ |
| GET /api/v1/users | ✓ | ✓ | ✗ |
| POST /api/v1/users | ✓ | ✗ | ✗ |
| POST /api/v1/users/sync | ✓ | ✗ | ✗ |
| PUT /api/v1/config/* | ✓ | ✗ | ✗ |

## Troubleshooting

### Token Validation Fails

1. Check Keycloak is running: `curl http://localhost:8090`
2. Verify realm URL in backend config
3. Check token expiration (default: 1 hour)
4. Verify client secret matches

### Cannot Login

1. Check user exists and is enabled
2. Verify client redirect URIs include your frontend URL
3. Check browser console for CORS errors
4. Ensure Keycloak web origins include frontend URL

### Role Not Working

1. Check token claims using jwt.io
2. Verify role is assigned to user
3. Check realm_access.roles array in token
4. Verify middleware extracts roles correctly

## Production Recommendations

1. **Use HTTPS**: Set `sslRequired: "all"` in realm config
2. **Change Passwords**: Update all default user passwords
3. **Rotate Secrets**: Regenerate client secrets
4. **Token Lifespan**: Consider shorter lifespans (15-30 min)
5. **Brute Force**: Already enabled with 5 attempts limit
6. **Email Verification**: Enable for user registration
7. **2FA**: Consider enabling OTP for admin users
8. **Session Limits**: Review and adjust session timeouts
9. **Audit Logs**: Enable event listeners for security auditing
10. **Backup**: Regularly export realm configuration

## References

- [Keycloak Documentation](https://www.keycloak.org/documentation)
- [JWT.io Debugger](https://jwt.io)
- [OpenID Connect Specification](https://openid.net/connect/)
