package services

import (
	"context"
	"time"

	"github.com/omniful/go_commons/httpclient"
	"github.com/omniful/go_commons/httpclient/request"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/redis"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/aditya-goyal-omniful/oms/pkg/models"
)

var (
	WebhookCollection *mongo.Collection
	RedisClient       *redis.Client
)

func InitRedis(ctx context.Context) {
	log.Infof(i18n.Translate(ctx, "Initializing Redis..."))
	config := &redis.Config{
		Hosts:        []string{"localhost:6379"},
		PoolSize:     50,
		MinIdleConn:  10,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		IdleTimeout:  10 * time.Minute,
	}
	RedisClient = redis.NewClient(config)
	log.Infof(i18n.Translate(ctx, "Redis initialized successfully!"))
}

func CacheWebhookURL(ctx context.Context, tenantID, url string) {
	_, err := RedisClient.Set(ctx, "webhook:"+tenantID, url, 0)
	if err != nil {
		log.Error(i18n.Translate(ctx, "Failed to cache webhook URL"), err)
	}
}

func GetCachedWebhookURL(ctx context.Context, tenantID string) string {
	val, err := RedisClient.Get(ctx, "webhook:"+tenantID)
	if err != nil {
		log.Warn(i18n.Translate(ctx, "Failed to get webhook URL from cache"), err)
		return ""
	}
	return val
}

func NotifyTenantWebhook(ctx context.Context, tenantID string, payload interface{}) {
	log.Infof(i18n.Translate(ctx, "Preparing to notify tenant webhook for TenantID=%s"), tenantID)

	urlStr := GetCachedWebhookURL(ctx, tenantID)
	if urlStr == "" {
		var wh models.Webhook
		err := WebhookCollection.FindOne(ctx, bson.M{"tenant_id": tenantID}).Decode(&wh)
		if err != nil {
			log.Warn(i18n.Translate(ctx, "No webhook found for tenant"), tenantID)
			return
		}
		urlStr = wh.URL
		CacheWebhookURL(ctx, tenantID, urlStr)
	}

	client := httpclient.New("")

	req, err := request.NewBuilder().
		SetMethod("POST").
		SetUri(urlStr).
		SetHeaders(map[string][]string{
			"Content-Type": {"application/json"},
		}).
		SetBody(payload).
		Build()
	if err != nil {
		log.Error(i18n.Translate(ctx, "Failed to build webhook request"), err)
		return
	}

	log.Infof(i18n.Translate(ctx, "Sending webhook to TenantID=%s at URL=%s"), tenantID, urlStr)

	_, err = client.Post(ctx, req)
	if err != nil {
		log.Error(i18n.Translate(ctx, "Failed to send webhook to ") + urlStr, err)
	}

	log.Infof(i18n.Translate(ctx, "Successfully sent webhook for TenantID=%s to URL=%s"), tenantID, urlStr)
}