package maintenance

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luna4dev/airlock/internal/model"
	"github.com/luna4dev/airlock/internal/service"
)

// UserServiceHandler struct holds dependencies for user service operations
type UserServiceHandler struct {
	sqliteService *service.SQLiteService
}

// NewUserServiceHandler creates a new user service handler with injected dependencies
func NewUserServiceHandler(sqliteService *service.SQLiteService) *UserServiceHandler {
	return &UserServiceHandler{
		sqliteService: sqliteService,
	}
}

// GetUserServices retrieves all services for a specific user
func (h *UserServiceHandler) GetUserServices(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	ctx := context.Background()

	// First check if user exists
	user, err := h.sqliteService.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get user's services
	services, err := h.sqliteService.GetUserServices(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user services"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"services": services,
		"count":    len(services),
	})
}

// AddUserServiceRequest represents the request payload for adding a service to a user
type AddUserServiceRequest struct {
	Service    string `json:"service" binding:"required"`
	Permission string `json:"permission" binding:"required"`
	ExpiresAt  *int64 `json:"expiresAt,omitempty"`
}

// AddUserService adds a service to a user
func (h *UserServiceHandler) AddUserService(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	var req AddUserServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing required fields"})
		return
	}

	ctx := context.Background()

	// Check if user exists
	user, err := h.sqliteService.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Validate service and permission values
	service := model.Luna4Service(req.Service)
	permission := model.UserServicePermission(req.Permission)

	// Create new user service
	userService := &model.Luna4UserService{
		ID:         uuid.New().String(),
		UserID:     userID,
		Service:    service,
		Permission: permission,
		ExpiresAt:  req.ExpiresAt,
	}

	// Add service to user
	err = h.sqliteService.CreateUserService(ctx, userService)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add service to user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Service added to user successfully",
		"user_id": userID,
		"service": userService,
	})
}

// RemoveUserService removes a service from a user
func (h *UserServiceHandler) RemoveUserService(c *gin.Context) {
	userID := c.Param("id")
	serviceID := c.Param("serviceId")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	if serviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service ID is required"})
		return
	}

	ctx := context.Background()

	// Check if user exists
	user, err := h.sqliteService.GetUserByID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if service exists and belongs to the user
	services, err := h.sqliteService.GetUserServices(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user services"})
		return
	}

	// Find the service to remove
	var serviceFound bool
	for _, service := range services {
		if service.ID == serviceID {
			serviceFound = true
			break
		}
	}

	if !serviceFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found for this user"})
		return
	}

	// Remove the service
	err = h.sqliteService.DeleteUserService(ctx, serviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove service from user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Service removed from user successfully",
		"user_id":    userID,
		"service_id": serviceID,
	})
}
