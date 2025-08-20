package handler

import (
	"context"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"airlock/service"
	"airlock/util"
	"github.com/gin-gonic/gin"
)

type AuthEmailRequest struct {
	Email string `json:"email" binding:"required"`
}

func AuthEmailHandler(c *gin.Context) {
	var req AuthEmailRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing email field"})
		return
	}
	
	email := strings.TrimSpace(req.Email)
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}
	
	if !isValidEmail(email) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}
	
	userService, err := service.NewUserService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	
	ctx := context.Background()
	user, err := userService.GetUserByEmail(ctx, email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query database"})
		return
	}
	
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	// Check if email auth exists and validate debounce
	if user.EmailAuth != nil {
		debounceSeconds := getEmailAuthDebounce()
		timeSinceLastSent := time.Now().UnixMilli() - user.EmailAuth.SentAt
		debounceMillis := int64(debounceSeconds * 1000)
		
		if timeSinceLastSent < debounceMillis {
			remainingSeconds := (debounceMillis - timeSinceLastSent) / 1000
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Email authentication request too recent",
				"retry_after_seconds": remainingSeconds,
			})
			return
		}
	}
	
	// Generate new email token
	token, err := util.GenerateEmailToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authentication token"})
		return
	}
	
	// Update user's email auth in database
	err = userService.UpdateUserEmailAuth(ctx, user.ID, token)
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
	
	err = emailService.SendAuthEmail(ctx, email, token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send authentication email"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"email": email,
		"message": "Authentication email sent successfully",
	})
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

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