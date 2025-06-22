package configs

import (
	"log"

	"github.com/Abhishek-Omniful/OMS/mycontext"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/sqs"
)

var newQueue *sqs.Queue

func SQSInit() {

	log.Println("Initializing SQS Queue")

	ctx := mycontext.GetContext()

	region := config.GetString(ctx, "aws.region")
	account := config.GetString(ctx, "aws.account")
	endpoint := config.GetString(ctx, "aws.sqsendpoint")
	queueName := config.GetString(ctx, "aws.sqsname")

	//log.Println("Region:", region, "Account:", account, "Endpoint:", endpoint)

	sqsCfg := sqs.GetSQSConfig(
		ctx,
		false,
		"",
		region,
		account,
		endpoint,
	)

	log.Println("SQS Config:", sqsCfg)

	// Ensure queue exists (create if missing)
	err := sqs.CreateQueue(ctx, sqsCfg, queueName, "standard")
	if err != nil {
		log.Fatal("Error creating SQS queue:", err)
		return 
	}

	newQueue, err = sqs.NewStandardQueue(ctx, queueName, sqsCfg)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Standard SQS Queue Successfully created")
}

func GetSqs() *sqs.Queue {
	return newQueue
}

