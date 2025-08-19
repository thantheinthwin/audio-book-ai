# Supabase Authentication Integration

This API has been refactored to use Supabase's authentication system with proper JWT validation using JWKS (JSON Web Key Set).

## Overview

The authentication system now follows Supabase's best practices:

1. **JWKS Validation**: Uses Supabase's JWKS endpoint to validate JWT tokens
2. **RSA Key Verification**: Properly validates RSA-signed JWT tokens
3. **Audience & Issuer Validation**: Validates token audience and issuer
4. **Role-based Access Control**: Extracts user roles from JWT claims

## Configuration

### Environment Variables

Add these environment variables to your `.env` file:

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_PUBLISHABLE_KEY=your-publishable-key
SUPABASE_SECRET_KEY=your-secret-key
SUPABASE_JWKS_URL=https://your-project.supabase.co/auth/v1/keys
SUPABASE_STORAGE_BUCKET=audio

# JWT Configuration
JWT_AUDIENCE=authenticated
```

### Getting Supabase Configuration

1. Go to your Supabase project dashboard
2. Navigate to Settings > API
3. Copy the following values:
   - Project URL → `SUPABASE_URL`
   - Project API keys (anon public) → `SUPABASE_PUBLISHABLE_KEY`
   - Project API keys (service_role secret) → `SUPABASE_SECRET_KEY`
   - JWKS URL → `SUPABASE_JWKS_URL` (usually `{SUPABASE_URL}/auth/v1/keys`)

## Authentication Flow

### 1. Token Validation

The API validates JWT tokens using the following process:

1. Extract token from `Authorization: Bearer <token>` header
2. Parse token header to get the key ID (`kid`)
3. Fetch the corresponding public key from Supabase's JWKS
4. Validate the token signature using the public key
5. Verify token claims (audience, issuer, expiration)
6. Extract user information and role from claims

### 2. User Context

After successful validation, the API creates a `UserContext` with:

```go
type UserContext struct {
    ID    string `json:"id"`      // User ID from JWT 'sub' claim
    Email string `json:"email"`   // User email from JWT 'email' claim
    Aud   string `json:"aud"`     // Audience from JWT 'aud' claim
    Role  string `json:"role"`    // Role from JWT 'app_metadata.role'
    Token string `json:"-"`       // Original JWT token
}
```

### 3. Role-based Access Control

The API supports role-based access control:

- **User Role**: `user` - Basic user permissions
- **Admin Role**: `admin` - Administrative permissions

Roles are extracted from the JWT `app_metadata.role` claim. If no role is specified, it defaults to `user`.

## API Endpoints

### Authentication Endpoints

- `POST /auth/validate` - Validate a JWT token
- `GET /auth/me` - Get current user profile

### Protected Endpoints

All protected endpoints require a valid JWT token in the `Authorization` header:

```
Authorization: Bearer <jwt-token>
```

### Middleware

The API provides several middleware functions:

- `AuthMiddleware` - Requires valid authentication
- `OptionalAuthMiddleware` - Optional authentication
- `RequireRole(role)` - Requires specific role
- `RequireAdmin()` - Requires admin role
- `RequireUser()` - Requires user or admin role

## Error Handling

The authentication system returns appropriate HTTP status codes:

- `401 Unauthorized` - Invalid or missing token
- `403 Forbidden` - Insufficient permissions
- `500 Internal Server Error` - JWKS fetch errors

## Security Features

1. **Key Rotation**: Automatically fetches new keys from JWKS every 5 minutes
2. **Token Expiration**: Validates token expiration time
3. **Audience Validation**: Ensures tokens are intended for this API
4. **Issuer Validation**: Validates token issuer (if configured)
5. **Signature Verification**: Uses RSA public keys for verification

## Testing

To test the authentication:

1. Get a JWT token from your Supabase client
2. Include it in the Authorization header:
   ```bash
   curl -H "Authorization: Bearer <your-jwt-token>" \
        http://localhost:8080/auth/me
   ```

## Troubleshooting

### Common Issues

1. **"kid not found in token header"**

   - Ensure you're using a valid Supabase JWT token
   - Check that the token is not malformed

2. **"key with kid X not found"**

   - The JWKS endpoint might be temporarily unavailable
   - Check your `SUPABASE_JWKS_URL` configuration

3. **"invalid audience"**

   - Ensure `JWT_AUDIENCE` matches the token's audience claim
   - Default audience is `authenticated`

4. **"invalid issuer"**
   - Ensure `JWT_ISSUER` matches your Supabase project URL
   - Format: `https://your-project.supabase.co`

### Debug Mode

Enable debug logging by setting `NODE_ENV=development` to see detailed authentication logs.


