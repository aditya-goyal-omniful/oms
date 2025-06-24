package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Webhook struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	TenantID  string             `bson:"tenant_id"`
	URL       string             `bson:"url"`
	CreatedAt time.Time          `bson:"created_at"`
}
