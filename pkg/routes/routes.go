package routes

import (
	"github.com/aditya-goyal-omniful/oms/pkg/controllers"
	"github.com/omniful/go_commons/http"
)

func InitServer(server *http.Server) {
	server.GET("/", controllers.ServeHome)

	v1 := server.Engine.Group("/api/v1")
	{
		orders := v1.Group("/order") 				// Contains csv file path
		{
			orders.POST("/bulkorder", controllers.CreateBulkOrder)
		}

		csv := v1.Group("/csv")
		{
			csv.POST("/filepath", controllers.StoreInS3)
		}
	}
}