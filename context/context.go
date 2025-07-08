package context

import (
	"context"
	"time"

	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
)

var ctx context.Context

func InitAppContext() {
	err := config.Init(time.Second * 10)
	if err != nil {
		log.Panicf("Error while initialising config, err: %v", err)
		panic(err)
	}

	ctx, err = config.TODOContext()
	if err != nil {
		log.Panicf("Failed to create context: %v", err)
	}
	log.Println(i18n.Translate(ctx, "Context initialized successfully!"))
}

func GetContext() context.Context {
	return ctx
}
