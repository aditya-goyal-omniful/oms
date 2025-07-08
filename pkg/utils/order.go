package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	nethttp "net/http"

	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/google/uuid"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type ValidationResponse struct {
	IsValid bool   `json:"is_valid"`
	Error   string `json:"error"`
}

var client *http.Client
var err error

func InitHTTPClient(ctx context.Context) {
	transport := &nethttp.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}

	serviceName := config.GetString(ctx, "client.serviceName")
	baseURL := config.GetString(ctx, "client.baseURL")
	timeout := config.GetDuration(ctx, "http.timeout")
	client, err = http.NewHTTPClient(
		serviceName,
		baseURL,
		transport,
		http.WithTimeout(timeout),
	)
}


var ValidateWithIMS = func(ctx context.Context, hubID, skuID uuid.UUID) bool {
	req := &http.Request{
		Url: fmt.Sprintf("validators/validate_order/%s/%s", hubID, skuID),
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Timeout: 5 * time.Second,
	}

	var response ValidationResponse
	_, err := client.Get(req, &response)
	if err != nil {
		log.Errorf(i18n.Translate(ctx, "Failed to call IMS validate API: %v"), err)
		return false
	}

	return response.IsValid
}


func ValidateOrder(ctx context.Context, order *models.Order) error {
	if order.OrderID == uuid.Nil {
		return errors.New("invalid OrderID")
	}
	if order.SKUID == uuid.Nil {
		return errors.New("invalid SKUID")
	}
	if order.HubID == uuid.Nil {
		return errors.New("invalid HubID")
	}
	if order.SellerID == uuid.Nil {
		return errors.New("invalid SellerID")
	}
	if order.TenantID == uuid.Nil {
		return errors.New("invalid TenantID")
	}
	if order.Quantity <= 0 {
		return errors.New("invalid Quantity")
	}
	if order.Price < 0 {
		return errors.New("invalid Price")
	}

	valid := ValidateWithIMS(ctx, order.HubID, order.SKUID)
	if !valid {
		return errors.New(i18n.Translate(ctx, "invalid HubID or SKUID"))
	}

	return nil
}


func saveOrder(ctx context.Context, order *models.Order, collection *mongo.Collection) error {
	log.Infof(i18n.Translate(ctx, "Attempting to insert order into DB: %+v"), order)
	order.Status = "on_hold"
	_, err := collection.InsertOne(ctx, order)
	if err != nil {
		log.Errorf(i18n.Translate(ctx, "Mongo insert error: %v"), err)
		return fmt.Errorf(i18n.Translate(ctx, "failed to insert order: %w"), err)
	}
	log.Infof(i18n.Translate(ctx, "Order successfully inserted: %v"), order.OrderID)
	return nil
}


func validateAndSaveOrder(ctx context.Context, order *models.Order, collection *mongo.Collection) error {
	if err := ValidateOrder(ctx, order); err != nil {
		return err
	}
	if err := saveOrder(ctx, order, collection); err != nil {
		return err
	}
	return nil
}