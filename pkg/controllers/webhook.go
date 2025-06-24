package controllers

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	localContext "github.com/aditya-goyal-omniful/oms/context"
	"github.com/aditya-goyal-omniful/oms/pkg/database"
	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/aditya-goyal-omniful/oms/pkg/services"
)

var ctx context.Context
var collection *mongo.Collection
var err error

type WebhookRegisterRequest struct {
	URL string `json:"url" binding:"required,url"`
}

func init() {
	ctx = localContext.GetContext()
	dbname := config.GetString(ctx, "mongo.dbname")
	collectionName := config.GetString(ctx, "mongo.webhookCollectionName") 

	collection, err = database.GetMongoCollection(dbname, collectionName) 	// Get Mongo Collection
	if err != nil {
		log.Panic(err)
	}
}

func RegisterWebhook(c *gin.Context) {
	var req WebhookRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(int(http.StatusBadRequest), gin.H{i18n.Translate(c, "invalid_request"): err})
		return
	}

	tenantID := c.GetHeader("X-Tenant-ID")
	if tenantID == "" {
		c.JSON(int(http.StatusBadRequest), gin.H{ i18n.Translate(c, "missing_tenant_id"): nil})
		return
	}

	ctx := context.Background()
	webhook := models.Webhook{
		TenantID:  tenantID,
		URL:       req.URL,
		CreatedAt: time.Now(),
	}

	filter := bson.M{"tenant_id": tenantID}
	update := bson.M{"$set": webhook}
	_, err := collection.UpdateOne(ctx, filter, update, mongoOptions(true))
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "Webhook registration failed"))
		c.JSON(int(http.StatusInternalServerError), gin.H{i18n.Translate(c, "internal_error"): err})
		return
	}

	services.CacheWebhookURL(ctx, tenantID, req.URL)
	c.JSON(int(http.StatusOK), gin.H{i18n.Translate(ctx, "message"): i18n.Translate(ctx, "Webhook registered successfully")})
}

func mongoOptions(upsert bool) *options.UpdateOptions {
	opt := options.Update().SetUpsert(upsert)
	return opt
}
