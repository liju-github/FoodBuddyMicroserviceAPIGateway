package clients

import (
	"errors"

	config "github.com/liju-github/FoodBuddyAPIGateway/configs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ClientConnections struct {
	ConnUser         *grpc.ClientConn
	ConnContent      *grpc.ClientConn
	ConnAdmin        *grpc.ClientConn
	ConnNotification *grpc.ClientConn
}

func InitClients(config config.Config) (*ClientConnections, error) {
	ConnUser, err := grpc.NewClient("localhost:"+config.UserGRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.New("could not Connect to User gRPC server: " + err.Error())
	}

	// ConnContent, err := grpc.NewClient("localhost:"+config.ContentGRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	ConnUser.Close() 
	// 	return nil, errors.New("could not Connect to Content gRPC server: " + err.Error())
	// }

	// ConnAdmin, err := grpc.NewClient("localhost:"+config.AdminGRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	ConnUser.Close()
	// 	ConnContent.Close()
	// 	return nil, errors.New("could not Connect to Admin gRPC server: " + err.Error())
	// }

	// ConnNotification, err := grpc.NewClient("localhost:"+config.NotificationGRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	ConnUser.Close()
	// 	ConnContent.Close()
	// 	ConnAdmin.Close()
	// 	return nil, errors.New("could not Connect to Notification gRPC server: " + err.Error())
	// }

	return &ClientConnections{
		ConnUser:         ConnUser,
		// ConnContent:      ConnContent,
		// ConnAdmin:        ConnAdmin,
		// ConnNotification: ConnNotification,
	}, nil
}


func (c *ClientConnections) Close() {
	connections := []*grpc.ClientConn{c.ConnUser, c.ConnContent, c.ConnAdmin, c.ConnNotification}
	for _, conn := range connections {
		if conn != nil {
			conn.Close()
		}
	}
}