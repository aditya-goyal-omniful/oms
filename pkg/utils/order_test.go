package utils

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

// Test ValidateOrder

var validateWithIMSStub func(ctx context.Context, hubID, skuID uuid.UUID) bool

func init() {
	// Replace the original function with the stub
	ValidateWithIMS = func(ctx context.Context, hubID, skuID uuid.UUID) bool {
		return validateWithIMSStub(ctx, hubID, skuID)
	}
}

func TestValidateOrder(t *testing.T) {
	ctx := context.TODO()
	validUUID := uuid.New()

	tests := []struct {
		name     string
		order    *models.Order
		mockIMS  bool
		wantErr  bool
	}{
		{
			name: "valid order",
			order: &models.Order{
				OrderID:   validUUID,
				SKUID:     validUUID,
				HubID:     validUUID,
				SellerID:  validUUID,
				TenantID:  validUUID,
				Price:     100.0,
				Quantity:  2,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			mockIMS: true,
			wantErr: false,
		},
		{
			name: "invalid SKUID",
			order: &models.Order{
				OrderID:   validUUID,
				SKUID:     uuid.Nil,
				HubID:     validUUID,
				SellerID:  validUUID,
				TenantID:  validUUID,
				Price:     100.0,
				Quantity:  2,
			},
			mockIMS: true,
			wantErr: true,
		},
		{
			name: "invalid price",
			order: &models.Order{
				OrderID:   validUUID,
				SKUID:     validUUID,
				HubID:     validUUID,
				SellerID:  validUUID,
				TenantID:  validUUID,
				Price:     -10,
				Quantity:  2,
			},
			mockIMS: true,
			wantErr: true,
		},
		{
			name: "IMS validation failed",
			order: &models.Order{
				OrderID:   validUUID,
				SKUID:     validUUID,
				HubID:     validUUID,
				SellerID:  validUUID,
				TenantID:  validUUID,
				Price:     100.0,
				Quantity:  2,
			},
			mockIMS: false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validateWithIMSStub = func(ctx context.Context, hubID, skuID uuid.UUID) bool {
				return tt.mockIMS
			}

			err := ValidateOrder(ctx, tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOrder() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// 	Test SaveOrder

type insertOneCollection interface {
	InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error)
}

func saveOrderTest(ctx context.Context, order *models.Order, collection insertOneCollection) error {
	order.Status = "on_hold"
	_, err := collection.InsertOne(ctx, order)
	return err
}

type mockCollection struct {
	shouldFail bool
}

func (m *mockCollection) InsertOne(ctx context.Context, doc interface{}) (*mongo.InsertOneResult, error) {
	if m.shouldFail {
		return nil, errors.New("mock insert failed")
	}
	return &mongo.InsertOneResult{InsertedID: uuid.New()}, nil
}

func TestSaveOrder(t *testing.T) {
	ctx := context.TODO()
	order := &models.Order{
		OrderID:   uuid.New(),
		SKUID:     uuid.New(),
		HubID:     uuid.New(),
		SellerID:  uuid.New(),
		TenantID:  uuid.New(),
		Price:     100.0,
		Quantity:  2,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name      string
		mock      insertOneCollection
		wantError bool
	}{
		{
			name:      "successful insert",
			mock:      &mockCollection{shouldFail: false},
			wantError: false,
		},
		{
			name:      "insert fails",
			mock:      &mockCollection{shouldFail: true},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := saveOrderTest(ctx, order, tt.mock)
			if (err != nil) != tt.wantError {
				t.Errorf("saveOrderTest() error = %v, wantErr = %v", err, tt.wantError)
			}
		})
	}
}

// Test ValidateAndSaveOrder

func  TestValidateAndSaveOrder(t *testing.T) {
	TestValidateOrder(t)
	TestSaveOrder(t)
}