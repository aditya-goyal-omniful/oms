package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/kafka"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/pubsub"
)

var kafkaProducer *kafka.ProducerClient

type OrderPublisher interface {
	Publish(order *models.Order, tenantID string)
}

type RealPublisher struct{}

func (RealPublisher) Publish(order *models.Order, tenantID string) {
	PublishOrder(order, tenantID)
}

func InitKafkaProducer(ctx context.Context) {
	log.Infof(i18n.Translate(ctx, "Initializing Kafka producer"))

	kafkaProducer = kafka.NewProducer(
		kafka.WithBrokers([]string{"localhost:9092"}),
		kafka.WithClientID("my-producer"),
		kafka.WithKafkaVersion("3.4.0"),
	)
}

func GetKafkaProducer() *kafka.ProducerClient {
	return kafkaProducer
}

func CloseKafkaProducer(ctx context.Context) {
	if kafkaProducer != nil {
		log.Infof(i18n.Translate(ctx, "Closing Kafka producer"))
		kafkaProducer.Close()
	}
}

func PublishOrder(order *models.Order, tenantID string) {
	ctx := context.WithValue(context.Background(), "request_id", fmt.Sprintf("req-%s", order.OrderID))

	// Marshal order into JSON
	jsonBytes, err := json.Marshal(order)
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "Failed to marshal order:"))
		return
	}

	msg := &pubsub.Message{
		Topic: "order.created",
		Key:   fmt.Sprintf("order-%s", order.OrderID),
		Value: jsonBytes,
		Headers: map[string]string{
			"source": "order-service",
			"X-Tenant-ID": tenantID,
		},
	}

	log.Infof("Publishing order to topic: %s", msg.Topic)
	err = kafkaProducer.Publish(ctx, msg)
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "Failed to publish order:"))
		panic(err)
	}
	log.Infof(i18n.Translate(ctx, "Order published to Kafka successfully: OrderID=%s"), order.OrderID)
}