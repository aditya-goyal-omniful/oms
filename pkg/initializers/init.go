package initializers

import (
	"time"

	localContext "github.com/aditya-goyal-omniful/oms/context"
	localConfig "github.com/aditya-goyal-omniful/oms/pkg/configs"
	"github.com/aditya-goyal-omniful/oms/pkg/database"
	"github.com/aditya-goyal-omniful/oms/pkg/services"
)

func InitServices() {
	ctx := localContext.GetContext()

	database.ConnectDB() 

	services.InitRedis(ctx)

	localConfig.ConnectS3(ctx) 				// Initialize S3 client

	localConfig.SQSInit()   				// Initialize SQS client
	newQueue := localConfig.GetSqs() 

	localConfig.PublisherInit(ctx, newQueue)    	// Initialize SQS Publisher
	localConfig.ConsumerInit(ctx)     		   	// Initialize SQS Consumer
	localConfig.StartConsumer(ctx) 			// Start the SQS consumer for processing CSV files

	go services.InitKafkaConsumer(ctx) 		// Initialize Kafka Producer

	time.Sleep(3 * time.Second)				// Sleep to allow consumer to initialize

	services.InitKafkaProducer(ctx)			// Then produce messages
}