package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/luna4dev/airlock/internal/handler"
	"github.com/luna4dev/airlock/internal/handler/maintenance"
	"github.com/luna4dev/airlock/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed web/*
var webFS embed.FS

//go:embed assets/templates/*
var templateFS embed.FS

//go:embed configs/sqlite-schema/*
var sqliteSchemaFS embed.FS

//go:embed configs/sqlite-migration/*
var sqliteMigrationFS embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// inject embed to services package
	service.TemplateFS = templateFS

	// Initialize SQLite service once
	sqliteService, err := service.NewSQLiteService("data/airlock.db", &sqliteSchemaFS, &sqliteMigrationFS)
	if err != nil {
		log.Fatal("Failed to initialize SQLite service:", err)
	}
	defer sqliteService.Close()

	// Initialize handlers with dependencies
	userHandler := maintenance.NewUserHandler(sqliteService)
	userServiceHandler := maintenance.NewUserServiceHandler(sqliteService)
	authHandler := handler.NewAuthHandler(sqliteService)

	router := gin.Default()

	router.GET("/", redirectToApp)
	router.GET("/health", healthCheck)

	// Serve embedded static files
	staticFS, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatal("Failed to create sub filesystem:", err)
	}
	router.StaticFS("/app", http.FS(staticFS))

	// API routes
	api := router.Group("/api")
	{
		// Authentication endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/email", authHandler.AuthEmailHandler)
			auth.GET("/email/verify", authHandler.AuthEmailVerifyHandler)
		}

		// Maintenance endpoints
		maintenance := api.Group("/maintenance")
		{
			// User management
			maintenance.GET("/user", userHandler.GetUsers)
			maintenance.GET("/user/:id", userHandler.GetUser)
			maintenance.POST("/user", userHandler.CreateUser)
			maintenance.PUT("/user/:id/suspend", userHandler.SuspendUser)
			maintenance.PUT("/user/:id/activate", userHandler.ActivateUser)
			maintenance.DELETE("/user/:id", userHandler.DeleteUser)

			// User service management
			maintenance.GET("/user/:id/service", userServiceHandler.GetUserServices)
			maintenance.POST("/user/:id/service", userServiceHandler.AddUserService)
			maintenance.DELETE("/user/:id/service/:serviceId", userServiceHandler.RemoveUserService)
		}
	}

	port := os.Getenv("PORT")
	router.Run(":" + port)
}

func redirectToApp(c *gin.Context) {
	// Build redirect URL with query parameters
	redirectURL := "/app"
	if c.Request.URL.RawQuery != "" {
		redirectURL += "?" + c.Request.URL.RawQuery
	}

	// Copy headers to the redirect response
	for key, values := range c.Request.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Redirect(http.StatusMovedPermanently, redirectURL)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "airlock",
	})
}
