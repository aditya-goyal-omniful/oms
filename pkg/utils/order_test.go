package utils

import (
	"context"
	"testing"
	"time"

	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/google/uuid"
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
