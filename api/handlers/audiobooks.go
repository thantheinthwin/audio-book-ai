package handlers

import (
	"audio-book-ai/api/models"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetAudioBooks returns a list of audiobooks for the authenticated user
// GET /audiobooks
func (h *Handler) GetAudioBooks(c *fiber.Ctx) error {
	log.Printf("GetAudioBooks: Request received from IP %s", c.IP())
	log.Printf("GetAudioBooks: User agent: %s", c.Get("User-Agent"))

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("GetAudioBooks: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	log.Printf("GetAudioBooks: User authenticated - UserID: %s", userCtx.ID)
	userID := uuid.MustParse(userCtx.ID)

	// Parse query parameters
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Cap at 100 items per page
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	log.Printf("GetAudioBooks: Fetching audiobooks for user %s with limit=%d, offset=%d", userID, limit, offset)

	// Get audiobooks from database
	audiobooks, total, err := h.repo.ListAudioBooks(context.Background(), limit, offset, nil)
	if err != nil {
		log.Printf("GetAudioBooks: Failed to fetch audiobooks: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch audiobooks",
		})
	}

	log.Printf("GetAudioBooks: Found %d audiobooks (total: %d)", len(audiobooks), total)

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	currentPage := (offset / limit) + 1

	return c.JSON(fiber.Map{
		"data": audiobooks,
		"pagination": fiber.Map{
			"total":        total,
			"limit":        limit,
			"offset":       offset,
			"current_page": currentPage,
			"total_pages":  totalPages,
		},
	})
}

// GetAudioBook returns a specific audiobook by ID
// GET /audiobooks/:id
func (h *Handler) GetAudioBook(c *fiber.Ctx) error {
	log.Printf("GetAudioBook: Request received from IP %s", c.IP())
	log.Printf("GetAudioBook: User agent: %s", c.Get("User-Agent"))

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("GetAudioBook: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("GetAudioBook: Fetching audiobook with ID: %s", audiobookID)

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("GetAudioBook: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("GetAudioBook: User authenticated - UserID: %s", userID)

	// Get audiobook with details from database
	audiobook, err := h.repo.GetAudioBookWithDetails(context.Background(), audiobookID)
	if err != nil {
		log.Printf("GetAudioBook: Failed to fetch audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Check if user owns this audiobook or if it's public
	// if audiobook.CreatedBy != userID && !audiobook.IsPublic {
	// 	log.Printf("GetAudioBook: Access denied - user does not own audiobook and it's not public")
	// 	return c.Status(http.StatusForbidden).JSON(fiber.Map{
	// 		"error": "Access denied",
	// 	})
	// }

	log.Printf("GetAudioBook: Successfully retrieved audiobook: %s", audiobook.Title)

	return c.JSON(fiber.Map{
		"data": audiobook,
	})
}

// UpdateAudioBook updates an existing audiobook
// PUT /audiobooks/:id
func (h *Handler) UpdateAudioBook(c *fiber.Ctx) error {
	log.Printf("UpdateAudioBook: Request received from IP %s", c.IP())
	log.Printf("UpdateAudioBook: User agent: %s", c.Get("User-Agent"))

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("UpdateAudioBook: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("UpdateAudioBook: Updating audiobook with ID: %s", audiobookID)

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("UpdateAudioBook: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("UpdateAudioBook: User authenticated - UserID: %s", userID)

	// Get existing audiobook
	existingAudiobook, err := h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("UpdateAudioBook: Failed to fetch audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Check if user owns this audiobook
	if existingAudiobook.CreatedBy != userID {
		log.Printf("UpdateAudioBook: Access denied - user does not own audiobook")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Parse request body
	var req models.UpdateAudioBookRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("UpdateAudioBook: Failed to parse request body: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		log.Printf("UpdateAudioBook: Validation failed: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
		})
	}

	// Update fields if provided
	if req.Title != nil {
		existingAudiobook.Title = *req.Title
	}
	if req.Author != nil {
		existingAudiobook.Author = *req.Author
	}
	if req.Language != nil {
		existingAudiobook.Language = *req.Language
	}
	if req.IsPublic != nil {
		existingAudiobook.IsPublic = *req.IsPublic
	}
	if req.Price != nil {
		existingAudiobook.Price = *req.Price
	}
	if req.CoverImageURL != nil {
		existingAudiobook.CoverImageURL = req.CoverImageURL
	}

	// Update in database
	if err := h.repo.UpdateAudioBook(context.Background(), existingAudiobook); err != nil {
		log.Printf("UpdateAudioBook: Failed to update audiobook: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update audiobook",
		})
	}

	log.Printf("UpdateAudioBook: Successfully updated audiobook: %s", existingAudiobook.Title)

	return c.JSON(fiber.Map{
		"data": existingAudiobook,
	})
}

// Cart handlers

// AddToCart adds an audiobook to the user's cart
// POST /user/cart
func (h *Handler) AddToCart(c *fiber.Ctx) error {
	log.Printf("AddToCart: Request received from IP %s", c.IP())

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("AddToCart: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("AddToCart: User authenticated - UserID: %s", userID)

	// Parse request body
	var req models.AddToCartRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("AddToCart: Failed to parse request body: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		log.Printf("AddToCart: Validation failed: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
		})
	}

	// Check if audiobook exists and is public
	// audiobook, err := h.repo.GetAudioBookByID(context.Background(), req.AudiobookID)
	// if err != nil {
	// 	log.Printf("AddToCart: Failed to get audiobook: %v", err)
	// 	return c.Status(http.StatusNotFound).JSON(fiber.Map{
	// 		"error": "Audiobook not found",
	// 	})
	// }

	// if !audiobook.IsPublic {
	// 	log.Printf("AddToCart: Audiobook is not public")
	// 	return c.Status(http.StatusForbidden).JSON(fiber.Map{
	// 		"error": "Cannot add private audiobook to cart",
	// 	})
	// }

	// Add to cart
	cartItemID, err := h.repo.AddToCart(context.Background(), userID, req.AudiobookID)
	if err != nil {
		log.Printf("AddToCart: Failed to add to cart: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add to cart",
		})
	}

	log.Printf("AddToCart: Successfully added audiobook %s to cart for user %s with cart item ID %s", req.AudiobookID, userID, cartItemID)

	return c.JSON(fiber.Map{
		"message": "Added to cart successfully",
		"data": fiber.Map{
			"cart_item_id": cartItemID,
			"audiobook_id": req.AudiobookID,
		},
	})
}

// RemoveFromCart removes an audiobook from the user's cart
// DELETE /user/cart/:audiobookId
func (h *Handler) RemoveFromCart(c *fiber.Ctx) error {
	log.Printf("RemoveFromCart: Request received from IP %s", c.IP())

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("RemoveFromCart: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("RemoveFromCart: User authenticated - UserID: %s", userID)

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("audiobookId"))
	if err != nil {
		log.Printf("RemoveFromCart: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	// Remove from cart
	if err := h.repo.RemoveFromCart(context.Background(), userID, audiobookID); err != nil {
		log.Printf("RemoveFromCart: Failed to remove from cart: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove from cart",
		})
	}

	log.Printf("RemoveFromCart: Successfully removed audiobook %s from cart for user %s", audiobookID, userID)

	return c.JSON(fiber.Map{
		"message": "Removed from cart successfully",
	})
}

// GetCart returns the user's cart items
// GET /user/cart
func (h *Handler) GetCart(c *fiber.Ctx) error {
	log.Printf("GetCart: Request received from IP %s", c.IP())

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("GetCart: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("GetCart: User authenticated - UserID: %s", userID)

	// Get cart items
	cartItems, err := h.repo.GetCartItems(context.Background(), userID)
	if err != nil {
		log.Printf("GetCart: Failed to get cart items: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get cart items",
		})
	}

	// Calculate total price
	totalPrice := 0.0
	for _, item := range cartItems {
		totalPrice += item.AudioBook.Price
	}

	response := models.CartResponse{
		Items:      cartItems,
		TotalItems: len(cartItems),
		TotalPrice: totalPrice,
	}

	log.Printf("GetCart: Successfully retrieved %d cart items for user %s", len(cartItems), userID)

	return c.JSON(fiber.Map{
		"data": response,
	})
}

// IsInCart checks if an audiobook is in the user's cart
// GET /user/cart/:audiobookId/check
func (h *Handler) IsInCart(c *fiber.Ctx) error {
	log.Printf("IsInCart: Request received from IP %s", c.IP())

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("IsInCart: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("IsInCart: User authenticated - UserID: %s", userID)

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("audiobookId"))
	if err != nil {
		log.Printf("IsInCart: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	// Check if in cart
	isInCart, err := h.repo.IsInCart(context.Background(), userID, audiobookID)
	if err != nil {
		log.Printf("IsInCart: Failed to check if in cart: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check cart status",
		})
	}

	log.Printf("IsInCart: Audiobook %s is in cart for user %s: %v", audiobookID, userID, isInCart)

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"is_in_cart": isInCart,
		},
	})
}

// Checkout processes the checkout of cart items
// POST /user/checkout
func (h *Handler) Checkout(c *fiber.Ctx) error {
	log.Printf("Checkout: Request received from IP %s", c.IP())

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("Checkout: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("Checkout: User authenticated - UserID: %s", userID)

	// Parse request body
	var req models.CheckoutRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Checkout: Failed to parse request body: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		log.Printf("Checkout: Validation failed: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Validation failed",
		})
	}

	// Get cart items to verify they belong to the user
	cartItems, err := h.repo.GetCartItems(context.Background(), userID)
	if err != nil {
		log.Printf("Checkout: Failed to get cart items: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get cart items",
		})
	}

	// Create a map of cart item IDs for quick lookup
	cartItemMap := make(map[uuid.UUID]models.CartItemWithDetails)
	for _, item := range cartItems {
		cartItemMap[item.ID] = item
	}

	// Verify all requested cart items belong to the user
	var validItems []models.CartItemWithDetails
	var totalAmount float64
	for _, cartItemID := range req.CartItemIDs {
		if item, exists := cartItemMap[cartItemID]; exists {
			validItems = append(validItems, item)
			totalAmount += item.AudioBook.Price
		} else {
			log.Printf("Checkout: Cart item %s not found or doesn't belong to user", cartItemID)
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid cart item",
			})
		}
	}

	if len(validItems) == 0 {
		log.Printf("Checkout: No valid items to checkout")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No items to checkout",
		})
	}

	// Generate transaction ID (in a real app, this would come from payment processor)
	transactionID := fmt.Sprintf("txn_%s_%d", userID.String()[:8], time.Now().Unix())

	// Create purchase records for each item
	var purchasedItems []models.PurchasedAudioBookWithDetails
	for _, item := range validItems {
		// Check if user already purchased this audiobook
		alreadyPurchased, err := h.repo.IsAudioBookPurchased(context.Background(), userID, item.AudiobookID)
		if err != nil {
			log.Printf("Checkout: Failed to check if audiobook is purchased: %v", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process checkout",
			})
		}

		if alreadyPurchased {
			log.Printf("Checkout: User already purchased audiobook %s", item.AudiobookID)
			continue // Skip already purchased items
		}

		// Create purchase record
		purchase := &models.PurchasedAudioBook{
			ID:            uuid.New(),
			UserID:        userID,
			AudiobookID:   item.AudiobookID,
			PurchasePrice: item.AudioBook.Price,
			TransactionID: &transactionID,
			PaymentStatus: "completed", // Mock payment status
		}

		if err := h.repo.CreatePurchasedAudioBook(context.Background(), purchase); err != nil {
			log.Printf("Checkout: Failed to create purchase record: %v", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to process checkout",
			})
		}

		// Add to purchased items list
		purchasedItems = append(purchasedItems, models.PurchasedAudioBookWithDetails{
			PurchasedAudioBook: *purchase,
			AudioBook:          item.AudioBook,
		})

		// Remove from cart
		if err := h.repo.RemoveFromCart(context.Background(), userID, item.AudiobookID); err != nil {
			log.Printf("Checkout: Failed to remove item from cart: %v", err)
			// Don't fail the checkout if cart removal fails
		}
	}

	// Generate order ID
	orderID := fmt.Sprintf("order_%s_%d", userID.String()[:8], time.Now().Unix())

	response := models.CheckoutResponse{
		OrderID:             orderID,
		PurchasedItems:      purchasedItems,
		TotalAmount:         totalAmount,
		TransactionID:       transactionID,
		CheckoutCompletedAt: time.Now(),
	}

	log.Printf("Checkout: Successfully processed checkout for user %s with %d items", userID, len(purchasedItems))

	return c.JSON(fiber.Map{
		"message": "Checkout completed successfully",
		"data":    response,
	})
}

