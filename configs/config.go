package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment        string
	APIGATEWAYPORT     string
	JWTSecretKey       string
	UserGRPCPort       string
	RestaurantGRPCPort string
	OrderCartGTPCPort  string
	AdminGRPCPort      string
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
		AdminGRPCPort:      os.Getenv("ADMINGRPCPORT"),
		Environment:        os.Getenv("ENVIRONMENT"),
	}
}
