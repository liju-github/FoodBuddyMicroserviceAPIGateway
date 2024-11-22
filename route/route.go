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

func InitializeServiceRoutes(router *gin.Engine, Client *clients.ClientConnections) {
	userClient := user.NewUserServiceClient(Client.ConnUser)
	userController := controller.NewUserController(userClient)
	SetupUserRoutes(router, userController)

	restaurantClient := restaurantPb.NewRestaurantServiceClient(Client.ConnRestaurant)
	restaurantController := controller.NewRestaurantController(restaurantClient)
	SetupRestaurantRoutes(router, restaurantController)

	orderCartClient := orderCartPb.NewOrderCartServiceClient(Client.ConnOrderCart)
	orderCartController := controller.NewOrderCartController(
		orderCartClient,
		userClient,
		restaurantClient,
	)
	SetupOrderCartRoutes(router, orderCartController)

	adminClient := adminPb.NewAdminServiceClient(Client.ConnAdmin)
	adminController := controller.NewAdminController(adminClient)
	SetUpAdminAuth(router, adminController)
}

func SetUpAdminAuth(router *gin.Engine, adminController *controller.AdminController) {
	router.POST("/admin/login", adminController.AdminLogin)
}

func SetupUserRoutes(router *gin.Engine, userController *controller.UserController) {
	auth := router.Group("/auth/user")
	{
		auth.POST("/signup", userController.Signup)
		auth.POST("/login", userController.Login)
		auth.POST("/verify-email", userController.VerifyEmail)
	}

	protected := router.Group("/api/users")
	protected.Use(middleware.JWTAuthMiddleware(), middleware.UserAuthMiddleware(), middleware.UserBanCheckMiddleware(userController.GetUserClient()))
	{
		profile := protected.Group("/profile")
		{
			profile.GET("", userController.GetProfile)
			profile.PUT("/update", userController.UpdateProfile)
		}

		address := protected.Group("/address")
		{
			address.POST("/add", userController.AddAddress)
			address.GET("/list", userController.GetAddresses)
			address.PUT("/update", userController.EditAddress)
			address.DELETE("/remove/:addressId", userController.DeleteAddress)
		}
	}

	admin := router.Group("/admin/users")
	admin.Use(middleware.JWTAuthMiddleware(), middleware.AdminAuthMiddleware())
	{
		admin.GET("/list", userController.GetAllUsers)
		admin.POST("/ban", userController.BanUser)
		admin.POST("/unban", userController.UnBanUser)
		admin.GET("/ban/status", userController.CheckBan)
	}
}

func SetupRestaurantRoutes(router *gin.Engine, restaurantController *controller.RestaurantController) {
	auth := router.Group("/auth/restaurant")
	{
		auth.POST("/signup", restaurantController.RestaurantSignup)
		auth.POST("/login", restaurantController.RestaurantLogin)
	}

	protected := router.Group("/api/restaurants")
	protected.Use(middleware.JWTAuthMiddleware())
	{
		restaurant := protected.Group("")
		restaurant.Use(middleware.RestaurantAuthMiddleware())
		{
			restaurant.PUT("/profile/update", restaurantController.EditRestaurant)

			products := restaurant.Group("/products")
			{
				products.POST("/add", restaurantController.AddProduct)
				products.PUT("/update", restaurantController.EditProduct)
				products.DELETE("/remove", restaurantController.DeleteProductByID)
				products.PUT("/stock/increment", restaurantController.IncrementProductStock)
				products.PUT("/stock/decrement", restaurantController.DecrementProductStock)
			}
		}

		admin := protected.Group("/admin")
		admin.Use(middleware.AdminAuthMiddleware())
		{
			admin.POST("/ban", restaurantController.BanRestaurant) 
			admin.POST("/unban", restaurantController.UnbanRestaurant)
		}
	}

	public := router.Group("/api/public/restaurants")
	{
		public.GET("/list", restaurantController.GetAllRestaurantWithProducts)
		public.GET("/products/list", restaurantController.GetRestaurantProductsByID)
		public.GET("/products/all", restaurantController.GetAllProducts)
		public.GET("/products/details", restaurantController.GetProductByID)
		public.GET("/products/stock", restaurantController.GetStockByProductID)
		public.GET("/lookup", restaurantController.GetRestaurantIDviaProductID)
	}
}

func SetupOrderCartRoutes(router *gin.Engine, orderCartController *controller.OrderCartController) {
	cart := router.Group("/api/cart")
	cart.Use(middleware.JWTAuthMiddleware(), middleware.UserAuthMiddleware())
	{
		cart.POST("/add", orderCartController.AddProductToCart)
		cart.GET("/items", orderCartController.GetCartItems)
		cart.GET("/list", orderCartController.GetAllCarts)
		cart.POST("/increment", orderCartController.IncrementProductQuantity)
		cart.POST("/decrement", orderCartController.DecrementProductQuantity)
		cart.POST("/remove", orderCartController.RemoveProductFromCart)
		cart.POST("/clear", orderCartController.ClearCart)
	}

	userOrder := router.Group("/api/orders")
	userOrder.Use(middleware.JWTAuthMiddleware(), middleware.UserAuthMiddleware())
	{
		userOrder.POST("/place", orderCartController.PlaceOrderByRestID)
		userOrder.GET("/list", orderCartController.GetOrderDetailsAll)
		userOrder.GET("/details", orderCartController.GetOrderDetailsByID)
		userOrder.POST("/cancel", orderCartController.CancelOrder)
	}

	restaurantOrder := router.Group("/api/restaurant/orders")
	restaurantOrder.Use(middleware.JWTAuthMiddleware(), middleware.RestaurantAuthMiddleware())
	{
		restaurantOrder.GET("/list", orderCartController.GetRestaurantOrders)
		restaurantOrder.POST("/confirm", orderCartController.ConfirmOrder)
	}
}
