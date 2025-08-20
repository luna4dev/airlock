package main

import (
	"log"
	"net/http"
	"os"

	"airlock/handler"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	router := gin.Default()

	router.GET("/", healthCheck)

	// Serve static files from root
	router.Static("/app", "./public")

	// API routes
	api := router.Group("/api")
	{
		api.POST("/auth/email", handler.AuthEmailHandler)
		api.GET("/auth/email/verify", handler.AuthEmailVerifyHandler)
	}

	port := os.Getenv("PORT")
	router.Run(":" + port)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "airlock",
	})
}
