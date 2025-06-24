package routes

import (
	"github.com/aditya-goyal-omniful/oms/pkg/controllers"
	"github.com/omniful/go_commons/http"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func InitServer(server *http.Server) {
	server.POST("/s3/filepath", controllers.StoreInS3)
	server.POST("/orders/bulkorder", controllers.CreateBulkOrder)
	server.POST("/orders", controllers.CreateOrder)
	server.GET("/orders", controllers.GetOrders)

	// Webhook Routes
	server.POST("webhooks/register", controllers.RegisterWebhook)

	// Swagger Routes
	server.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}