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

	"github.com/aditya-goyal-omniful/oms/pkg/database"
	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/aditya-goyal-omniful/oms/pkg/services"
)

type WebhookRegisterRequest struct {
	URL string `json:"url" binding:"required,url"`
}

type WebhookStore interface {
	UpdateOne(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error)
}

type WebhookCacher interface {
	Cache(ctx context.Context, tenantID, url string)
}

type realWebhookCacher struct{}

func (realWebhookCacher) Cache(ctx context.Context, tenantID, url string) {
	services.CacheWebhookURL(ctx, tenantID, url)
}

var (
	ctx context.Context
	collection *mongo.Collection
	err error
	// Global for injection
	webhookStore  WebhookStore  = collection
	webhookCacher WebhookCacher = realWebhookCacher{}
)


func InitWebhook(ctx context.Context) {
	dbname := config.GetString(ctx, "mongo.dbname")
	collectionName := config.GetString(ctx, "mongo.webhookCollectionName") 

	collection, err = database.GetMongoCollection(dbname, collectionName) 	// Get Mongo Collection
	if err != nil {
		log.Panic(err)
	}
}

// RegisterWebhook godoc
// @Summary Register a webhook
// @Description Save a webhook URL for a tenant
// @Tags Webhook
// @Accept json
// @Produce json
// @Param webhook body models.Webhook true "Webhook Payload"
// @Success 201 {object} models.Webhook
// @Router /webhooks/register [post]
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
	_, err := webhookStore.UpdateOne(ctx, filter, update, mongoOptions(true))
	if err != nil {
		log.WithError(err).Error(i18n.Translate(ctx, "Webhook registration failed"))
		c.JSON(int(http.StatusInternalServerError), gin.H{i18n.Translate(c, "internal_error"): err})
		return
	}

	webhookCacher.Cache(ctx, tenantID, req.URL)
	c.JSON(int(http.StatusOK), gin.H{i18n.Translate(ctx, "message"): i18n.Translate(ctx, "Webhook registered successfully")})
}

func mongoOptions(upsert bool) *options.UpdateOptions {
	opt := options.Update().SetUpsert(upsert)
	return opt
}
