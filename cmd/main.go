package main

import (
	"log"

	"github.com/aditya-goyal-omniful/oms/context"
	"github.com/aditya-goyal-omniful/oms/pkg/routes"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/http"
)

func main() {
	ctx := context.GetContext()

	server := http.InitializeServer(
		config.GetString(ctx, "server.port"),
		config.GetDuration(ctx, "server.read_timeout"),  // Read timeout
		config.GetDuration(ctx, "server.write_timeout"), // Write timeout
		config.GetDuration(ctx, "server.idle_timeout"),  // Idle timeout
		false,
	)

	routes.InitServer(server)
	err := server.StartServer(config.GetString(ctx, "server.name"))
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}

}
