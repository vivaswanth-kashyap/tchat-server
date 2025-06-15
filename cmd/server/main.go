package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/vivaswanth-kashyap/tchat-server/internal/db"
	"github.com/vivaswanth-kashyap/tchat-server/internal/handlers"
	"github.com/vivaswanth-kashyap/tchat-server/internal/models"
	"gorm.io/gorm"
)

func setupRoutes(router *gin.Engine, database *gorm.DB) {
	messageHandlers := handlers.NewMessageHandlers(database)

	msgs := router.Group("/messages")
	{
		msgs.POST("", messageHandlers.NewMessage)
		msgs.GET("", messageHandlers.ReadMessage)
		msgs.GET("/chat", messageHandlers.ReadChat)
		msgs.GET("/last", messageHandlers.ReadLastSent)
	}
}

func main() {
	database, err := db.ConnectDb()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("Failed to get underlying SQL DB: %v", err)
	}
	defer sqlDB.Close()

	err = database.AutoMigrate(&models.User{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	log.Println("Users Database Migration completed")

	err = database.AutoMigrate(&models.Message{})
	if err != nil {
		log.Fatalf("Failed to auto migrate database: %v", err)
	}
	log.Println("Messages Database Migration completed")

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
