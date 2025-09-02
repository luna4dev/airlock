package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/luna4dev/airlock/internal/handler"
	"github.com/luna4dev/airlock/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

//go:embed web/*
var webFS embed.FS

//go:embed assets/templates/*
var templateFS embed.FS

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// inject embed to services package
	service.TemplateFS = templateFS

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
		api.POST("/auth/email", handler.AuthEmailHandler)
		api.GET("/auth/email/verify", handler.AuthEmailVerifyHandler)
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
