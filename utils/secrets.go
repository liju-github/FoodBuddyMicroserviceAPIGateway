package utils

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// EnvVariables holds all the environment variables for the app.
type EnvVariables struct {
	JWTSecret              string
	APIGatewayURL          string
	AuthServiceURL         string
	UserServiceURL         string
	RestaurantServiceURL   string
	ProductServiceURL      string
	CartOrderServiceURL    string
	PaymentServiceURL      string
	NotificationServiceURL string
}

// LoadSecrets loads environment variables from the .env file and returns a populated EnvVariables struct.
func LoadSecrets() EnvVariables {
	// Load .env file from the current working directory
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found or error loading .env file: %v", err)
	}

	// Return the EnvVariables struct populated with values from environment
	return EnvVariables{
		JWTSecret:              os.Getenv("JWTSECRET"),
		APIGatewayURL:          os.Getenv("APIGATEWAYURL"),
		AuthServiceURL:         os.Getenv("AUTHSERVICEURL"),
		UserServiceURL:         os.Getenv("USERSERVICEURL"),
		RestaurantServiceURL:   os.Getenv("RESTAURANTSERVICEURL"),
		ProductServiceURL:      os.Getenv("PRODUCTSERVICEURL"),
		CartOrderServiceURL:    os.Getenv("CARTORDERSERVICEURL"),
		PaymentServiceURL:      os.Getenv("PAYMENTSERVICEURL"),
		NotificationServiceURL: os.Getenv("NOTIFICATIONSERVICEURL"),
	}
}
