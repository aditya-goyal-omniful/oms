package database

import (
	"context"

	localContext "github.com/aditya-goyal-omniful/oms/context"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	mongoClient *mongo.Client
	err error
	Collection *mongo.Collection
)

func ConnectDB(ctx context.Context) {
	log.Infof(i18n.Translate(ctx, "Connecting to MongoDB..."))

	mongoURI := config.GetString(ctx, "mongo.uri")
	mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Panic(err)
		return
	}
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Panic(err)
		return
	}
	log.Infof(i18n.Translate(ctx, "Connected to MongoDB successfully"))
}

func GetDB() *mongo.Client {
	return mongoClient
}

func GetMongoCollection(dbname string, collectionName string) (*mongo.Collection, error) {
	ctx := localContext.GetContext()
	mongoClient := GetDB()

	Collection = mongoClient.Database(dbname).Collection(collectionName)
	
	log.Infof(i18n.Translate(ctx, "Connected to MongoDB collection: %s in database: %s"), collectionName, dbname)
	return Collection, err
}
