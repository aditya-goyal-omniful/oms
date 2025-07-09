package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mockWebhookStore struct {
	shouldFail bool
}

func (m mockWebhookStore) UpdateOne(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
	if m.shouldFail {
		return nil, errors.New("mock mongo failure")
	}
	return &mongo.UpdateResult{}, nil
}

type mockWebhookCacher struct {
	called bool
}

func (m *mockWebhookCacher) Cache(ctx context.Context, tenantID, url string) {
	m.called = true
}

func TestRegisterWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           map[string]interface{}
		headers        map[string]string
		store          WebhookStore
		cacher         *mockWebhookCacher
		expectedStatus int
	}{
		{
			name:           "Missing Tenant Header",
			body:           map[string]interface{}{"url": "https://example.com"},
			headers:        map[string]string{},
			store:          mockWebhookStore{},
			cacher:         &mockWebhookCacher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON Body",
			body:           map[string]interface{}{}, // url is missing
			headers:        map[string]string{"X-Tenant-ID": uuid.New().String()},
			store:          mockWebhookStore{},
			cacher:         &mockWebhookCacher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Mongo Update Fails",
			body:           map[string]interface{}{"url": "https://example.com"},
			headers:        map[string]string{"X-Tenant-ID": uuid.New().String()},
			store:          mockWebhookStore{shouldFail: true},
			cacher:         &mockWebhookCacher{},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Success",
			body:           map[string]interface{}{"url": "https://example.com"},
			headers:        map[string]string{"X-Tenant-ID": uuid.New().String()},
			store:          mockWebhookStore{},
			cacher:         &mockWebhookCacher{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Inject mocks
			webhookStore = tc.store
			webhookCacher = tc.cacher

			router := gin.Default()
			router.POST("/webhooks/register", RegisterWebhook)

			bodyBytes, _ := json.Marshal(tc.body)
			req, _ := http.NewRequest(http.MethodPost, "/webhooks/register", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("[%s] expected status %d, got %d", tc.name, tc.expectedStatus, w.Code)
			}
		})
	}
}