// GetPurchaseHistory returns the user's purchase history
// GET /user/purchases
func (h *Handler) GetPurchaseHistory(c *fiber.Ctx) error {
	log.Printf("GetPurchaseHistory: Request received from IP %s", c.IP())

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("GetPurchaseHistory: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("GetPurchaseHistory: User authenticated - UserID: %s", userID)

	// Parse query parameters
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Cap at 100 items per page
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get purchase history
	purchaseHistory, err := h.repo.GetPurchaseHistory(context.Background(), userID, limit, offset)
	if err != nil {
		log.Printf("GetPurchaseHistory: Failed to get purchase history: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get purchase history",
		})
	}

	log.Printf("GetPurchaseHistory: Successfully retrieved %d purchases for user %s", len(purchaseHistory.Purchases), userID)

	return c.JSON(fiber.Map{
		"data": purchaseHistory,
	})
}

// IsAudioBookPurchased checks if a user has purchased a specific audiobook
// GET /user/audiobooks/:audiobookId/purchased
func (h *Handler) IsAudioBookPurchased(c *fiber.Ctx) error {
	log.Printf("IsAudioBookPurchased: Request received from IP %s", c.IP())

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("IsAudioBookPurchased: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("IsAudioBookPurchased: User authenticated - UserID: %s", userID)

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("audiobookId"))
	if err != nil {
		log.Printf("IsAudioBookPurchased: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	// Check if purchased
	isPurchased, err := h.repo.IsAudioBookPurchased(context.Background(), userID, audiobookID)
	if err != nil {
		log.Printf("IsAudioBookPurchased: Failed to check if purchased: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check purchase status",
		})
	}

	log.Printf("IsAudioBookPurchased: Audiobook %s is purchased by user %s: %v", audiobookID, userID, isPurchased)

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"is_purchased": isPurchased,
		},
	})
}

