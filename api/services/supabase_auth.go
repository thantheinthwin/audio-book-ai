package services

import (
	"audio-book-ai/api/config"
	"audio-book-ai/api/models"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SupabaseAuthService handles Supabase authentication
type SupabaseAuthService struct {
	cfg        *config.Config
	jwksClient *http.Client
	keys       map[string]*rsa.PublicKey
	lastFetch  time.Time
}

// JWKS represents the JSON Web Key Set structure
type JWKS struct {
	Keys []JWK `json:"keys"`
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

// SupabaseClaims represents the JWT claims from Supabase
type SupabaseClaims struct {
	jwt.RegisteredClaims
	Email        string                 `json:"email"`
	Phone        string                 `json:"phone"`
	AppMetadata  map[string]interface{} `json:"app_metadata"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
	Role         string                 `json:"role"`
	Aud          string                 `json:"aud"`
}

// NewSupabaseAuthService creates a new Supabase authentication service
func NewSupabaseAuthService(cfg *config.Config) *SupabaseAuthService {
	return &SupabaseAuthService{
		cfg:        cfg,
		jwksClient: &http.Client{Timeout: 10 * time.Second},
		keys:       make(map[string]*rsa.PublicKey),
	}
}

// ValidateToken validates a Supabase JWT token
func (s *SupabaseAuthService) ValidateToken(tokenString string) (*models.UserContext, error) {
	// Parse token without validation first to get the key ID
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &SupabaseClaims{})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// Get the key ID from the token header
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("kid not found in token header")
	}

	// Get the public key for this key ID
	publicKey, err := s.getPublicKey(kid)
	if err != nil {
		fmt.Println("error", err)
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}

	// Parse and validate the token with the public key
	parsedToken, err := jwt.ParseWithClaims(tokenString, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(*SupabaseClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate audience
	if claims.Aud != s.cfg.JWTAudience {
		return nil, fmt.Errorf("invalid audience: expected %s, got %s", s.cfg.JWTAudience, claims.Aud)
	}

	// Validate issuer if configured
	if s.cfg.SupabaseURL != "" && claims.Issuer != s.cfg.SupabaseURL {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", s.cfg.SupabaseURL, claims.Issuer)
	}

	// Extract role from app_metadata
	role := models.RoleUser
	if appMetadata, ok := claims.AppMetadata["role"].(string); ok && appMetadata != "" {
		role = appMetadata
	}

	// Create user context
	userContext := &models.UserContext{
		ID:    claims.Subject,
		Email: claims.Email,
		Aud:   claims.Aud,
		Role:  role,
		Token: tokenString,
	}

	return userContext, nil
}

// getPublicKey retrieves the public key for a given key ID
func (s *SupabaseAuthService) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check if we have the key cached
	if key, exists := s.keys[kid]; exists {
		return key, nil
	}

	// Fetch JWKS if we haven't recently
	if time.Since(s.lastFetch) > 5*time.Minute {
		if err := s.fetchJWKS(); err != nil {
			return nil, err
		}
	}

	// Check again after fetching
	if key, exists := s.keys[kid]; exists {
		return key, nil
	}

	return nil, fmt.Errorf("key with kid %s not found", kid)
}

// fetchJWKS fetches the JSON Web Key Set from Supabase
func (s *SupabaseAuthService) fetchJWKS() error {
	if s.cfg.SupabaseJWKSURL == "" {
		return fmt.Errorf("JWKS URL not configured")
	}

	resp, err := s.jwksClient.Get(s.cfg.SupabaseJWKSURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	// Clear existing keys and parse new ones
	s.keys = make(map[string]*rsa.PublicKey)
	for _, key := range jwks.Keys {
		if key.Kty == "RSA" && key.Use == "sig" {
			publicKey, err := s.parseRSAPublicKey(key.N, key.E)
			if err != nil {
				continue // Skip invalid keys
			}
			s.keys[key.Kid] = publicKey
		}
	}

	s.lastFetch = time.Now()
	return nil
}

// parseRSAPublicKey parses RSA public key from JWK components
func (s *SupabaseAuthService) parseRSAPublicKey(n, e string) (*rsa.PublicKey, error) {
	// Decode the modulus
	modulusBytes, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode the exponent
	exponentBytes, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert to big integers
	modulus := new(big.Int).SetBytes(modulusBytes)
	exponent := new(big.Int).SetBytes(exponentBytes)

	// Create RSA public key
	publicKey := &rsa.PublicKey{
		N: modulus,
		E: int(exponent.Int64()),
	}

	return publicKey, nil
}

// ExtractTokenFromHeader extracts the token from the Authorization header
func (s *SupabaseAuthService) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("authorization header required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("bearer token required")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("empty token")
	}

	return token, nil
}
