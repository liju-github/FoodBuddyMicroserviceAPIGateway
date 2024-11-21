package router

import (
	"github.com/gin-gonic/gin"
	adminPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Admin"
	orderCartPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/OrderCart"
	restaurantPb "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/Restaurant"
	user "github.com/liju-github/CentralisedFoodbuddyMicroserviceProto/User"
	"github.com/liju-github/FoodBuddyAPIGateway/clients"
	"github.com/liju-github/FoodBuddyAPIGateway/controller"
	"github.com/liju-github/FoodBuddyAPIGateway/middleware"
)

// InitializeServiceRoutes initializes gRPC clients for each service, creates controllers,
// and configures routes for each service.
func InitializeServiceRoutes(router *gin.Engine, Client *clients.ClientConnections) {
	// User Client setup
	userClient := user.NewUserServiceClient(Client.ConnUser)
	userController := controller.NewUserController(userClient)
	SetupUserRoutes(router, userController)

	// Restaurant Client setup
	restaurantClient := restaurantPb.NewRestaurantServiceClient(Client.ConnRestaurant)
	restaurantController := controller.NewRestaurantController(restaurantClient)
	SetupRestaurantRoutes(router, restaurantController)

	// OrderCart Client setup
	orderCartClient := orderCartPb.NewOrderCartServiceClient(Client.ConnOrderCart)
	orderCartController := controller.NewOrderCartController(
		orderCartClient,
		userClient,
		restaurantClient,
	)
	SetupOrderCartRoutes(router, orderCartController)

	//Admin Client setup
	adminClient := adminPb.NewAdminServiceClient(Client.ConnAdmin)
	adminController := controller.NewAdminController(adminClient)
	SetUpAdminAuth(router, adminController)
}

func SetUpAdminAuth(router *gin.Engine, adminController *controller.AdminController) {
	router.POST("/admin/login", adminController.AdminLogin)
}

// SetupUserRoutes configures routes for User-related operations
func SetupUserRoutes(router *gin.Engine, userController *controller.UserController) {
	// Public routes (no authentication required)
	auth := router.Group("/auth/user")
	{
		auth.POST("/signup", userController.Signup)
		auth.POST("/login", userController.Login)
		auth.POST("/verify-email", userController.VerifyEmail)
	}

	// Protected routes (require authentication)
	protected := router.Group("/api/users")
	protected.Use(middleware.JWTAuthMiddleware(), middleware.UserAuthMiddleware())
	{
		// Profile management
		profile := protected.Group("/profile")
		{
			profile.GET("", userController.GetProfile) // Query: none (uses JWT)
			profile.PUT("/update", userController.UpdateProfile)
		}

		// Address management
		address := protected.Group("/address")
		{
			address.POST("/add", userController.AddAddress)
			address.GET("/list", userController.GetAddresses) // Query: none (uses JWT)
			address.PUT("/update", userController.EditAddress)
			address.DELETE("/remove", userController.DeleteAddress)
		}
	}

	// Admin routes
	admin := router.Group("/admin/users")
	admin.Use(middleware.JWTAuthMiddleware(), middleware.AdminAuthMiddleware())
	{
		admin.GET("/list", userController.GetAllUsers) // Query: none
		admin.POST("/ban", userController.BanUser)
		admin.POST("/unban", userController.UnBanUser)
		admin.GET("/ban/status", userController.CheckBan) // Query: userId
	}
}

// SetupRestaurantRoutes configures routes for Restaurant-related operations
func SetupRestaurantRoutes(router *gin.Engine, restaurantController *controller.RestaurantController) {
	// Public routes (no authentication required)
	auth := router.Group("/auth/restaurant")
	{
		auth.POST("/signup", restaurantController.RestaurantSignup)
		auth.POST("/login", restaurantController.RestaurantLogin)
	}

	// Protected routes (require authentication)
	protected := router.Group("/api/restaurants")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		// Restaurant management (requires restaurant authentication)
		restaurant := protected.Group("")
		restaurant.Use(middleware.RestaurantAuthMiddleware())
		{
			restaurant.PUT("/profile/update", restaurantController.EditRestaurant)

			// Product management
			products := restaurant.Group("/products")
			{
				products.POST("/add", restaurantController.AddProduct)
				products.PUT("/update", restaurantController.EditProduct)
				products.DELETE("/remove", restaurantController.DeleteProductByID) // Query: productId
				products.PUT("/stock/increment", restaurantController.IncrementProductStock)
				products.PUT("/stock/decrement", restaurantController.DecrementProductStock)
			}
		}

		// Admin routes (requires admin authentication)
		admin := protected.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware())
		{
			admin.POST("/ban", restaurantController.BanRestaurant)
			admin.POST("/unban", restaurantController.UnbanRestaurant)
		}
	}

	// Public restaurant and product routes (no authentication required)
	public := router.Group("/api/public/restaurants")
	{
		public.GET("/list", restaurantController.GetAllRestaurantWithProducts)       // Query: none
		public.GET("/products/list", restaurantController.GetRestaurantProductsByID) // Query: restaurantId
		public.GET("/products/all", restaurantController.GetAllProducts)             // Query: none
		public.GET("/products/details", restaurantController.GetProductByID)         // Query: productId
		public.GET("/products/stock", restaurantController.GetStockByProductID)      // Query: productId
		public.GET("/lookup", restaurantController.GetRestaurantIDviaProductID)      // Query: productId
	}
}

// SetupOrderCartRoutes configures routes for OrderCart-related operations
func SetupOrderCartRoutes(router *gin.Engine, orderCartController *controller.OrderCartController) {
	// Cart Operations
	cart := router.Group("/api/cart")
	cart.Use(middleware.JWTAuthMiddleware(), middleware.UserAuthMiddleware())
	{
		cart.POST("/add", orderCartController.AddProductToCart)
		cart.GET("/items", orderCartController.GetCartItems) // Query: restaurantId (JWT for userId)
		cart.GET("/list", orderCartController.GetAllCarts)   // Query: none (JWT for userId)
		cart.POST("/increment", orderCartController.IncrementProductQuantity)
		cart.POST("/decrement", orderCartController.DecrementProductQuantity)
		cart.POST("/remove", orderCartController.RemoveProductFromCart)
		cart.POST("/clear", orderCartController.ClearCart)
	}

	// Order Operations for Users
	userOrder := router.Group("/api/orders")
	userOrder.Use(middleware.JWTAuthMiddleware(), middleware.UserAuthMiddleware())
	{
		userOrder.POST("/place", orderCartController.PlaceOrderByRestID)
		userOrder.GET("/list", orderCartController.GetOrderDetailsAll)     // Query: status (JWT for userId)
		userOrder.GET("/details", orderCartController.GetOrderDetailsByID) // Query: orderId (JWT for userId)
		userOrder.POST("/cancel", orderCartController.CancelOrder)
	}

	// Order Operations for Restaurants
	restaurantOrder := router.Group("/api/restaurant/orders")
	restaurantOrder.Use(middleware.JWTAuthMiddleware(), middleware.RestaurantAuthMiddleware())
	{
		restaurantOrder.PUT("/status/update", orderCartController.UpdateOrderStatus)
		restaurantOrder.GET("/list", orderCartController.GetRestaurantOrders) // Query: status (JWT for restaurantId)
		restaurantOrder.POST("/confirm", orderCartController.ConfirmOrder)
	}
}
