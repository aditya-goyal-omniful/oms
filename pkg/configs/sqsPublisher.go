package configs

import (
	"context"

	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/sqs"
)

var publisher *sqs.Publisher

func PublisherInit(ctx context.Context, newQueue *sqs.Queue) {
	log.Infof(i18n.Translate(ctx, "Initializing SQS Publisher"))
	publisher = sqs.NewPublisher(newQueue)
	log.Println(i18n.Translate(ctx, "SQS Publisher successfully created"))
}

func GetPublisher() *sqs.Publisher {
	return publisher
}
