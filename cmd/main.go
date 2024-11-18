package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/liju-github/FoodBuddyAPIGateway/configs"
	"github.com/liju-github/FoodBuddyAPIGateway/clients"
)

func main() {
	// Load environment variables
	config := config.LoadConfig()

	// Initialize gRPC clients
	Client, err := clients.InitClients(config)
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer Client.Close()

	// Create a new Gin router
	ginRouter := gin.Default()

	// Setup all routes
	ginRouter.InitializeServiceRoutes(ginRouter, Client)

	// Start the HTTP server (API Gateway)
	log.Printf("API Gateway is running on port %s", config.HTTPPort)
	if err := ginRouter.Run(":" + config.HTTPPort); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
