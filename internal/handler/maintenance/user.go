package maintenance

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luna4dev/airlock/internal/model"
	"github.com/luna4dev/airlock/internal/service"
)

// UserHandler struct holds the SQLite service dependency
type UserHandler struct {
	sqliteService *service.SQLiteService
}

// NewUserHandler creates a new user handler with injected dependencies
func NewUserHandler(sqliteService *service.SQLiteService) *UserHandler {
	return &UserHandler{
		sqliteService: sqliteService,
	}
}

// Status handles maintenance status requests
func (h *UserHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Maintenance endpoint",
		"status":  "operational",
	})
}

// GetUsers returns all users with their services using injected SQLite service
func (h *UserHandler) GetUsers(c *gin.Context) {
	ctx := context.Background()
	users, err := h.sqliteService.GetAllUsers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
		return
	}

	// Create response with users and their services
	type UserWithServices struct {
		*model.Luna4User
		Services []model.Luna4UserService `json:"services"`
	}

	var usersWithServices []UserWithServices
	for _, user := range users {
		services, err := h.sqliteService.GetUserServices(ctx, user.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user services"})
			return
		}

		userWithServices := UserWithServices{
			Luna4User: user,
			Services:  services,
		}
		usersWithServices = append(usersWithServices, userWithServices)
	}

	c.JSON(http.StatusOK, gin.H{
		"users": usersWithServices,
		"count": len(usersWithServices),
	})
}

// GetUser returns a single user with their services by ID
func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	ctx := context.Background()

	// Get user by ID
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

	// Create response with user and services
	type UserWithServices struct {
		*model.Luna4User
		Services []model.Luna4UserService `json:"services"`
	}

	userWithServices := UserWithServices{
		Luna4User: user,
		Services:  services,
	}

	c.JSON(http.StatusOK, gin.H{
		"user": userWithServices,
	})
}

// CreateUserRequest represents the request payload for creating a user
type CreateUserRequest struct {
	Email    string                     `json:"email" binding:"required"`
	Status   string                     `json:"status"`
	Services []CreateUserServiceRequest `json:"services,omitempty"`
}

// CreateUserServiceRequest represents service permissions for user creation
type CreateUserServiceRequest struct {
	Service    string `json:"service"`
	Permission string `json:"permission"`
	ExpiresAt  *int64 `json:"expiresAt,omitempty"`
}

// CreateUser creates a new user using injected SQLite service
func (h *UserHandler) CreateUser(c *gin.Context) {
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
	userID := uuid.New().String()
	user := &model.Luna4User{
		ID:        userID,
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

	// Handle services - use provided services or default PRUNK-USER
	var servicesToCreate []model.Luna4UserService
	if len(req.Services) > 0 {
		// Use provided services
		for _, svc := range req.Services {
			userService := model.Luna4UserService{
				ID:         uuid.New().String(),
				UserID:     userID,
				Service:    model.Luna4Service(svc.Service),
				Permission: model.UserServicePermission(svc.Permission),
				ExpiresAt:  svc.ExpiresAt,
			}
			servicesToCreate = append(servicesToCreate, userService)
		}
	} else {
		// Create default PRUNK-USER service
		defaultService := model.Luna4UserService{
			ID:         uuid.New().String(),
			UserID:     userID,
			Service:    model.Luna4ServicePrunk,
			Permission: model.UserServiceUser,
			ExpiresAt:  nil, // No expiry
		}
		servicesToCreate = append(servicesToCreate, defaultService)
	}

	// Create user services
	for _, userService := range servicesToCreate {
		err = h.sqliteService.CreateUserService(ctx, &userService)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user service"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "User created successfully",
		"user":     user,
		"services": servicesToCreate,
	})
}

// SuspendUser sets a user's status to suspended
func (h *UserHandler) SuspendUser(c *gin.Context) {
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
func (h *UserHandler) ActivateUser(c *gin.Context) {
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
func (h *UserHandler) DeleteUser(c *gin.Context) {
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
			"error":          "User must be suspended before deletion",
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
