package configs

import (
	"log"

	"github.com/omniful/go_commons/sqs"
)

var publisher *sqs.Publisher

func PublisherInit(newQueue *sqs.Queue) {
	log.Println("Initializing SQS Publisher")
	publisher = sqs.NewPublisher(newQueue)
	log.Println("SQS Publisher successfully created")
}

func GetPublisher() *sqs.Publisher {
	return publisher
}