// DeleteAudioBook deletes an audiobook
// DELETE /audiobooks/:id
func (h *Handler) DeleteAudioBook(c *fiber.Ctx) error {
	log.Printf("DeleteAudioBook: Request received from IP %s", c.IP())
	log.Printf("DeleteAudioBook: User agent: %s", c.Get("User-Agent"))

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("DeleteAudioBook: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("DeleteAudioBook: Deleting audiobook with ID: %s", audiobookID)

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("DeleteAudioBook: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	userID := uuid.MustParse(userCtx.ID)
	log.Printf("DeleteAudioBook: User authenticated - UserID: %s", userID)

	// Get existing audiobook
	existingAudiobook, err := h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("DeleteAudioBook: Failed to fetch audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Check if user owns this audiobook
	if existingAudiobook.CreatedBy != userID {
		log.Printf("DeleteAudioBook: Access denied - user does not own audiobook")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	// Delete from database
	if err := h.repo.DeleteAudioBook(context.Background(), audiobookID); err != nil {
		log.Printf("DeleteAudioBook: Failed to delete audiobook: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete audiobook",
		})
	}

	log.Printf("DeleteAudioBook: Successfully deleted audiobook: %s", existingAudiobook.Title)

	return c.JSON(fiber.Map{
		"message": "Audiobook deleted successfully",
	})
}

// GetPublicAudioBooks returns a list of public audiobooks
// GET /audiobooks (public route)
func (h *Handler) GetPublicAudioBooks(c *fiber.Ctx) error {
	log.Printf("GetPublicAudioBooks: Request received from IP %s", c.IP())
	log.Printf("GetPublicAudioBooks: User agent: %s", c.Get("User-Agent"))

	// Parse query parameters
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Cap at 100 items per page
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	log.Printf("GetPublicAudioBooks: Fetching public audiobooks with limit=%d, offset=%d", limit, offset)

	// Set isPublic to true to only get public audiobooks
	isPublic := true

	// Get public audiobooks from database
	audiobooks, total, err := h.repo.ListAudioBooks(context.Background(), limit, offset, &isPublic)
	if err != nil {
		log.Printf("GetPublicAudioBooks: Failed to fetch audiobooks: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch audiobooks",
		})
	}

	log.Printf("GetPublicAudioBooks: Found %d public audiobooks (total: %d)", len(audiobooks), total)

	// Calculate pagination info
	totalPages := (total + limit - 1) / limit
	currentPage := (offset / limit) + 1

	return c.JSON(fiber.Map{
		"data": audiobooks,
		"pagination": fiber.Map{
			"total":        total,
			"limit":        limit,
			"offset":       offset,
			"current_page": currentPage,
			"total_pages":  totalPages,
		},
	})
}

// GetPublicAudioBook returns a specific public audiobook by ID
// GET /audiobooks/:id (public route)
func (h *Handler) GetPublicAudioBook(c *fiber.Ctx) error {
	log.Printf("GetPublicAudioBook: Request received from IP %s", c.IP())
	log.Printf("GetPublicAudioBook: User agent: %s", c.Get("User-Agent"))

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("GetPublicAudioBook: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("GetPublicAudioBook: Fetching public audiobook with ID: %s", audiobookID)

	// Get audiobook with details from database
	audiobook, err := h.repo.GetAudioBookWithDetails(context.Background(), audiobookID)
	if err != nil {
		log.Printf("GetPublicAudioBook: Failed to fetch audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Check if audiobook is public
	if !audiobook.IsPublic {
		log.Printf("GetPublicAudioBook: Access denied - audiobook is not public")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	log.Printf("GetPublicAudioBook: Successfully retrieved public audiobook: %s", audiobook.Title)

	return c.JSON(fiber.Map{
		"data": audiobook,
	})
}

// CreateAudioBook creates an audio book from a completed upload session
// POST /v1/admin/audiobooks
func (h *Handler) CreateAudioBook(c *fiber.Ctx) error {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("CreateAudioBook: PANIC recovered: %v", r)
		}
	}()

	log.Printf("CreateAudioBook: Request received from IP %s", c.IP())
	log.Printf("CreateAudioBook: User agent: %s", c.Get("User-Agent"))

	// Log raw request body for debugging
	body := c.Body()
	log.Printf("CreateAudioBook: Raw request body: %s", string(body))

	var req models.CreateAudioBookFromUploadRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("CreateAudioBook: Failed to parse request body: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	log.Printf("CreateAudioBook: Request parsed successfully - UploadID: %s, Title: %s, Author: %s, Language: %s, IsPublic: %v",
		req.UploadID, req.Title, req.Author, req.Language, req.IsPublic)

	if err := req.Validate(); err != nil {
		log.Printf("CreateAudioBook: Validation failed: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Validation error: %v", err),
		})
	}

	log.Printf("CreateAudioBook: Request validation passed")

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	log.Printf("CreateAudioBook: User context: %v", userCtx)
	if !ok || userCtx == nil {
		log.Printf("CreateAudioBook: User not authenticated - user context not found in context")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	log.Printf("CreateAudioBook: User authenticated - UserID: %s", userCtx.ID)
	userID := uuid.MustParse(userCtx.ID)

	// Get upload session
	log.Printf("CreateAudioBook: Looking up upload session with ID: %s", req.UploadID)
	upload, err := h.repo.GetUploadByID(context.Background(), req.UploadID)
	if err != nil {
		log.Printf("CreateAudioBook: Failed to get upload session: %v", err)
		// Log more details about the error
		if strings.Contains(err.Error(), "connection") {
			log.Printf("CreateAudioBook: Database connection error detected")
		}
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Upload session not found",
		})
	}

	log.Printf("CreateAudioBook: Upload session found - Status: %s, UserID: %s, UploadType: %s",
		upload.Status, upload.UserID, upload.UploadType)

	// Check if user owns this upload
	log.Printf("CreateAudioBook: Checking ownership - Upload UserID: %s, Request UserID: %s", upload.UserID, userID)
	if upload.UserID != userID {
		log.Printf("CreateAudioBook: Access denied - user does not own this upload")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
		})
	}

	log.Printf("CreateAudioBook: Ownership verified")

	// Check if upload is completed
	log.Printf("CreateAudioBook: Checking upload status - Current: %s, Expected: %s", upload.Status, models.UploadStatusCompleted)
	if upload.Status != models.UploadStatusCompleted {
		log.Printf("CreateAudioBook: Upload session is not completed")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Upload session is not completed",
		})
	}

	log.Printf("CreateAudioBook: Upload status verified as completed")

	// Get upload files
	log.Printf("CreateAudioBook: Getting upload files for upload ID: %s", req.UploadID)
	uploadFiles, err := h.repo.GetUploadFiles(context.Background(), req.UploadID)
	if err != nil {
		log.Printf("CreateAudioBook: Failed to get upload files: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get upload files",
		})
	}

	log.Printf("CreateAudioBook: Found %d upload files", len(uploadFiles))

	if len(uploadFiles) == 0 {
		log.Printf("CreateAudioBook: No files found in upload session")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No files found in upload session",
		})
	}

	// Create audio book
	log.Printf("CreateAudioBook: Creating audio book with title: %s, author: %s", req.Title, req.Author)
	audiobook := &models.AudioBook{
		ID:        uuid.New(),
		Title:     req.Title,
		Author:    req.Author,
		Language:  req.Language,
		IsPublic:  req.IsPublic,
		Price:     0.0,                     // Default price, can be updated later
		Status:    models.StatusProcessing, // Start with processing status
		CreatedBy: userID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set price if provided
	if req.Price != nil {
		audiobook.Price = *req.Price
	}

	// Set description if provided
	if req.Description != nil && *req.Description != "" {
		audiobook.Summary = req.Description
	}

	// Set cover image URL if provided
	if req.CoverImageURL != nil && *req.CoverImageURL != "" {
		audiobook.CoverImageURL = req.CoverImageURL
	}

	// Calculate total duration from all upload files
	var totalDurationSeconds int
	for _, file := range uploadFiles {
		if file.DurationSeconds != nil {
			totalDurationSeconds += *file.DurationSeconds
		}
	}

	// Set the total duration on the audiobook
	if totalDurationSeconds > 0 {
		audiobook.DurationSeconds = &totalDurationSeconds
		log.Printf("CreateAudioBook: Total duration calculated: %d seconds", totalDurationSeconds)
	}

	// Handle different upload types
	log.Printf("CreateAudioBook: Processing upload type: %s", upload.UploadType)
	var chapters []*models.Chapter

	if upload.UploadType == models.UploadTypeSingle {
		// Single file upload - create one chapter
		log.Printf("CreateAudioBook: Processing single file upload")
		if len(uploadFiles) != 1 {
			log.Printf("CreateAudioBook: Single upload type requires exactly one file, found %d", len(uploadFiles))
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Single upload type requires exactly one file",
			})
		}

		file := uploadFiles[0]
		log.Printf("CreateAudioBook: Creating single chapter with file: %s, size: %d", file.FilePath, file.FileSize)

		chapter := &models.Chapter{
			ID:            uuid.New(),
			AudiobookID:   audiobook.ID,
			UploadFileID:  &file.ID,
			ChapterNumber: 1,
			Title:         req.Title, // Use audiobook title for single file
			FilePath:      file.FilePath,
			FileURL:       &file.FilePath, // For now, use file path as URL
			FileSizeBytes: &file.FileSize,
			MimeType:      &file.MimeType,
			CreatedAt:     time.Now(),
		}
		chapters = append(chapters, chapter)

	} else if upload.UploadType == models.UploadTypeChapters {
		// Chaptered upload
		log.Printf("CreateAudioBook: Processing chaptered upload")
		if len(uploadFiles) == 0 {
			log.Printf("CreateAudioBook: No chapter files found")
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "No chapter files found",
			})
		}

		// Sort files by chapter number
		chapterFiles := make(map[int]*models.UploadFile)
		for _, file := range uploadFiles {
			log.Printf("CreateAudioBook: Processing file: %s, chapter number: %v", file.FileName, file.ChapterNumber)
			if file.ChapterNumber == nil {
				log.Printf("CreateAudioBook: File %s missing chapter number", file.FileName)
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "All files must have chapter numbers for chaptered uploads",
				})
			}
			chapterFiles[*file.ChapterNumber] = &file
		}

		log.Printf("CreateAudioBook: Processed %d chapter files", len(chapterFiles))

		// Check if chapter 1 exists
		if chapterFiles[1] == nil {
			log.Printf("CreateAudioBook: Chapter 1 is required but not found")
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Chapter 1 is required",
			})
		}

		// Create chapter records
		for chapterNum, file := range chapterFiles {
			chapterTitle := file.FileName
			if file.ChapterTitle != nil && *file.ChapterTitle != "" {
				chapterTitle = *file.ChapterTitle
			}

			chapter := &models.Chapter{
				ID:            uuid.New(),
				AudiobookID:   audiobook.ID,
				UploadFileID:  &file.ID,
				ChapterNumber: chapterNum,
				Title:         chapterTitle,
				FilePath:      file.FilePath,
				FileURL:       &file.FilePath, // For now, use file path as URL
				FileSizeBytes: &file.FileSize,
				MimeType:      &file.MimeType,
				CreatedAt:     time.Now(),
			}
			chapters = append(chapters, chapter)
		}
	}

	// Save audio book to database first (required for foreign key constraint)
	log.Printf("CreateAudioBook: Saving audio book to database with ID: %s", audiobook.ID)
	if err := h.repo.CreateAudioBook(context.Background(), audiobook); err != nil {
		log.Printf("CreateAudioBook: Failed to create audio book in database: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create audio book",
		})
	}
	log.Printf("CreateAudioBook: Audio book saved to database successfully")

	// Save chapters to database
	log.Printf("CreateAudioBook: Creating %d chapters in database", len(chapters))
	for _, chapter := range chapters {
		if err := h.repo.CreateChapter(context.Background(), chapter); err != nil {
			log.Printf("CreateAudioBook: Failed to create chapter %d: %v", chapter.ChapterNumber, err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create chapter",
			})
		}
	}
	log.Printf("CreateAudioBook: All chapters created successfully")

	// Create transcription jobs for each chapter and enqueue them to Redis
	log.Printf("CreateAudioBook: Creating transcription jobs for %d chapters", len(chapters))
	jobsCreated := 0

	for _, chapter := range chapters {
		log.Printf("CreateAudioBook: Creating transcription job for chapter %d", chapter.ChapterNumber)
		job := &models.ProcessingJob{
			ID:          uuid.New(),
			AudiobookID: audiobook.ID,
			JobType:     models.JobTypeTranscribe,
			Status:      models.JobStatusPending,
			ChapterID:   &chapter.ID,
			RetryCount:  0,
			MaxRetries:  3,
			CreatedAt:   time.Now(),
		}

		// Save job to database
		if err := h.repo.CreateProcessingJob(context.Background(), job); err != nil {
			log.Printf("CreateAudioBook: Failed to create transcription job for chapter %d: %v", chapter.ChapterNumber, err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create transcription job",
			})
		}
		log.Printf("CreateAudioBook: Transcription job for chapter %d created in database", chapter.ChapterNumber)

		// Enqueue job to Redis if Redis service is available
		if h.redisQueue != nil {
			log.Printf("CreateAudioBook: Enqueueing transcription job for chapter %d with file path: %s", chapter.ChapterNumber, chapter.FilePath)
			if err := h.redisQueue.EnqueueTranscriptionJob(context.Background(), job, chapter.FilePath); err != nil {
				log.Printf("CreateAudioBook: Failed to enqueue transcription job for chapter %d to Redis: %v", chapter.ChapterNumber, err)
				// Continue with other jobs even if one fails to enqueue
			} else {
				log.Printf("CreateAudioBook: Transcription job for chapter %d enqueued to Redis successfully", chapter.ChapterNumber)
				jobsCreated++
			}
		} else {
			log.Printf("CreateAudioBook: Redis queue service not available, transcription job for chapter %d created in database only", chapter.ChapterNumber)
			jobsCreated++
		}
	}

	// Create summarize job in idle state - it will be activated when TriggerSummarizeAndTagJobs is called
	log.Printf("CreateAudioBook: Creating summarize job for audiobook in idle state")
	summarizeJob := &models.ProcessingJob{
		ID:          uuid.New(),
		AudiobookID: audiobook.ID,
		JobType:     models.JobTypeSummarize,
		Status:      models.JobStatusIdle,
		RetryCount:  0,
		MaxRetries:  3,
		CreatedAt:   time.Now(),
	}

	// Save summarize job to database
	if err := h.repo.CreateProcessingJob(context.Background(), summarizeJob); err != nil {
		log.Printf("CreateAudioBook: Failed to create summarize job: %v", err)
		// Don't fail the request if summarize job creation fails
	} else {
		log.Printf("CreateAudioBook: Summarize job created in database (ID: %s) in idle state - will be activated when transcription is complete", summarizeJob.ID)
	}

	// Update upload status to indicate it's been processed
	log.Printf("CreateAudioBook: Updating upload status to processed")
	upload.Status = "processed"
	upload.UpdatedAt = time.Now()
	if err := h.repo.UpdateUpload(context.Background(), upload); err != nil {
		// Log error but don't fail the request
		log.Printf("CreateAudioBook: Failed to update upload status: %v", err)
	} else {
		log.Printf("CreateAudioBook: Upload status updated successfully")
	}

	log.Printf("CreateAudioBook: Request completed successfully - AudioBookID: %s, TranscriptionJobsCreated: %d/%d",
		audiobook.ID, jobsCreated, len(chapters))
	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"audiobook_id":               audiobook.ID,
		"status":                     audiobook.Status,
		"message":                    "Audio book created successfully",
		"transcription_jobs_created": jobsCreated,
		"total_chapters":             len(chapters),
	})
}

