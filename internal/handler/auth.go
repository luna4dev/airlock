package handler

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/luna4dev/airlock/internal/model"
	"github.com/luna4dev/airlock/internal/service"
	"github.com/luna4dev/airlock/internal/util"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	sqliteService *service.SQLiteService
}

func NewAuthHandler(sqliteService *service.SQLiteService) *AuthHandler {
	return &AuthHandler{
		sqliteService: sqliteService,
	}
}

// AuthEmailRequest represents the request payload for email authentication
type AuthEmailRequest struct {
	Email    string `json:"email" binding:"required"`
	Redirect string `json:"redirect"`
}

// AuthEmailHandler handles the initial email authentication request
func (h *AuthHandler) AuthEmailHandler(c *gin.Context) {
	var req AuthEmailRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing email field"})
		return
	}

	email := strings.TrimSpace(req.Email)
	redirect := strings.TrimSpace(req.Redirect)
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	if !isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	ctx := context.Background()
	user, err := h.sqliteService.GetUserByEmail(ctx, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get User"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	latestEmailAuth, _ := h.sqliteService.GetLatestEmailAuth(ctx, user.ID)
	if latestEmailAuth != nil {
		debounceSeconds := getEmailAuthDebounce()
		timeSinceLastSent := time.Now().UnixMilli() - latestEmailAuth.SentAt
		debounceMillis := int64(debounceSeconds * 1000)

		if timeSinceLastSent < debounceMillis {
			remainingSeconds := (debounceMillis - timeSinceLastSent) / 1000
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":               "Email authentication request too recent",
				"retry_after_seconds": remainingSeconds,
			})
			return
		}
	}

	// Generate new email token, tokenHash
	token, tokenHash, err := util.GenerateEmailToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}

	// Update user's email auth in database
	// Create new user
	emailAuthID := uuid.New().String()
	emailAuth := &model.Luna4EmailAuth{
		ID:        emailAuthID,
		UserID:    user.ID,
		Token:     tokenHash,
		SentAt:    time.Now().UnixMilli(),
		Completed: false,
	}

	err = h.sqliteService.CreateEmailAuth(ctx, emailAuth)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user authentication"})
		return
	}

	// Send authentication email
	emailService, err := service.NewEmailService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize email service"})
		return
	}

	err = emailService.SendAuthEmail(ctx, email, token, redirect)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send authentication email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":   email,
		"message": "Authentication email sent successfully",
	})
}

// isValidEmail validates if the provided email string is in a valid format
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// AuthEmailVerifyHandler handles email verification and token validation
func (h *AuthHandler) AuthEmailVerifyHandler(c *gin.Context) {
	token := c.Query("token")
	email := c.Query("email")

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	email = strings.TrimSpace(email)
	if !isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	ctx := context.Background()
	user, err := h.sqliteService.GetUserByEmail(ctx, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query database"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	latestEmailAuth, err := h.sqliteService.GetLatestEmailAuth(ctx, user.ID)
	if err != nil {

	}

	// Check if email auth exists
	if latestEmailAuth == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No email authentication token found"})
		return
	}

	// Check if token matches
	if !util.VerifyEmailToken(token, latestEmailAuth.Token) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authentication token"})
		return
	}

	// Check if token has expired
	expirySeconds := getEmailAuthExpiry()
	timeSinceTokenSent := time.Now().UnixMilli() - latestEmailAuth.SentAt
	expiryMillis := int64(expirySeconds * 1000)

	if timeSinceTokenSent > expiryMillis {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication token has expired"})
		return
	}

	// Check if token has already been completed
	if latestEmailAuth.Completed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authentication token has already been used"})
		return
	}

	// Complete email authentication and update lastLoginAt
	err = h.sqliteService.MarkEmailAuthCompleted(ctx, latestEmailAuth.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete authentication"})
		return
	}

	// Generate JWT bearer token
	bearerToken, err := util.GenerateBearerToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Email verification successful",
		"access_token": bearerToken,
		"token_type":   "Bearer",
		"expires_in":   30 * 24 * 60 * 60, // 30 days in seconds
		"user": gin.H{
			"id":     user.ID,
			"email":  user.Email,
			"status": user.Status,
		},
	})
}

// getEmailAuthDebounce returns the debounce time in seconds for email authentication requests
func getEmailAuthDebounce() int {
	debounceStr := os.Getenv("EMAIL_AUTH_DEBOUNCE")
	if debounceStr == "" {
		return 180 // Default 3 minutes
	}

	debounce, err := strconv.Atoi(debounceStr)
	if err != nil {
		return 180 // Default 3 minutes on error
	}

	return debounce
}

// getEmailAuthExpiry returns the expiry time in seconds for email authentication tokens
func getEmailAuthExpiry() int {
	expiryStr := os.Getenv("EMAIL_AUTH_EXPIRY")
	if expiryStr == "" {
		return 900 // Default 15 minutes
	}

	expiry, err := strconv.Atoi(expiryStr)
	if err != nil {
		return 900 // Default 15 minutes on error
	}

	return expiry
}
