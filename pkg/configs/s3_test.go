package configs

import (
	"os"
	"testing"
)

func TestGetLocalCSV(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "testfile-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := []byte("order_id,sku_id,hub_id,quantity,price\n123,456,789,10,100.0\n")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	result := GetLocalCSV(tmpFile.Name())

	if string(result) != string(content) {
		t.Errorf("GetLocalCSV returned unexpected content. Got: %s, Want: %s", string(result), string(content))
	}
}
