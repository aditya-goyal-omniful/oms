package services

import (
	"context"
	"encoding/json"

	"github.com/aditya-goyal-omniful/oms/pkg/helpers"
	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/omniful/go_commons/i18n"
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

func InitKafkaConsumer(ctx context.Context) {
	log.Infof(i18n.Translate(ctx, "Initializing Kafka consumer..."))

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
	ctx := context.Background()

	defer func() {
		log.Infof(i18n.Translate(ctx, "Closing Kafka consumer"))
		kafkaConsumer.Close()
	}()

	log.Infof(i18n.Translate(ctx, "Attaching NewRelic interceptor to consumer"))
	kafkaConsumer.SetInterceptor(interceptor.NewRelicInterceptor())

	handler := &MessageHandler{}
	topic := "order.created"

	log.Infof(i18n.Translate(ctx, "Registering handler for topic: %s"), topic)
	kafkaConsumer.RegisterHandler(topic, handler)

	log.Infof(i18n.Translate(ctx, "Subscribing to topic: %s"), topic)
	go kafkaConsumer.Subscribe(ctx)

	select {} // Block forever
}

func (h *MessageHandler) Handle(ctx context.Context, msg *pubsub.Message) error {
	var order models.Order
	err := json.Unmarshal(msg.Value, &order)
	if err != nil {
		log.Errorf(i18n.Translate(ctx, "Failed to unmarshal Kafka message: %v"), err)
		return err
	}

	helpers.CheckAndUpdateOrder(ctx, order)

	tenantID := msg.Headers["X-Tenant-ID"]
	if tenantID == "" {
		log.Warnf("TenantID not found in Kafka headers")
	} else {
		NotifyTenantWebhook(ctx, tenantID, order)
	}


	return nil
}