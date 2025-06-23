package routes

import (
	"github.com/aditya-goyal-omniful/oms/pkg/controllers"
	"github.com/omniful/go_commons/http"
)

func InitServer(server *http.Server) {
	server.POST("/orderS/bulkorder", controllers.CreateBulkOrder)
	server.POST("/s3/filepath", controllers.StoreInS3)
	server.POST("/orders", controllers.CreateOrder)
	server.GET("/orders", controllers.GetOrders)
}