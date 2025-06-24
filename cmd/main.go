package main

import (
	"github.com/aditya-goyal-omniful/oms/context"
	"github.com/aditya-goyal-omniful/oms/docs"
	"github.com/aditya-goyal-omniful/oms/pkg/initializers"
	"github.com/aditya-goyal-omniful/oms/pkg/middlewares"
	"github.com/aditya-goyal-omniful/oms/pkg/routes"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
)

// @title Order Management Service
// @version 1.0
// @description This is the OMS for managing orders.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@omniful.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8082
// @BasePath /

func main() {
	ctx := context.GetContext()

	initializers.InitServices(ctx)

	// Swagger metadata
	docs.SwaggerInfo.Title = "Order Management Service"
	docs.SwaggerInfo.Description = "API documentation for OMS"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8082"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http"}

	server := http.InitializeServer(
		config.GetString(ctx, "server.port"),
		config.GetDuration(ctx, "server.read_timeout"),  // Read timeout
		config.GetDuration(ctx, "server.write_timeout"), // Write timeout
		config.GetDuration(ctx, "server.idle_timeout"),  // Idle timeout
		false,
	)

	server.Use(middlewares.RequestLogger(ctx))
	server.Static("/public", "./public")

	routes.InitServer(server)
	err := server.StartServer(config.GetString(ctx, "server.name"))
	if err != nil {
		log.Panic(i18n.Translate(ctx, "Failed to start server: "), err)
	}

}
