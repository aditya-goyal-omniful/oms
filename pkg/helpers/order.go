package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/aditya-goyal-omniful/oms/pkg/database"
	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/google/uuid"
	"github.com/omniful/go_commons/httpclient"
	"github.com/omniful/go_commons/httpclient/request"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var client httpclient.Client

func InitHTTPClient() {
	client = httpclient.New("http://localhost:8087")
}

func SendInventoryCheckRequest(ctx context.Context, order models.Order, httpClient httpclient.Client) ([]byte, error) {
	payload := map[string]interface{}{
		"sku_id":   order.SKUID,
		"hub_id":   order.HubID,
		"quantity": order.Quantity,
	}

	req, _ := request.NewBuilder().
		SetUri("/inventory/check-and-update").
		SetMethod("POST").
		SetBody(payload).
		Build()

	resp, err := httpClient.Send(ctx, req)
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "HTTP call failed for order %s:"), order.OrderID)
		return nil, err
	}

	return resp.Body(), nil
}

func EvaluateInventoryResponse(body []byte) string {
	var result struct {
		Available bool `json:"available"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "error"
	}

	if result.Available {
		return "new_order"
	}
	return "on_hold"
}

func CheckOrder(ctx context.Context, order models.Order, httpClient httpclient.Client) string {
	body, err := SendInventoryCheckRequest(ctx, order, httpClient)
	if err != nil {
		return "error"
	}
	return EvaluateInventoryResponse(body)
}

func UpdateOrderStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	collection, err := database.GetMongoCollection("oms", "orders")
	if err != nil {
		return err
	}

	filter := bson.M{"order_id": orderID}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "MongoDB update failed:"))
	}
	return err
}

func CheckAndUpdateOrder(ctx context.Context, order models.Order) {
	newStatus := CheckOrder(ctx, order, client)
	if newStatus == "error" {
		return
	}

	if err := UpdateOrderStatus(ctx, uuid.UUID(order.OrderID), newStatus); err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "Failed to update status for order %s:"), order.OrderID)
	}
}

func GetOnHoldOrders(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order

	cursor, err := database.Collection.Find(ctx, bson.M{"status": "on_hold"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var order models.Order
		if err := cursor.Decode(&order); err != nil {
			continue
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func FetchOrders(ctx context.Context, sellerID uuid.UUID, status string, startDate, endDate time.Time) ([]models.Order, error) {
	filter := bson.M{}

	if sellerID != uuid.Nil {
		filter["seller_id"] = sellerID
	}
	if status != "" {
		filter["status"] = status
	}
	if !startDate.IsZero() || !endDate.IsZero() {
		dateRange := bson.M{}
		if !startDate.IsZero() {
			dateRange["$gte"] = startDate
		}
		if !endDate.IsZero() {
			dateRange["$lte"] = endDate
		}
		filter["created_at"] = dateRange
	}

	collection, err := database.GetMongoCollection("oms", "orders")
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "MongoDB collection error:"))
		return nil, err
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "MongoDB query error:"))
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	for cursor.Next(ctx) {
		var o models.Order
		if err := cursor.Decode(&o); err != nil {
			log.Warnf(i18n.Translate(ctx, "Failed to decode order: %v"), err)
			continue
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func ValidateSKUAndHubs(ctx context.Context, skuID, hubID, tenantID uuid.UUID) (bool, error) {
	// Set up headers with tenant ID
	headers := url.Values{}
	headers.Set("X-Tenant-ID", tenantID.String())

	// Validate SKU
	skuReq, err := request.NewBuilder().
		SetUri(fmt.Sprintf("/skus/%s", skuID)).
		SetMethod("GET").
		SetHeaders(headers).
		Build()
	if err != nil {
		log.Warnf(i18n.Translate(ctx, "Failed to build SKU request: %v"), err)
		return false, err
	}

	skuResp, err := client.Send(ctx, skuReq)
	if err != nil || skuResp.StatusCode() != 200 {
		log.Warnf(i18n.Translate(ctx, "SKU validation failed: %v"), err)
		return false, nil
	}

	// Validate Hub
	hubReq, err := request.NewBuilder().
		SetUri(fmt.Sprintf("/hubs/%s", hubID)).
		SetMethod("GET").
		SetHeaders(headers).
		Build()
	if err != nil {
		log.Warnf(i18n.Translate(ctx, "Failed to build Hub request: %v"), err)
		return false, err
	}

	hubResp, err := client.Send(ctx, hubReq)
	if err != nil || hubResp.StatusCode() != 200 {
		log.Warnf(i18n.Translate(ctx, "Hub validation failed: %v"), err)
		return false, nil
	}

	return true, nil
}