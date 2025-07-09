package controllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aditya-goyal-omniful/oms/pkg/entities"

	"github.com/gin-gonic/gin"
)

type mockUploader struct{}

func (m *mockUploader) Store(req *entities.StoreCSV) error {
	return nil
}

type failingUploader struct{}

func (f *failingUploader) Store(req *entities.StoreCSV) error {
	return errors.New("mock failure")
}

func TestStoreInS3(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type testCase struct {
		name           string
		body           map[string]interface{}
		mockUploader   entities.S3Uploader
		expectedStatus int
	}

	tests := []testCase{
		{
			name: "Successful Upload",
			body: map[string]interface{}{
				"filePath": "test-folder/sample.csv",
			},
			mockUploader: &mockUploader{}, // returns nil
			expectedStatus: http.StatusOK,
		},
		{
			name: "Uploader Fails",
			body: map[string]interface{}{
				"filePath": "test-folder/sample.csv",
			},
			mockUploader: &failingUploader{}, // returns error
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			body:           map[string]interface{}{}, // missing filePath
			mockUploader:   &mockUploader{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Uploader = tc.mockUploader

			router := gin.Default()
			router.POST("/upload", StoreInS3)

			jsonBody, err := json.Marshal(tc.body)
			if err != nil {
				t.Fatalf("failed to marshal request body: %v", err)
			}

			req, err := http.NewRequest(http.MethodPost, "/upload", bytes.NewBuffer(jsonBody))
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("[%s] Expected status %d but got %d", tc.name, tc.expectedStatus, w.Code)
			}
		})
	}
}


type mockPusher struct{}

func (m *mockPusher) Push(req *entities.BulkOrderRequest) error {
	return nil
}

type failingPusher struct{}

func (f *failingPusher) Push(req *entities.BulkOrderRequest) error {
	return errors.New("mock push failure")
}

func TestCreateBulkOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           map[string]interface{}
		mockPusher     entities.SQSPusher
		expectedStatus int
	}{
		{
			name: "Success",
			body: map[string]interface{}{
				"filePath": "test-folder/sample.csv",
			},
			mockPusher:     &mockPusher{},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid JSON",
			body: map[string]interface{}{}, // missing filePath
			mockPusher:     &mockPusher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Push fails",
			body: map[string]interface{}{
				"filePath": "test-folder/sample.csv",
			},
			mockPusher:     &failingPusher{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			Pusher = tc.mockPusher

			router := gin.Default()
			router.POST("/orders/bulk", CreateBulkOrder)

			payload, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest(http.MethodPost, "/orders/bulk", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("[%s] Expected status %d but got %d", tc.name, tc.expectedStatus, w.Code)
			}
		})
	}
}