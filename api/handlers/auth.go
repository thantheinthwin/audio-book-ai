package handlers

import (
	"audio-book-ai/api/config"
	"audio-book-ai/api/models"

	"github.com/gofiber/fiber/v2"
	"github.com/supabase-community/gotrue-go"
	"github.com/supabase-community/gotrue-go/types"
)

// Register handles user registration
func Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Initialize Supabase client
	client := gotrue.New(config.New().SupabaseURL, config.New().SupabasePublishableKey)

	// Register user with Supabase
	signUpData := types.SignupRequest{
		Email:    req.Email,
		Password: req.Password,
		Data: map[string]interface{}{
			"username":   req.Username,
			"first_name": req.FirstName,
			"last_name":  req.LastName,
		},
	}

	user, err := client.Signup(signUpData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Registration failed: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"user": fiber.Map{
			"id":    user.User.ID,
			"email": user.User.Email,
		},
	})
}

// Login handles user authentication
func Login(c *fiber.Ctx) error {
	var req models.AuthRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Initialize Supabase client
	client := gotrue.New(config.New().SupabaseURL, config.New().SupabasePublishableKey)

	// Sign in with Supabase
	session, err := client.SignInWithEmailPassword(req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"user": fiber.Map{
			"id":    session.User.ID,
			"email": session.User.Email,
		},
		"access_token":  session.AccessToken,
		"refresh_token": session.RefreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600, // 1 hour
	})
}

// Logout handles user logout
func Logout(c *fiber.Ctx) error {
	// Get token from header
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No token provided",
		})
	}

	// Initialize Supabase client with token
	client := gotrue.New(config.New().SupabaseURL, config.New().SupabasePublishableKey).WithToken(token)

	// Sign out with Supabase
	err := client.Logout()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Logout failed",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Logout successful",
	})
}

// RefreshToken handles token refresh
func RefreshToken(c *fiber.Ctx) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Initialize Supabase client
	client := gotrue.New(config.New().SupabaseURL, config.New().SupabasePublishableKey)

	// Refresh token with Supabase
	session, err := client.RefreshToken(req.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid refresh token",
		})
	}

	return c.JSON(fiber.Map{
		"access_token":  session.AccessToken,
		"refresh_token": session.RefreshToken,
		"token_type":    "Bearer",
		"expires_in":    3600, // 1 hour
	})
}

// ForgotPassword handles password reset request
func ForgotPassword(c *fiber.Ctx) error {
	var req models.PasswordResetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Initialize Supabase client
	client := gotrue.New(config.New().SupabaseURL, config.New().SupabasePublishableKey)

	// Send password reset email
	err := client.Recover(types.RecoverRequest{
		Email: req.Email,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send reset email",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password reset email sent",
	})
}

// ResetPassword handles password reset
func ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Initialize Supabase client
	client := gotrue.New(config.New().SupabaseURL, config.New().SupabasePublishableKey)

	// Update password
	password := req.Password
	_, err := client.UpdateUser(types.UpdateUserRequest{
		Password: &password,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to reset password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password reset successful",
	})
}
