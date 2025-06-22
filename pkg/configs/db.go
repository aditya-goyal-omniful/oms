package configs

import (
	"log"

	"github.com/Abhishek-Omniful/OMS/mycontext"
	"github.com/omniful/go_commons/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var mongoClient *mongo.Client
var err error
var collection *mongo.Collection

func ConnectDB() {
	ctx := mycontext.GetContext()
	log.Println("Connecting to MongoDB...")
	mongoURI := config.GetString(ctx, "mongo.uri")
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
		return
	}
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("Connected to MongoDB successfully")
}

func GetDB() *mongo.Client {
	return mongoClient
}

func GetMongoCollection(dbname string, collectionName string) (*mongo.Collection, error) {
	mongoClient := GetDB()
	collection = mongoClient.Database(dbname).Collection(collectionName)
	log.Printf("Connected to MongoDB collection: %s in database: %s", collectionName, dbname)
	return collection, err
}
