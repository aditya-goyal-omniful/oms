package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestExtractOrderFromRow(t *testing.T) {
	validUUID := uuid.New().String()

	colIdx := map[string]int{
		"order_id":  0,
		"sku_id":    1,
		"hub_id":    2,
		"seller_id": 3,
		"tenant_id": 4,
		"price":     5,
		"quantity":  6,
	}

	tests := []struct {
		name    string
		row     []string
		wantErr bool
	}{
		{
			name:    "Valid row",
			row:     []string{validUUID, validUUID, validUUID, validUUID, validUUID, "99.99", "5"},
			wantErr: false,
		},
		{
			name:    "Invalid UUID in order_id",
			row:     []string{"invalid-uuid", validUUID, validUUID, validUUID, validUUID, "99.99", "5"},
			wantErr: true,
		},
		{
			name:    "Invalid price",
			row:     []string{validUUID, validUUID, validUUID, validUUID, validUUID, "abc", "5"},
			wantErr: true,
		},
		{
			name:    "Invalid quantity",
			row:     []string{validUUID, validUUID, validUUID, validUUID, validUUID, "99.99", "five"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := extractOrderFromRow(tt.row, colIdx)
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}