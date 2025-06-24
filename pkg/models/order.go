package models

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID  uuid.UUID `json:"order_id" csv:"order_id" bson:"order_id"`
	SKUID    uuid.UUID `json:"sku_id" csv:"sku_id" bson:"sku_id"`
	HubID    uuid.UUID `json:"hub_id" csv:"hub_id" bson:"hub_id"`
	SellerID uuid.UUID `json:"seller_id" csv:"seller_id" bson:"seller_id"`
	TenantID uuid.UUID    `json:"tenant_id" bson:"tenant_id"`
	Quantity int       `json:"quantity" csv:"quantity" bson:"quantity"`
	Price    float64   `json:"price" csv:"price" bson:"price"`
	Status   string    `json:"status" csv:"status" bson:"status"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}