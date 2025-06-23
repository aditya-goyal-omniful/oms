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
	"github.com/aditya-goyal-omniful/oms/pkg/initializers"
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

var err error
var client *awsS3.Client 			// Returned by s3.NewDefaultAWSS3Client() (go_commons)
var ctx context.Context
var collection *mongo.Collection
var publisher *sqs.Publisher

func init() {
	initializers.InitServices()

	ctx = localContext.GetContext()
	dbname := config.GetString(ctx, "mongo.dbname")
	collectionName := config.GetString(ctx, "mongo.collectionName") 

	collection, err = database.GetMongoCollection(dbname, collectionName) 	// Get Mongo Collection
	if err != nil {
		log.Panic(err)
	}

	client = localConfig.GetS3Client() 			// Get S3 Client

	publisher = localConfig.GetPublisher() 		// Get Publisher
}

func StoreInS3(s *StoreCSV) error {
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
	log.Infof(i18n.Translate(ctx, "Validating S3 path:"))
	filePath := req.FilePath

	if !strings.HasPrefix(filePath, "s3://") {
		return errors.New(i18n.Translate(ctx, "invalid S3 path format: must start with s3://"))
	}

	path := strings.TrimPrefix(filePath, "s3://")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return errors.New(i18n.Translate(ctx, "invalid S3 path: must be in s3://bucket/key format"))
	}

	bucket := parts[0]
	key := parts[1]

	log.Infof(bucket, key)

	_, err := client.HeadObject(ctx, &awsS3.HeadObjectInput{
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

func PushToSQS(bucket string, key string) error {

	payload := fmt.Sprintf(`{"bucket":"%s", "key":"%s"}`, bucket, key)
	msg := &sqs.Message{
		Value: []byte(payload),
	}

	// Publish the message to SQS
	err = publisher.Publish(ctx, msg)
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "Failed to publish message to SQS:"))
		return err
	}
	log.Infof(i18n.Translate(ctx, "Message successfully published to SQS"))
	return nil
}