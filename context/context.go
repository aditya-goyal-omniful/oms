package context

import (
	"context"
	"log"
	"time"

	"github.com/omniful/go_commons/config"
)

var ctx context.Context

func init() {
	err := config.Init(time.Second * 10) 		// loads the config.yaml
	if err != nil {
		log.Panicf("Error while initialising config, err: %v", err)
		panic(err)
	}

	ctx, err = config.TODOContext() 			//global context
	if err != nil {
		log.Panicf("Failed to create context: %v", err)
	}
	log.Println("Context initialized successfully!")
}

func GetContext() context.Context {
	return ctx
}
