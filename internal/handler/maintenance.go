package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/luna4dev/airlock/internal/model"
	"github.com/luna4dev/airlock/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MaintenanceHandler struct holds the SQLite service dependency
type MaintenanceHandler struct {
	sqliteService *service.SQLiteService
}

// NewMaintenanceHandler creates a new maintenance handler with injected dependencies
func NewMaintenanceHandler(sqliteService *service.SQLiteService) *MaintenanceHandler {
	return &MaintenanceHandler{
		sqliteService: sqliteService,
	}
}

// Status handles maintenance status requests
func (h *MaintenanceHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Maintenance endpoint",
		"status":  "operational",
	})
}

// GetUsers returns all users using injected SQLite service
func (h *MaintenanceHandler) GetUsers(c *gin.Context) {
	ctx := context.Background()
	users, err := h.sqliteService.GetAllUsers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"count": len(users),
	})
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Email  string `json:"email" binding:"required"`
	Status string `json:"status"`
}

// CreateUser creates a new user using injected SQLite service
func (h *MaintenanceHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON or missing required fields"})
		return
	}

	// Validate status or set default
	status := model.UserStatusActive
	if req.Status != "" {
		if req.Status == string(model.UserStatusActive) || req.Status == string(model.UserStatusSuspended) {
			status = model.UserStatus(req.Status)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status. Must be ACTIVE or SUSPENDED"})
			return
		}
	}

	// Create new user
	user := &model.Luna4User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Status:    status,
		CreatedAt: time.Now().UnixMilli(),
		UpdatedAt: time.Now().UnixMilli(),
	}

	ctx := context.Background()
	err := h.sqliteService.CreateUser(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

// SuspendUser sets a user's status to suspended
func (h *MaintenanceHandler) SuspendUser(c *gin.Context) {
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

	// Update user status to suspended
	err = h.sqliteService.UpdateUserStatus(ctx, userID, model.UserStatusSuspended)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to suspend user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User suspended successfully",
		"user_id": userID,
		"status":  string(model.UserStatusSuspended),
	})
}

// ActivateUser sets a user's status to active
func (h *MaintenanceHandler) ActivateUser(c *gin.Context) {
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

	// Update user status to active
	err = h.sqliteService.UpdateUserStatus(ctx, userID, model.UserStatusActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User activated successfully",
		"user_id": userID,
		"status":  string(model.UserStatusActive),
	})
}

// DeleteUser permanently deletes a user (only if suspended)
func (h *MaintenanceHandler) DeleteUser(c *gin.Context) {
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

	// Check if user is suspended before allowing deletion
	if user.Status != model.UserStatusSuspended {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User must be suspended before deletion",
			"current_status": string(user.Status),
		})
		return
	}

	// Delete the user
	err = h.sqliteService.DeleteUser(ctx, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
		"user_id": userID,
	})
}