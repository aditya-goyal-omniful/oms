package entities

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"strings"

	localContext "github.com/aditya-goyal-omniful/oms/context"
	localConfig "github.com/aditya-goyal-omniful/oms/pkg/configs"
	"github.com/aditya-goyal-omniful/oms/pkg/database"
	"go.mongodb.org/mongo-driver/mongo"

	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/sqs"
)

type StoreCSV struct {
	FilePath string `json:"filePath"`
}

type BulkOrderRequest struct {
	FilePath string `json:"filePath"`
}

var (
	err error
	client *awsS3.Client 			// Returned by s3.NewDefaultAWSS3Client() (go_commons)
	ctx context.Context
	collection *mongo.Collection
	publisher *sqs.Publisher
)

func InitCSV(ctx context.Context) {
	dbname := config.GetString(ctx, "mongo.dbname")
	collectionName := config.GetString(ctx, "mongo.collectionName") 

	collection, err = database.GetMongoCollection(dbname, collectionName) 	// Get Mongo Collection
	if err != nil {
		log.Panic(err)
	}

	client = localConfig.GetS3Client() 			// Get S3 Client

	publisher = localConfig.GetPublisher() 		// Get Publisher
}

func IsValidS3Path(path string) (bucket, key string, err error) {
	if !strings.HasPrefix(path, "s3://") {
		return "", "", errors.New("invalid S3 path format: must start with s3://")
	}

	path = strings.TrimPrefix(path, "s3://")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid S3 path: must be in s3://bucket/key format")
	}

	return parts[0], parts[1], nil
}


func StoreInS3(s *StoreCSV) error {
	ctx := localContext.GetContext()
	filepath := s.FilePath
	fileBytes := localConfig.GetLocalCSV(filepath)

	bucketName := config.GetString(ctx, "s3.bucketName")
	filename := config.GetString(ctx, "s3.fileName")

	input := &awsS3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &filename,
		Body:   bytes.NewReader(fileBytes),
	}

	_, err := client.PutObject(ctx, input)
	if err != nil {
		log.Error(err)
		return errors.New(i18n.Translate(ctx, "failed to upload to s3"))
	}
	log.Infof(i18n.Translate(ctx, "File uploaded to S3!"))
	return nil
}

func ValidateAndPushToSQS(req *BulkOrderRequest) error {
	ctx := localContext.GetContext()
	log.Infof(i18n.Translate(ctx, "Validating S3 path:"))
	filePath := req.FilePath

	bucket, key, err := IsValidS3Path(filePath)
	if err != nil {
		log.Error(err)
	}

	log.Infof(bucket, key)

	_, err = client.HeadObject(ctx, &awsS3.HeadObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})

	if err != nil {
		log.Error(err)
		return errors.New(i18n.Translate(ctx, "file does not exist at specified S3 path"))
	}
	log.Infof(i18n.Translate(ctx, "S3 path is valid Successfully!"))
	log.Infof(i18n.Translate(ctx, "Pushing to SQS..."))

	err = PushToSQS(bucket, key)
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "Failed to push to SQS:"))
	}
	log.Infof(i18n.Translate(ctx, "Successfully pushed to SQS!"))
	return nil
}

func BuildSQSMessage(bucket string, key string) *sqs.Message {
	payload := fmt.Sprintf(`{"bucket":"%s", "key":"%s"}`, bucket, key)
	return &sqs.Message{
		Value: []byte(payload),
	}
}

func PushToSQS(bucket string, key string) error {
	ctx := localContext.GetContext()
	msg := BuildSQSMessage(bucket, key)

	// Publish the message to SQS
	err = publisher.Publish(ctx, msg)
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "Failed to publish message to SQS:"))
		return err
	}
	log.Infof(i18n.Translate(ctx, "Message successfully published to SQS"))
	return nil
}