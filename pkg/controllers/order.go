package controllers

import (
	"net/http"
	"time"

	"github.com/aditya-goyal-omniful/oms/pkg/helpers"
	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/aditya-goyal-omniful/oms/pkg/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omniful/go_commons/log"
)

// CreateOrder godoc
// @Summary Create a new order (async via Kafka)
// @Description Accepts an order payload, validates SKU and Hub with IMS, sets status to `on_hold`, and publishes to Kafka for further processing.
// @Tags Orders
// @Accept json
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param order body models.Order true "Order payload (OrderID optional; generated if missing)"
// @Success 202 {object} map[string]interface{} "Accepted with order_id and status"
// @Failure 400 {object} map[string]string "Invalid input or missing fields"
// @Failure 500 {object} map[string]string "Internal server error while publishing"
// @Router /orders [post]
func CreateOrder(c *gin.Context) {
	var order models.Order

	if err := c.ShouldBindJSON(&order); err != nil {
		log.Errorf("Invalid JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	tenantIDStr := c.GetHeader("X-Tenant-ID")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}


	// Validate SKU and Hub via Redis + IMS
	isValid, err := helpers.ValidateSKUAndHubs(c.Request.Context(), order.SKUID, order.HubID, tenantID)
	if err != nil || !isValid {
		log.Warnf("Invalid SKU or Hub: sku_id=%s, hub_id=%s", order.SKUID, order.HubID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SKU ID or Hub ID"})
		return
	}

	order.Status = "on_hold"

	if order.OrderID == uuid.Nil {
		order.OrderID = uuid.New()
	}


	// Push to Kafka
	services.PublishOrder(&order)

	c.JSON(http.StatusAccepted, gin.H{
		"message":  "Order queued for processing",
		"order_id": order.OrderID,
		"status":   order.Status,
	})
}

// GetOrders godoc
// @Summary List orders with filters
// @Description Returns all orders for a tenant with optional filters: seller_id, status, and created date range.
// @Tags Orders
// @Produce json
// @Param X-Tenant-ID header string true "Tenant ID"
// @Param seller_id query string false "UUID of the seller"
// @Param status query string false "Order status filter (e.g., new_order, on_hold)"
// @Param start_date query string false "Filter orders created after this date (YYYY-MM-DD)"
// @Param end_date query string false "Filter orders created before this date (YYYY-MM-DD)"
// @Success 200 {array} models.Order "Filtered list of orders"
// @Failure 400 {object} map[string]string "Invalid query or header values"
// @Failure 500 {object} map[string]string "Failed to retrieve orders"
// @Router /orders [get]
func GetOrders(c *gin.Context) {
	tenantIDStr := c.GetHeader("X-Tenant-ID")
	_, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid X-Tenant-ID"})
		return
	}

	sellerIDStr := c.Query("seller_id")
	var sellerID uuid.UUID
	if sellerIDStr != "" {
		sellerID, err = uuid.Parse(sellerIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid seller_id"})
			return
		}
	}

	status := c.Query("status")

	var startDate, endDate time.Time
	if s := c.Query("start_date"); s != "" {
		startDate, err = time.Parse("2006-01-02", s)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date"})
			return
		}
	}
	if e := c.Query("end_date"); e != "" {
		endDate, err = time.Parse("2006-01-02", e)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date"})
			return
		}
	}

	orders, err := helpers.FetchOrders(c.Request.Context(), sellerID, status, startDate, endDate)
	if err != nil {
		log.Errorf("Failed to fetch orders: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}
