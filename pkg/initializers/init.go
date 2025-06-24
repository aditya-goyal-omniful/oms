package initializers

import (
	"context"
	"time"

	localConfig "github.com/aditya-goyal-omniful/oms/pkg/configs"
	"github.com/aditya-goyal-omniful/oms/pkg/controllers"
	"github.com/aditya-goyal-omniful/oms/pkg/database"
	"github.com/aditya-goyal-omniful/oms/pkg/entities"
	"github.com/aditya-goyal-omniful/oms/pkg/services"
)

func InitServices(ctx context.Context) {
	database.ConnectDB(ctx) 							// Initialize Mongo Client

	services.InitRedis(ctx)							// Initialize Redis

	localConfig.ConnectS3(ctx) 						// Initialize S3 client

	localConfig.SQSInit(ctx)   						// Initialize SQS client
	newQueue := localConfig.GetSqs() 

	localConfig.PublisherInit(ctx, newQueue)    	// Initialize SQS Publisher
	localConfig.ConsumerInit(ctx)     		   		// Initialize SQS Consumer
	localConfig.StartConsumer(ctx) 					// Start the SQS consumer for processing CSV files

	entities.InitCSV(ctx)							// Initialize Order Mongo Collection

	go services.InitKafkaConsumer(ctx) 				// Initialize Kafka Producer

	time.Sleep(3 * time.Second)						// Sleep to allow consumer to initialize

	services.InitKafkaProducer(ctx)					// Then produce messages
	services.StartOrderRetryWorker()

	controllers.InitWebhook(ctx)					// Initialize Webhook Mongo Collection
}