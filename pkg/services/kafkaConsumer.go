package services

import (
	"context"
	"encoding/json"

	"github.com/aditya-goyal-omniful/oms/pkg/helpers"
	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/omniful/go_commons/kafka"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/pubsub"
	"github.com/omniful/go_commons/pubsub/interceptor"
)

var kafkaConsumer *kafka.ConsumerClient

// Implement message handler
type MessageHandler struct{}

// Required Process method
func (h *MessageHandler) Process(ctx context.Context, msg *pubsub.Message) error {
	return h.Handle(ctx, msg)
}

func InitKafkaConsumer() {
	log.Println("Initializing Kafka consumer...")

	kafkaConsumer = kafka.NewConsumer(
		kafka.WithBrokers([]string{"localhost:9092"}),
		kafka.WithConsumerGroup("my-consumer-group"),
		kafka.WithClientID("my-consumer"),
		kafka.WithKafkaVersion("3.4.0"),
	)

	helpers.InitHTTPClient()

	ReceiveOrder()
}

func GetKafkaConsumer() *kafka.ConsumerClient {
	return kafkaConsumer
}

func ReceiveOrder() {
	defer func() {
		log.Println("Closing Kafka consumer")
		kafkaConsumer.Close()
	}()

	log.Println("Attaching NewRelic interceptor to consumer")
	kafkaConsumer.SetInterceptor(interceptor.NewRelicInterceptor())

	handler := &MessageHandler{}
	topic := "order.created"

	log.Printf("Registering handler for topic: %s", topic)
	kafkaConsumer.RegisterHandler(topic, handler)

	log.Printf("Subscribing to topic: %s", topic)
	ctx := context.Background()
	go kafkaConsumer.Subscribe(ctx)

	select {} // Block forever
}

func (h *MessageHandler) Handle(ctx context.Context, msg *pubsub.Message) error {
	var order models.Order
	err := json.Unmarshal(msg.Value, &order)
	if err != nil {
		log.Errorf("Failed to unmarshal Kafka message: %v", err)
		return err
	}

	helpers.CheckAndUpdateOrder(ctx, order)
	return nil
}