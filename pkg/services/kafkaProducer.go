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

func InitKafkaProducer() {
	log.Println("Initializing Kafka producer")

	kafkaProducer = kafka.NewProducer(
		kafka.WithBrokers([]string{"localhost:9092"}),
		kafka.WithClientID("my-producer"),
		kafka.WithKafkaVersion("3.4.0"),
	)
}

func GetKafkaProducer() *kafka.ProducerClient {
	return kafkaProducer
}

func CloseKafkaProducer() {
	if kafkaProducer != nil {
		log.Println("Closing Kafka producer")
		kafkaProducer.Close()
	}
}

func PublishOrder(order *models.Order) {
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