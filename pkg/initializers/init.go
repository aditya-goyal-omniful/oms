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

	localConfig.ConnectS3() 				// Initialize S3 client

	localConfig.SQSInit()   				// Initialize SQS client
	newQueue := localConfig.GetSqs() 

	localConfig.PublisherInit(newQueue)    	// Initialize SQS Publisher
	localConfig.ConsumerInit()     		   	// Initialize SQS Consumer
	localConfig.StartConsumer(ctx) 			// Start the SQS consumer for processing CSV files

	go services.InitKafkaConsumer() 		// Initialize Kafka Producer

	time.Sleep(3 * time.Second)				// Sleep to allow consumer to initialize

	services.InitKafkaProducer()			// Then produce messages
}