package models

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"

	"strings"

	localContext "github.com/aditya-goyal-omniful/oms/context"
	localConfig "github.com/aditya-goyal-omniful/oms/pkg/configs"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/sqs"
	"go.mongodb.org/mongo-driver/mongo"
)

type StoreCSV struct {
	FilePath string `json:"filePath"`
}

type BulkOrderRequest struct {
	FilePath string `json:"filePath"`
}

type Order struct {
	OrderID  int64   `json:"order_id" csv:"order_id"`
	SKUID    int64   `json:"sku_id" csv:"sku_id"`
	Quantity int     `json:"quantity" csv:"quantity"`
	SellerID int64   `json:"seller_id" csv:"seller_id"`
	HubID    int64   `json:"hub_id" csv:"hub_id"`
	Price    float64 `json:"price" csv:"price"`
}

var mongoClinet *mongo.Client
var err error
var client *awsS3.Client //  this is being returned to me by s3.NewDefaultAWSS3Client() of gocommons
var ctx context.Context
var collection *mongo.Collection
var publisher *sqs.Publisher // Publisher for SQS messages

func init() {
	ctx = localContext.GetContext()
	dbname := config.GetString(ctx, "mongo.dbname")
	collectionName := config.GetString(ctx, "mongo.collectionName")

	localConfig.ConnectDB()                                                  // Connect to MongoDB
	collection, err = localConfig.GetMongoCollection(dbname, collectionName) //get the collection
	if err != nil {
		log.Fatal(err)
	}

	localConfig.ConnectS3()            // Initialize S3 client
	client = localConfig.GetS3Client() //get s3 client

	localConfig.SQSInit()            // Initialize SQS client
	newQueue := localConfig.GetSqs() // Get the SQS queue

	localConfig.PublisherInit(newQueue)    // Initialize SQS Publisher
	publisher = localConfig.GetPublisher() // get Publisher

	localConfig.ConsumerInit()     // Initialize SQS Consumer
	localConfig.StartConsumer(ctx) // Start the SQS consumer for processing CSV files
}

func StoreInS3(s *StoreCSV) error {
	filepath := s.FilePath
	fileBytes := localConfig.GetLocalCSV(filepath)

	bucketName := config.GetString(ctx, "s3.bucketName")
	filename := config.GetString(ctx, "s3.fileName")

	// bucketName := os.Getenv("S3_BUCKETNAME")
	// filename := os.Getenv("S3_FILENAME")

	input := &awsS3.PutObjectInput{
		Bucket: &bucketName,
		Key:    &filename,
		Body:   bytes.NewReader(fileBytes),
	}

	_, err := client.PutObject(ctx, input)
	if err != nil {
		//log.Println("here is error1")
		log.Println(err)
		return errors.New("failed to upload to s3")
	}
	log.Println("File uploaded to S3!")
	return nil
}

func ValidateS3Path_PushToSQS(req *BulkOrderRequest) error {
	log.Println("Validating S3 path:")
	filePath := req.FilePath

	if !strings.HasPrefix(filePath, "s3://") {
		return errors.New("invalid S3 path format: must start with s3://")
	}

	path := strings.TrimPrefix(filePath, "s3://")
	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		return errors.New("invalid S3 path: must be in s3://bucket/key format")
	}

	bucket := parts[0]
	key := parts[1]

	log.Println(bucket, key)

	_, err := client.HeadObject(ctx, &awsS3.HeadObjectInput{
		Bucket: &bucket, //bucket name
		Key:    &key,    // file name
	})
	//log.Println(location)

	if err != nil {
		log.Println(err)
		return errors.New("file does not exist at specified S3 path")
	}
	log.Println("S3 path is valid Successfully!")
	log.Println("Pushing to SQS...")

	err = PushToSQS(bucket, key)
	if err != nil {
		log.Println("Failed to push to SQS:", err)
	}
	log.Println("Successfully pushed to SQS!")
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
		log.Println("Failed to publish message to SQS:", err)
		return err
	}
	log.Println("Message successfully published to SQS")
	return nil
}