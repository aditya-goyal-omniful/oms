package configs

import (
	"github.com/aditya-goyal-omniful/oms/context"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/sqs"
)

var newQueue *sqs.Queue

func SQSInit() {

	ctx := context.GetContext()

	log.Infof(i18n.Translate(ctx, "Initializing SQS Queue"))


	region := config.GetString(ctx, "aws.region")
	account := config.GetString(ctx, "aws.account")
	endpoint := config.GetString(ctx, "aws.sqsendpoint")
	queueName := config.GetString(ctx, "aws.sqsname")

	sqsCfg := sqs.GetSQSConfig(
		ctx,
		false,
		"",
		region,
		account,
		endpoint,
	)

	log.Infof(i18n.Translate(ctx, "SQS Config:"), sqsCfg)

	// Ensure queue exists (create if missing)
	err := sqs.CreateQueue(ctx, sqsCfg, queueName, "standard")
	if err != nil {
		log.Panic(i18n.Translate(ctx, "Error creating SQS queue:"), err)
		return 
	}

	newQueue, err = sqs.NewStandardQueue(ctx, queueName, sqsCfg)

	if err != nil {
		log.Panic(err)
	}

	log.Infof(i18n.Translate(ctx, "Standard SQS Queue Successfully created"))
}

func GetSqs() *sqs.Queue {
	return newQueue
}

