package configs

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"

	"github.com/aditya-goyal-omniful/oms/pkg/database"
	parse_csv "github.com/aditya-goyal-omniful/oms/pkg/utils"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/sqs"
)

var consumer *sqs.Consumer
var err error

func ConsumerInit(ctx context.Context) {
	sqsQueue := GetSqs()
	consumer, err = sqs.NewConsumer(
		sqsQueue,
		1,
		1,
		&queueHandler{}, // defined below
		10,
		30,
		false,			 // isAsync
		false,			 // sendBatchMessage
	)

	if err != nil {
		log.Panicf(i18n.Translate(ctx, "Failed to start SQS consumer: %v"), err)
	}

	log.Infof(i18n.Translate(ctx, "SQS consumer initialized"))
}

func StartConsumer(ctx context.Context) {
	consumer.Start(ctx)
	log.Infof(i18n.Translate(ctx, "SQS consumer started"))
}

type queueHandler struct{}

func (h *queueHandler) Process(ctx context.Context, msgs *[]sqs.Message) error {
	if err != nil {
		log.Errorf(i18n.Translate(ctx, "Failed to create S3 client: %v"), err)
		return err
	}
	for _, msg := range *msgs {
		// Parse message payload
		var payload struct {
			Bucket string `json:"bucket"`
			Key    string `json:"key"`
		}
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			log.Errorf(i18n.Translate(ctx, "Invalid message payload: %v"), err)
			continue
		}

		// Download from S3
		_, err := s3Client.GetObject(ctx, &awsS3.GetObjectInput{
			Bucket: aws.String(payload.Bucket),
			Key:    aws.String(payload.Key),
		})
		if err != nil {
			log.Errorf(i18n.Translate(ctx, "Failed to download S3 object: %v"), err)
			continue
		}

		// Download CSV to local temp file
		tmpFile := filepath.Join(os.TempDir(), filepath.Base(payload.Key))
		getObjOutput, err := s3Client.GetObject(ctx, &awsS3.GetObjectInput{
			Bucket: aws.String(payload.Bucket),
			Key:    aws.String(payload.Key),
		})
		if err != nil {
			log.Errorf(i18n.Translate(ctx, "failed to download CSV from S3: %v"), err)
			continue
		}
		log.Infof(i18n.Translate(ctx, "Downloaded CSV from S3"))

		defer getObjOutput.Body.Close()

		outFile, err := os.Create(tmpFile)
		if err != nil {
			log.Errorf(i18n.Translate(ctx, "failed to create temp file: %v"), err)
			continue
		}
		log.Infof(i18n.Translate(ctx, "Created temp file to store downloaded CSV"))

		defer outFile.Close()

		_, err = io.Copy(outFile, getObjOutput.Body)
		if err != nil {
			log.Errorf(i18n.Translate(ctx, "failed to write CSV data to file: %v"), err)
			continue
		}

		log.Infof(i18n.Translate(ctx, "CSV data written to temp file: %s"), tmpFile)
		log.Infof(i18n.Translate(ctx, "Starting to parse CSV file: %s"), tmpFile)

		// Parse the CSV file
		err = parse_csv.ParseCSV(tmpFile, ctx, database.Collection)
		if err != nil {
			log.Errorf(i18n.Translate(ctx, "failed to parse CSV file: %v"), err)
			continue
		}
		log.Infof(i18n.Translate(ctx, "CSV file parsed successfully : %s"), tmpFile)
	}

	return nil
}
