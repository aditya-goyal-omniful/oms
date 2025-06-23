package helpers

import (
	"context"

	"github.com/aditya-goyal-omniful/oms/pkg/database"
	"github.com/aditya-goyal-omniful/oms/pkg/entities"
	"github.com/google/uuid"
	"github.com/omniful/go_commons/httpclient"
	"github.com/omniful/go_commons/httpclient/request"
	"github.com/omniful/go_commons/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var client httpclient.Client

func InitHTTPClient() {
	client = httpclient.New("http://localhost:8087")
}



func UpdateOrderStatus(orderID uuid.UUID, status string) error {
	ctx := context.TODO()
	collection, err := database.GetMongoCollection("oms", "orders")
	if err != nil {
		return err
	}

	filter := bson.M{"order_id": orderID}
	update := bson.M{"$set": bson.M{"status": status}}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("MongoDB update failed:", err)
	}
	return err
}

func CheckAndUpdateOrder(ctx context.Context, order entities.Order) {
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

	resp, err := client.Send(ctx, req)
	if err != nil {
		log.Printf("HTTP call failed for order %s: %v", order.OrderID, err)
		return
	}

	var result struct {
		Available bool `json:"available"`
	}
	if err := resp.UnmarshalBody(&result); err != nil {
		log.Printf("Failed to unmarshal IMS response for order %s: %v", order.OrderID, err)
		return
	}

	newStatus := "on_hold"
	if result.Available {
		newStatus = "new_order"
	}

	if err := UpdateOrderStatus(uuid.UUID(order.OrderID), newStatus); err != nil {
		log.Infof("Failed to update status for order %s: %v", order.OrderID, err)
	}
}

func GetOnHoldOrders(ctx context.Context) ([]entities.Order, error) {
	var orders []entities.Order

	cursor, err := database.Collection.Find(ctx, bson.M{"status": "on_hold"})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var order entities.Order
		if err := cursor.Decode(&order); err != nil {
			continue
		}
		orders = append(orders, order)
	}

	return orders, nil
}