// TriggerSummarizeAndTagJobs triggers summarize and tag jobs for an audiobook after transcription is complete
// POST /v1/internal/audiobooks/{id}/trigger-summarize-tag (internal service-to-service)
// POST /v1/admin/audiobooks/{id}/trigger-summarize-tag (admin authenticated)
func (h *Handler) TriggerSummarizeAndTagJobs(c *fiber.Ctx) error {
	log.Printf("TriggerSummarizeAndTagJobs: Request received from IP %s", c.IP())

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("TriggerSummarizeAndTagJobs: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("TriggerSummarizeAndTagJobs: Processing audiobook ID: %s", audiobookID)

	// Get audiobook to check if it exists
	_, err = h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("TriggerSummarizeAndTagJobs: Failed to get audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Get all chapters for this audiobook
	chapters, err := h.repo.GetChaptersByAudioBookID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("TriggerSummarizeAndTagJobs: Failed to get chapters: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get chapters",
		})
	}

	if len(chapters) == 0 {
		log.Printf("TriggerSummarizeAndTagJobs: No chapters found for audiobook")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No chapters found for audiobook",
		})
	}

	// Check if chapter 1 has been transcribed successfully
	chapter1Transcribed := false

	for _, chapter := range chapters {
		if chapter.ChapterNumber == 1 {
			// Check if chapter 1 has a transcript
			transcript, err := h.repo.GetChapterTranscriptByChapterID(context.Background(), chapter.ID)
			if err == nil && transcript != nil {
				chapter1Transcribed = true
				log.Printf("TriggerSummarizeAndTagJobs: Chapter 1 has transcript")
			} else {
				log.Printf("TriggerSummarizeAndTagJobs: Chapter 1 has no transcript")
			}
			break
		}
	}

	if !chapter1Transcribed {
		log.Printf("TriggerSummarizeAndTagJobs: Chapter 1 is not transcribed")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Chapter 1 is not transcribed",
		})
	}

	log.Printf("TriggerSummarizeAndTagJobs: Chapter 1 is transcribed, proceeding with summarize and tag jobs")

	// Find existing summarize job for this audiobook
	existingJobs, err := h.repo.GetProcessingJobsByAudioBookID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("TriggerSummarizeAndTagJobs: Failed to get existing processing jobs: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get existing processing jobs",
		})
	}

	var summarizeJob *models.ProcessingJob
	for _, job := range existingJobs {
		if job.JobType == models.JobTypeSummarize {
			summarizeJob = &job
			break
		}
	}

	if summarizeJob == nil {
		log.Printf("TriggerSummarizeAndTagJobs: No existing summarize job found for audiobook %s", audiobookID)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "No summarize job found for this audiobook",
		})
	}

	log.Printf("TriggerSummarizeAndTagJobs: Found existing summarize job (ID: %s) with status: %s", summarizeJob.ID, summarizeJob.Status)

	// Check if job is already in progress, completed, or already pending
	if summarizeJob.Status == models.JobStatusRunning || summarizeJob.Status == models.JobStatusCompleted || summarizeJob.Status == models.JobStatusPending {
		log.Printf("TriggerSummarizeAndTagJobs: Summarize job is already %s", summarizeJob.Status)
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error":  fmt.Sprintf("Summarize job is already %s", summarizeJob.Status),
			"job_id": summarizeJob.ID,
		})
	}

	// Check if job is in idle state (expected state)
	if summarizeJob.Status != models.JobStatusIdle {
		log.Printf("TriggerSummarizeAndTagJobs: Unexpected job status: %s (expected: idle)", summarizeJob.Status)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":  fmt.Sprintf("Unexpected job status: %s (expected: idle)", summarizeJob.Status),
			"job_id": summarizeJob.ID,
		})
	}

	// Change job status from idle to pending
	summarizeJob.Status = models.JobStatusPending
	if err := h.repo.UpdateProcessingJob(context.Background(), summarizeJob); err != nil {
		log.Printf("TriggerSummarizeAndTagJobs: Failed to update job status to pending: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update job status",
		})
	}
	log.Printf("TriggerSummarizeAndTagJobs: Job status updated from idle to pending")

	// Enqueue job to Redis if Redis service is available
	if h.redisQueue != nil {
		if err := h.redisQueue.EnqueueAIJob(context.Background(), summarizeJob); err != nil {
			log.Printf("TriggerSummarizeAndTagJobs: Failed to enqueue summarize and tag job to Redis: %v", err)
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to enqueue summarize and tag job",
			})
		}
		log.Printf("TriggerSummarizeAndTagJobs: Summarize and tag job enqueued to Redis successfully")
	} else {
		log.Printf("TriggerSummarizeAndTagJobs: Redis queue service not available, job status updated to pending but not enqueued")
	}

	log.Printf("TriggerSummarizeAndTagJobs: Successfully triggered summarize and tag jobs for audiobook %s", audiobookID)

	return c.JSON(fiber.Map{
		"audiobook_id":      audiobookID,
		"message":           "Summarize and tag jobs triggered successfully",
		"job_id":            summarizeJob.ID,
		"chapters_verified": len(chapters),
	})
}

