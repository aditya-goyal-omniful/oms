package services

import (
	"context"
	"log"

	"github.com/omniful/go_commons/kafka"
	"github.com/omniful/go_commons/pubsub"
	"github.com/omniful/go_commons/pubsub/interceptor"
)

var kafkaConsumer *kafka.ConsumerClient

// Implement message handler
type MessageHandler struct{}

func (h *MessageHandler) Handle(ctx context.Context, msg *pubsub.Message) error {
	log.Printf("Handling message from topic: %s, key: %s, value: %s", msg.Topic, msg.Key, string(msg.Value))
	// Add processing logic here
	return nil
}

// Implement the required Process method for IPubSubMessageHandler interface
func (h *MessageHandler) Process(ctx context.Context, msg *pubsub.Message) error {
	return h.Handle(ctx, msg)
}

func InitKafkaConsumer() {
	log.Println("Initializing Kafka consumer...")

	kafkaConsumer = kafka.NewConsumer(
		kafka.WithBrokers([]string{"localhost:9092"}), // or "shared-kafka:9092" inside Docker
		kafka.WithConsumerGroup("my-consumer-group"),
		kafka.WithClientID("my-consumer"),
		kafka.WithKafkaVersion("3.4.0"),
	)
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
	topic := "my-topic"

	log.Printf("Registering handler for topic: %s", topic)
	kafkaConsumer.RegisterHandler(topic, handler)

	log.Printf("Subscribing to topic: %s", topic)
	ctx := context.Background()
	go kafkaConsumer.Subscribe(ctx) // running as a goroutine

	// BLOCK forever so consumer can keep running
	select {}
}