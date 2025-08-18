package models

import "time"

// UserContext represents the authenticated user context
type UserContext struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Aud   string `json:"aud"`
	Role  string `json:"role"`
	Token string `json:"-"`
}

// User represents a user in the system
type User struct {
	ID         string    `json:"id"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	IsActive   bool      `json:"is_active"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AuthRequest represents authentication request
type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	Username  string `json:"username" validate:"required,min=3"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// PasswordResetRequest represents password reset request
type PasswordResetRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// PasswordUpdateRequest represents password update request
type PasswordUpdateRequest struct {
	Password string `json:"password" validate:"required,min=6"`
}

// ProfileUpdateRequest represents profile update request
type ProfileUpdateRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}
