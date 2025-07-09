package helpers

import (
	"testing"
)

func TestEvaluateInventoryResponse(t *testing.T) {
	tests := []struct {
		name     string
		body     []byte
		expected string
	}{
		{"Available True", []byte(`{"available": true}`), "new_order"},
		{"Available False", []byte(`{"available": false}`), "on_hold"},
		{"Invalid JSON", []byte(`not-json`), "error"},
		{"Missing Field", []byte(`{}`), "on_hold"}, // fallback to default false
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EvaluateInventoryResponse(tt.body)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}
