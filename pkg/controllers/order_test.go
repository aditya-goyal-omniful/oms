package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aditya-goyal-omniful/oms/pkg/helpers"
	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/aditya-goyal-omniful/oms/pkg/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockFetcher struct{}

func (m mockFetcher) FetchOrders(ctx context.Context, sellerID uuid.UUID, status string, start, end time.Time) ([]models.Order, error) {
	return []models.Order{{OrderID: uuid.New()}}, nil
}

type failingFetcher struct{}

func (f failingFetcher) FetchOrders(ctx context.Context, sellerID uuid.UUID, status string, start, end time.Time) ([]models.Order, error) {
	return nil, errors.New("mock failure")
}

func TestGetOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	validTenantID := uuid.New().String()
	validSellerID := uuid.New().String()

	tests := []struct {
		name           string
		headers        map[string]string
		query          string
		mockFetcher    helpers.OrderFetcher
		expectedStatus int
	}{
		{
			name:           "Missing Tenant Header",
			headers:        map[string]string{},
			query:          "",
			mockFetcher:    mockFetcher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Tenant ID",
			headers:        map[string]string{"X-Tenant-ID": "not-a-uuid"},
			query:          "",
			mockFetcher:    mockFetcher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Seller ID",
			headers:        map[string]string{"X-Tenant-ID": validTenantID},
			query:          "?seller_id=not-a-uuid",
			mockFetcher:    mockFetcher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid Start Date",
			headers:        map[string]string{"X-Tenant-ID": validTenantID},
			query:          "?start_date=2025-99-99",
			mockFetcher:    mockFetcher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Fetcher Fails",
			headers:        map[string]string{"X-Tenant-ID": validTenantID},
			query:          "?seller_id=" + validSellerID,
			mockFetcher:    failingFetcher{},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "Success",
			headers:        map[string]string{"X-Tenant-ID": validTenantID},
			query:          "?seller_id=" + validSellerID + "&status=on_hold&start_date=2025-01-01&end_date=2025-12-31",
			mockFetcher:    mockFetcher{},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			OrderFetcher = tc.mockFetcher

			router := gin.Default()
			router.GET("/orders", GetOrders)

			req, _ := http.NewRequest(http.MethodGet, "/orders"+tc.query, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("[%s] Expected status %d, got %d", tc.name, tc.expectedStatus, w.Code)
			}
		})
	}
}

type mockValidator struct {
	isValid bool
}

func (m mockValidator) Validate(ctx context.Context, skuID, hubID, tenantID uuid.UUID) (bool, error) {
	return m.isValid, nil
}

type mockPublisher struct{}

func (m *mockPublisher) Publish(order *models.Order, tenantID string) {}

func TestCreateOrder(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type args struct {
		body    map[string]interface{}
		headers map[string]string
	}

	tests := []struct {
		name           string
		args           args
		mockValidator  helpers.SKUValidator
		mockPublisher  services.OrderPublisher
		expectedStatus int
	}{
		{
			name: "Success",
			args: args{
				body: map[string]interface{}{
					"sku_id": uuid.New().String(),
					"hub_id": uuid.New().String(),
				},
				headers: map[string]string{
					"X-Tenant-ID": uuid.New().String(),
				},
			},
			mockValidator: mockValidator{isValid: true},
			mockPublisher: &mockPublisher{},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid JSON",
			args: args{
				body: map[string]interface{}{}, // missing required fields
				headers: map[string]string{
					"X-Tenant-ID": uuid.New().String(),
				},
			},
			mockValidator: mockValidator{isValid: true},
			mockPublisher: &mockPublisher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid Tenant ID",
			args: args{
				body: map[string]interface{}{
					"sku_id": uuid.New().String(),
					"hub_id": uuid.New().String(),
				},
				headers: map[string]string{
					"X-Tenant-ID": "not-a-uuid",
				},
			},
			mockValidator: mockValidator{isValid: true},
			mockPublisher: &mockPublisher{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid SKU/Hub",
			args: args{
				body: map[string]interface{}{
					"sku_id": uuid.New().String(),
					"hub_id": uuid.New().String(),
				},
				headers: map[string]string{
					"X-Tenant-ID": uuid.New().String(),
				},
			},
			mockValidator: mockValidator{isValid: false},
			mockPublisher: &mockPublisher{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			SKUValidator = tc.mockValidator
			OrderPublisher = tc.mockPublisher

			router := gin.Default()
			router.POST("/orders", CreateOrder)

			body, _ := json.Marshal(tc.args.body)
			req, _ := http.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			for k, v := range tc.args.headers {
				req.Header.Set(k, v)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("[%s] Expected status %d but got %d", tc.name, tc.expectedStatus, w.Code)
			}
		})
	}
}
