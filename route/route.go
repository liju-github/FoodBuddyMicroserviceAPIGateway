package router

import (
	"github.com/gin-gonic/gin"
	"github.com/liju-github/FoodBuddyAPIGateway/clients"
	"github.com/liju-github/FoodBuddyAPIGateway/controller"
	"github.com/liju-github/FoodBuddyAPIGateway/middleware"

	// "github.com/liju-github/FoodBuddyAPIGateway/proto/admin"
	"github.com/liju-github/FoodBuddyAPIGateway/proto/content"
	// "github.com/liju-github/FoodBuddyAPIGateway/proto/notification"
	"github.com/liju-github/FoodBuddyAPIGateway/proto/user"
)

// InitializeServiceRoutes initializes gRPC clients for each service, creates controllers,
// and configures routes for each service.
func InitializeServiceRoutes(router *gin.Engine, Client *clients.ClientConnections) {
	// User Client setup
	userClient := user.NewUserServiceClient(Client.ConnUser)
	userController := controller.NewUserController(userClient)
	SetupUserRoutes(router, userController)


	// // Admin Client setup
	// adminClient := admin.NewAdminServiceClient(Client.ConnAdmin)
	// adminController := controller.NewAdminController(adminClient)
	// SetupAdminRoutes(router, adminController)

	// // Notification Client setup
	// notificationClient := notification.NewNotificationServiceClient(Client.ConnNotification)
	// notificationController := controller.NewNotificationController(notificationClient)
	// SetupNotificationRoutes(router, notificationController)
}

// SetupUserRoutes configures routes for User-related operations
func SetupUserRoutes(router *gin.Engine, userController *controller.UserController) {
	// Public routes
	router.POST("/register", userController.RegisterHandler)
	router.POST("/login", userController.LoginHandler)

	// userProtected routes with JWT middleware
	userProtected := router.Group("/")
	userProtected.Use(middleware.JWTAuthMiddleware)
	{
		userProtected.GET("/profile", userController.GetProfileHandler)
		userProtected.PATCH("/update-profile", userController.UpdateProfileHandler)
	}

}
