package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	APIGATEWAYPORT     string
	JWTSecretKey       string
	UserGRPCPort       string
	RestaurantGRPCPort string
	OrderCartGTPCPort  string
}

func LoadConfig() Config {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return Config{
		APIGATEWAYPORT:     os.Getenv("APIGATEWAYPORT"),
		JWTSecretKey:       os.Getenv("JWTSECRET"),
		UserGRPCPort:       os.Getenv("USERGRPCPORT"),
		RestaurantGRPCPort: os.Getenv("RESTAURANTGRPCPORT"),
		OrderCartGTPCPort:  os.Getenv("ORDERCARTGRPCPORT"),
	}
}
