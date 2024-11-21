package clients

import (
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	config "github.com/liju-github/FoodBuddyAPIGateway/configs"
)

type ClientConnections struct {
	ConnUser       *grpc.ClientConn
	ConnRestaurant *grpc.ClientConn
	ConnAdmin       *grpc.ClientConn
	ConnOrderCart  *grpc.ClientConn
}

func InitClients(config *config.Config) (*ClientConnections, error) {
	// User Service Connection
	ConnUser, err := grpc.Dial("localhost:"+config.UserGRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("could not Connect to User gRPC server: " + err.Error())
	}

	// Restaurant Service Connection
	ConnRestaurant, err := grpc.Dial("localhost:"+config.RestaurantGRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		ConnUser.Close() // Close the user connection if restaurant connection fails
		return nil, errors.New("could not Connect to Restaurant gRPC server: " + err.Error())
	}

	// Admin Service Connection
	ConnAdmin, err := grpc.Dial("localhost:"+config.AdminGRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		ConnUser.Close()   // Close the user connection if admin connection fails
		ConnRestaurant.Close() // Close the restaurant connection if admin connection fails
		return nil, errors.New("could not Connect to Admin gRPC server: " + err.Error())
	}

	// OrderCart Service Connection
	ConnOrderCart, err := grpc.Dial("localhost:"+config.OrderCartGRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		ConnUser.Close()   // Close the user connection if admin connection fails
		ConnRestaurant.Close() // Close the restaurant connection if admin connection fails
		return nil, errors.New("could not Connect to Admin gRPC server: " + err.Error())
	}

	return &ClientConnections{
		ConnUser:       ConnUser,
		ConnRestaurant: ConnRestaurant,
		ConnAdmin:      ConnAdmin,
		ConnOrderCart:  ConnOrderCart,
	}, nil
}

func (c *ClientConnections) Close() {
	if c.ConnUser != nil {
		c.ConnUser.Close()
	}
	if c.ConnRestaurant != nil {
		c.ConnRestaurant.Close()
	}
	if c.ConnAdmin != nil {
		c.ConnAdmin.Close()
	}
	if c.ConnOrderCart != nil {
		c.ConnOrderCart.Close()
	}
}
