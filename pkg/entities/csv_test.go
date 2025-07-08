package entities

import (
	"encoding/json"
	"testing"
)

func TestIsValidS3Path(t *testing.T) {
	tests := []struct {
		input       string
		wantBucket  string
		wantKey     string
		expectError bool
	}{
		{"s3://my-bucket/file.csv", "my-bucket", "file.csv", false},
		{"s3://bucket/path/to/file.csv", "bucket", "path/to/file.csv", false},
		{"http://bucket/file.csv", "", "", true},
		{"s3://onlybucket", "", "", true},
		{"", "", "", true},
	}

	for _, tt := range tests {
		bucket, key, err := IsValidS3Path(tt.input)
		if (err != nil) != tt.expectError {
			t.Errorf("IsValidS3Path(%q) error = %v, wantErr = %v", tt.input, err, tt.expectError)
			continue
		}
		if bucket != tt.wantBucket || key != tt.wantKey {
			t.Errorf("Expected bucket=%s, key=%s but got bucket=%s, key=%s", tt.wantBucket, tt.wantKey, bucket, key)
		}
	}
}

func TestBuildSQSMessage(t *testing.T) {
	bucket := "test-bucket"
	key := "orders/orders.csv"

	msg := BuildSQSMessage(bucket, key)

	// Parse back the message to check format
	var payload map[string]string
	err := json.Unmarshal(msg.Value, &payload)
	if err != nil {
		t.Fatalf("Failed to unmarshal SQS message payload: %v", err)
	}

	if payload["bucket"] != bucket {
		t.Errorf("Expected bucket %s, got %s", bucket, payload["bucket"])
	}
	if payload["key"] != key {
		t.Errorf("Expected key %s, got %s", key, payload["key"])
	}
}