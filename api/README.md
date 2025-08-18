# Audio Book AI API

This is the Golang backend API for the Audio Book AI application. The API is designed to work seamlessly with the Next.js frontend that uses Supabase for authentication.

## Authentication Flow

The API uses Supabase JWT tokens for authentication. Here's how it works:

1. **Frontend Authentication**: The Next.js frontend handles all authentication (login, signup, password reset) through Supabase
2. **Token Validation**: The backend validates Supabase JWT tokens and extracts user information
3. **Protected Routes**: API endpoints that require authentication use the `AuthMiddleware`
4. **Optional Authentication**: Some routes can work with or without authentication using `OptionalAuthMiddleware`

## API Endpoints

### Authentication Endpoints

- `POST /auth/validate` - Validate a Supabase JWT token
- `GET /auth/me` - Get current user information (requires authentication)
- `GET /auth/health` - Health check endpoint

### Protected Endpoints

All protected endpoints require a valid Supabase JWT token in the Authorization header:

```
Authorization: Bearer <supabase-jwt-token>
```

- `GET /profile` - Get user profile
- `PUT /profile` - Update user profile
- `DELETE /profile` - Delete user profile
- `GET /audiobooks` - Get user's audio books
- `POST /audiobooks` - Create new audio book
- And many more...

### Optional Authentication Endpoints

These endpoints work with or without authentication:

- `GET /public/audiobooks` - Get public audio books
- `GET /public/audiobooks/:id` - Get specific public audio book

## Environment Variables

Required environment variables:

```env
# Supabase Configuration
SUPABASE_URL=your_supabase_url
SUPABASE_PUBLISHABLE_KEY=your_supabase_publishable_key
SUPABASE_SECRET_KEY=your_supabase_secret_key

# Server Configuration
API_PORT=8080
NODE_ENV=development

# CORS Configuration
CORS_ORIGIN=http://localhost:3000
```

## Running the API

1. Install dependencies:

   ```bash
   go mod tidy
   ```

2. Set up environment variables (see `.env.example`)

3. Run the server:
   ```bash
   go run main.go
   ```

## Integration with Next.js Frontend

The frontend should include the Supabase JWT token in API requests:

```typescript
// Example of making an authenticated API request
const supabase = createClient();
const {
  data: { session },
} = await supabase.auth.getSession();

const response = await fetch("/api/profile", {
  headers: {
    Authorization: `Bearer ${session?.access_token}`,
    "Content-Type": "application/json",
  },
});
```

## Token Validation

The API validates Supabase JWT tokens by:

1. Checking the Authorization header format
2. Parsing the JWT token
3. Validating the signature using the Supabase publishable key
4. Extracting user information from the token claims

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "Error message",
  "message": "Optional message",
  "user": {
    /* user object if applicable */
  }
}
```

## Development

- The API uses Fiber as the web framework
- JWT validation is handled by the `golang-jwt/jwt` package
- CORS is configured to work with the Next.js frontend
- Environment-specific configuration is handled by the config package
