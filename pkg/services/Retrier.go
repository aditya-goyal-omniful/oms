package services

import (
	"context"
	"time"

	"github.com/aditya-goyal-omniful/oms/pkg/helpers"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
)

func StartOrderRetryWorker() {
	ctx := context.Background()
	go func() {
		ticker := time.NewTicker(2 * time.Minute) 
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				log.Info(i18n.Translate(ctx, "Running retry logic for on_hold orders"))
				processOnHoldOrders()
			}
		}
	}()
}

func processOnHoldOrders() {
	ctx := context.Background()

	orders, err := helpers.GetOnHoldOrders(ctx)
	if err != nil {
		log.Errorf(i18n.Translate(ctx, "Failed to fetch on_hold orders: %v"), err)
		return
	}

	for _, order := range orders {
		log.Infof(i18n.Translate(ctx, "Retrying order: %s"), order.OrderID)

		helpers.CheckAndUpdateOrder(ctx, order)
	}
}