// GetJobStatus returns the status of processing jobs for an audiobook
// GET /v1/admin/audiobooks/{id}/jobs
func (h *Handler) GetJobStatus(c *fiber.Ctx) error {
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}
	userID := uuid.MustParse(userCtx.ID)

	// Get audiobook to check ownership
	audiobook, err := h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audio book not found",
		})
	}

	// Check if user owns this audiobook or is admin
	if audiobook.CreatedBy != userID {
		// Check if user is admin
		if userCtx.Role != "admin" {
			return c.Status(http.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied",
			})
		}
	}

	// Get processing jobs
	jobs, err := h.repo.GetProcessingJobsByAudioBookID(context.Background(), audiobookID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get processing jobs",
		})
	}

	// Calculate overall progress
	totalJobs := len(jobs)
	completedJobs := 0
	failedJobs := 0

	for _, job := range jobs {
		if job.Status == models.JobStatusCompleted {
			completedJobs++
		} else if job.Status == models.JobStatusFailed {
			failedJobs++
		}
	}

	var progress float64
	if totalJobs > 0 {
		progress = float64(completedJobs) / float64(totalJobs)
	}

	// Determine overall status
	var overallStatus models.AudioBookStatus
	if failedJobs > 0 {
		overallStatus = models.StatusFailed
	} else if completedJobs == totalJobs {
		overallStatus = models.StatusCompleted
	} else {
		// Check what type of jobs are currently running to provide more specific status
		hasTranscribingJobs := false
		hasSummarizingJobs := false

		for _, job := range jobs {
			if job.Status == models.JobStatusRunning || job.Status == models.JobStatusPending {
				if job.JobType == models.JobTypeTranscribe {
					hasTranscribingJobs = true
				} else if job.JobType == models.JobTypeSummarize {
					hasSummarizingJobs = true
				}
			}
		}

		// Prioritize transcribing over summarizing for status display
		if hasTranscribingJobs {
			overallStatus = models.StatusTranscribing
		} else if hasSummarizingJobs {
			overallStatus = models.StatusSummarizing
		} else {
			overallStatus = models.StatusProcessing
		}
	}

	// Update audiobook status if needed
	if overallStatus != audiobook.Status {
		h.repo.UpdateAudioBookStatus(context.Background(), audiobookID, overallStatus)
	}

	return c.JSON(fiber.Map{
		"data": fiber.Map{
			"audiobook_id":   audiobookID,
			"jobs":           jobs,
			"overall_status": overallStatus,
			"progress":       progress,
			"total_jobs":     totalJobs,
			"completed_jobs": completedJobs,
			"failed_jobs":    failedJobs,
		},
	})
}

