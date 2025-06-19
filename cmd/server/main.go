package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/vivaswanth-kashyap/tchat-server/internal/db"
	"github.com/vivaswanth-kashyap/tchat-server/internal/handlers"
	"github.com/vivaswanth-kashyap/tchat-server/internal/models"
	"github.com/vivaswanth-kashyap/tchat-server/utils"
	"gorm.io/gorm"
)

type UserIDContextKey string

const CtxUserID UserIDContextKey = "userID"

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer") {
			log.Println("AuthMiddleware: Missing or malformed Authorization header")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or expired token"})
			return
		}

		jwtToken := strings.TrimPrefix(tokenString, "Bearer ")
		jwtToken = strings.TrimSpace(jwtToken)
		log.Printf("DEBUG: AuthMiddleware received raw JWT token (length %d): %q", len(jwtToken), jwtToken)
		log.Printf("DEBUG: AuthMiddleware received raw JWT token (hex): %x", []byte(jwtToken))

		claims, err := utils.ParseAccessToken(jwtToken)
		if err != nil {
			log.Printf("AuthMiddleware: Failed to parse or validate access token: %v", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Invalid or expired token"})
			return
		}

		c.Set(string(CtxUserID), claims.UserID)
		c.Next()
	}
}

func setupRoutes(router *gin.Engine, database *gorm.DB) {

	authHandlers := &handlers.AuthHandler{DB: database}

	router.POST("/signup", authHandlers.Signup)
	router.POST("/login", authHandlers.Login)

	protectedAPI := router.Group("/api")
	protectedAPI.Use(AuthMiddleware())
	{
		protectedAPI.GET("/user/me", func(c *gin.Context) {
			userID, exists := c.Get(string(CtxUserID))
			if !exists {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "User ID not found in context after authentication"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Authenticated successfully!", "user_id": userID})
		})

		messageHandlers := handlers.NewMessageHandlers(database)

		msgs := protectedAPI.Group("/messages")
		{
			msgs.POST("", messageHandlers.NewMessage)
			msgs.GET("", messageHandlers.ReadMessage)
			msgs.GET("/chat", messageHandlers.ReadChat)
			msgs.GET("/last", messageHandlers.ReadLastSent)
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file in main.go:", err) // Fatal if .env doesn't load
	}

	database, err := db.ConnectDb()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying SQL DB: %v", err)
	}
	defer sqlDB.Close()

	err = database.AutoMigrate(&models.User{}, &models.Message{}, &models.RefreshToken{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	log.Println("Database Migrations completed: Users, Messages, RefreshTokens")

	router := gin.Default()

	setupRoutes(router, database)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port: %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