// UpdateJobStatus updates the status of a processing job (called by workers)
// POST /v1/admin/jobs/{job_id}/status
func (h *Handler) UpdateJobStatus(c *fiber.Ctx) error {
	jobID, err := uuid.Parse(c.Params("job_id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid job ID",
		})
	}

	var req struct {
		Status       models.JobStatus `json:"status" validate:"required"`
		ErrorMessage *string          `json:"error_message,omitempty"`
		StartedAt    *time.Time       `json:"started_at,omitempty"`
		CompletedAt  *time.Time       `json:"completed_at,omitempty"`
		RetryCount   *int             `json:"retry_count,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get the job
	job, err := h.repo.GetProcessingJobByID(context.Background(), jobID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	// If status is failed and retry count is provided, increment it
	if req.Status == models.JobStatusFailed && req.RetryCount != nil {
		// Increment retry count in database
		if err := h.repo.IncrementRetryCount(context.Background(), jobID); err != nil {
			fmt.Printf("Failed to increment retry count for job %s: %v\n", jobID, err)
			// Continue with the update even if retry count increment fails
		} else {
			// Refetch the job to get the updated retry count
			updatedJob, err := h.repo.GetProcessingJobByID(context.Background(), jobID)
			if err == nil {
				job = updatedJob
			}
		}
	}

	// Update job status
	job.Status = req.Status
	job.ErrorMessage = req.ErrorMessage
	job.StartedAt = req.StartedAt
	job.CompletedAt = req.CompletedAt

	fmt.Println("job retry count after increment", job.RetryCount)

	if err := h.repo.UpdateProcessingJob(context.Background(), job); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update job status",
		})
	}

	// Check and update audiobook status
	if err := h.repo.CheckAndUpdateAudioBookStatus(context.Background(), job.AudiobookID); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to update audiobook status: %v\n", err)
	}

	return c.JSON(fiber.Map{
		"job_id":  jobID,
		"status":  req.Status,
		"message": "Job status updated successfully",
	})
}

// UpdateAudioBookPrice updates the price of an audiobook (admin only)
// PUT /admin/audiobooks/:id/price
func (h *Handler) UpdateAudioBookPrice(c *fiber.Ctx) error {
	log.Printf("UpdateAudioBookPrice: Request received from IP %s", c.IP())
	log.Printf("UpdateAudioBookPrice: User agent: %s", c.Get("User-Agent"))

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("UpdateAudioBookPrice: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("UpdateAudioBookPrice: Updating price for audiobook with ID: %s", audiobookID)

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("UpdateAudioBookPrice: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Check if user is admin
	if userCtx.Role != models.RoleAdmin {
		log.Printf("UpdateAudioBookPrice: Access denied - user is not admin")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Admin access required",
		})
	}

	// Parse request body
	var req struct {
		Price float64 `json:"price" validate:"required,min=0"`
	}

	if err := c.BodyParser(&req); err != nil {
		log.Printf("UpdateAudioBookPrice: Failed to parse request body: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate price
	if req.Price < 0 {
		log.Printf("UpdateAudioBookPrice: Invalid price: %f", req.Price)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Price must be non-negative",
		})
	}

	// Get existing audiobook
	existingAudiobook, err := h.repo.GetAudioBookByID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("UpdateAudioBookPrice: Failed to fetch audiobook: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Audiobook not found",
		})
	}

	// Update price
	existingAudiobook.Price = req.Price
	existingAudiobook.UpdatedAt = time.Now()

	// Update in database
	if err := h.repo.UpdateAudioBook(context.Background(), existingAudiobook); err != nil {
		log.Printf("UpdateAudioBookPrice: Failed to update audiobook price: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update audiobook price",
		})
	}

	log.Printf("UpdateAudioBookPrice: Successfully updated price for audiobook: %s to $%.2f", existingAudiobook.Title, req.Price)

	return c.JSON(fiber.Map{
		"data":    existingAudiobook,
		"message": "Audiobook price updated successfully",
	})
}

// RetryJob retries a failed processing job
// POST /admin/audiobooks/:id/jobs/:job_id/retry
func (h *Handler) RetryJob(c *fiber.Ctx) error {
	log.Printf("RetryJob: Request received from IP %s", c.IP())

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("RetryJob: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	// Parse job ID
	jobID, err := uuid.Parse(c.Params("job_id"))
	if err != nil {
		log.Printf("RetryJob: Invalid job ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid job ID",
		})
	}

	log.Printf("RetryJob: Processing retry for audiobook %s, job %s", audiobookID, jobID)

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("RetryJob: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Check if user is admin
	if userCtx.Role != models.RoleAdmin {
		log.Printf("RetryJob: Access denied - user is not admin")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Admin access required",
		})
	}

	// Get the job
	job, err := h.repo.GetProcessingJobByID(context.Background(), jobID)
	if err != nil {
		log.Printf("RetryJob: Failed to get job: %v", err)
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Job not found",
		})
	}

	// Verify job belongs to the audiobook
	if job.AudiobookID != audiobookID {
		log.Printf("RetryJob: Job does not belong to audiobook")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Job does not belong to audiobook",
		})
	}

	// Check if job is in failed state
	if job.Status != models.JobStatusFailed {
		log.Printf("RetryJob: Job is not in failed state, current status: %s", job.Status)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Job is not in failed state, current status: %s", job.Status),
		})
	}

	// Check if job has reached max retries
	if job.RetryCount >= job.MaxRetries {
		log.Printf("RetryJob: Job has reached max retries (%d/%d)", job.RetryCount, job.MaxRetries)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Job has reached maximum retry attempts",
		})
	}

	// Reset job status to pending and clear error
	job.Status = models.JobStatusPending
	job.ErrorMessage = nil
	job.StartedAt = nil
	job.CompletedAt = nil

	// Update job in database
	if err := h.repo.UpdateProcessingJob(context.Background(), job); err != nil {
		log.Printf("RetryJob: Failed to update job status: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update job status",
		})
	}

	// Re-enqueue job to Redis if Redis service is available
	if h.redisQueue != nil {
		if job.JobType == models.JobTypeTranscribe {
			// For transcription jobs, we need the file path from the chapter
			if job.ChapterID != nil {
				chapter, err := h.repo.GetChapterByID(context.Background(), *job.ChapterID)
				if err != nil {
					log.Printf("RetryJob: Failed to get chapter for transcription job: %v", err)
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to get chapter information",
					})
				}

				if err := h.redisQueue.EnqueueTranscriptionJob(context.Background(), job, chapter.FilePath); err != nil {
					log.Printf("RetryJob: Failed to enqueue transcription job to Redis: %v", err)
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to enqueue job for retry",
					})
				}
			} else {
				log.Printf("RetryJob: Transcription job missing chapter ID")
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "Transcription job missing chapter information",
				})
			}
		} else {
			// For other job types (summarize, tag, etc.)
			if err := h.redisQueue.EnqueueAIJob(context.Background(), job); err != nil {
				log.Printf("RetryJob: Failed to enqueue AI job to Redis: %v", err)
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to enqueue job for retry",
				})
			}
		}

		log.Printf("RetryJob: Job successfully re-enqueued to Redis")
	} else {
		log.Printf("RetryJob: Redis queue service not available, job status updated to pending only")
	}

	log.Printf("RetryJob: Successfully initiated retry for job %s", jobID)

	return c.JSON(fiber.Map{
		"message":      "Job retry initiated successfully",
		"job_id":       jobID,
		"audiobook_id": audiobookID,
		"status":       job.Status,
		"retry_count":  job.RetryCount,
	})
}

// RetryAllFailedJobs retries all failed processing jobs for an audiobook
// POST /admin/audiobooks/:id/retry-all
func (h *Handler) RetryAllFailedJobs(c *fiber.Ctx) error {
	log.Printf("RetryAllFailedJobs: Request received from IP %s", c.IP())

	// Parse audiobook ID
	audiobookID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		log.Printf("RetryAllFailedJobs: Invalid audiobook ID: %v", err)
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid audiobook ID",
		})
	}

	log.Printf("RetryAllFailedJobs: Processing retry for all failed jobs in audiobook %s", audiobookID)

	// Get user ID from context
	userCtx, ok := c.Locals("user").(*models.UserContext)
	if !ok || userCtx == nil {
		log.Printf("RetryAllFailedJobs: User not authenticated")
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Check if user is admin
	if userCtx.Role != models.RoleAdmin {
		log.Printf("RetryAllFailedJobs: Access denied - user is not admin")
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Admin access required",
		})
	}

	// Get all jobs for the audiobook
	allJobs, err := h.repo.GetProcessingJobsByAudioBookID(context.Background(), audiobookID)
	if err != nil {
		log.Printf("RetryAllFailedJobs: Failed to get jobs: %v", err)
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get processing jobs",
		})
	}

	// Filter failed jobs that can be retried
	var retriableJobs []models.ProcessingJob
	for _, job := range allJobs {
		if job.Status == models.JobStatusFailed && job.RetryCount >= job.MaxRetries {
			retriableJobs = append(retriableJobs, job)
		}
	}

	if len(retriableJobs) == 0 {
		log.Printf("RetryAllFailedJobs: No retriable failed jobs found")
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No failed jobs available for retry",
		})
	}

	log.Printf("RetryAllFailedJobs: Found %d retriable failed jobs", len(retriableJobs))

	var retriedJobs []string
	var failedRetries []string

	// Process each retriable job
	for _, job := range retriableJobs {
		log.Printf("RetryAllFailedJobs: Processing job %s (type: %s)", job.ID, job.JobType)

		// Reset job status to pending and clear error
		job.Status = models.JobStatusPending
		job.ErrorMessage = nil
		job.StartedAt = nil
		job.CompletedAt = nil
		job.RetryCount = 0

		// Update job in database
		if err := h.repo.UpdateProcessingJob(context.Background(), &job); err != nil {
			log.Printf("RetryAllFailedJobs: Failed to update job %s status: %v", job.ID, err)
			failedRetries = append(failedRetries, job.ID.String())
			continue
		}

		// Re-enqueue job to Redis if Redis service is available
		if h.redisQueue != nil {
			var enqueueErr error

			if job.JobType == models.JobTypeTranscribe {
				// For transcription jobs, we need the file path from the chapter
				if job.ChapterID != nil {
					chapter, err := h.repo.GetChapterByID(context.Background(), *job.ChapterID)
					if err != nil {
						log.Printf("RetryAllFailedJobs: Failed to get chapter for job %s: %v", job.ID, err)
						failedRetries = append(failedRetries, job.ID.String())
						continue
					}

					enqueueErr = h.redisQueue.EnqueueTranscriptionJob(context.Background(), &job, chapter.FilePath)
				} else {
					log.Printf("RetryAllFailedJobs: Transcription job %s missing chapter ID", job.ID)
					failedRetries = append(failedRetries, job.ID.String())
					continue
				}
			} else {
				// For other job types (summarize, tag, etc.)
				enqueueErr = h.redisQueue.EnqueueAIJob(context.Background(), &job)
			}

			if enqueueErr != nil {
				log.Printf("RetryAllFailedJobs: Failed to enqueue job %s to Redis: %v", job.ID, enqueueErr)
				failedRetries = append(failedRetries, job.ID.String())
				continue
			}
		}

		retriedJobs = append(retriedJobs, job.ID.String())
		log.Printf("RetryAllFailedJobs: Successfully retried job %s", job.ID)
	}

	log.Printf("RetryAllFailedJobs: Completed - %d jobs retried, %d failed to retry",
		len(retriedJobs), len(failedRetries))

	response := fiber.Map{
		"message":              "Bulk retry completed",
		"audiobook_id":         audiobookID,
		"total_jobs_found":     len(retriableJobs),
		"successfully_retried": len(retriedJobs),
		"failed_to_retry":      len(failedRetries),
		"retried_job_ids":      retriedJobs,
	}

	if len(failedRetries) > 0 {
		response["failed_job_ids"] = failedRetries
	}

	return c.JSON(response)
}
