package controllers

import (
	"bytes"
	"io"
	"log"

	"github.com/aditya-goyal-omniful/oms/pkg/entities"
	"github.com/gin-gonic/gin"
)

// StoreInS3 godoc
// @Summary Upload file path to S3 (via localstack)
// @Description Accepts file path in JSON and uploads the file to S3
// @Tags Orders
// @Accept json
// @Produce json
// @Param input body entities.StoreCSV true "File Path Request"
// @Success 200 {object} map[string]string "message: File uploaded to S3!"
// @Failure 400 {object} map[string]string "error: Failed Parse request or upload failure"
// @Router /s3/filepath [post]
func StoreInS3(c *gin.Context) {
	var req = &entities.StoreCSV{}

	body, _ := c.GetRawData()
	log.Println("Raw Body:", string(body))

	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	if err := c.ShouldBindJSON(req); err != nil {
		log.Println("Bind Error:", err)
		c.JSON(400, gin.H{
			"error": "Failed Parse request",
		})
		return
	}

	log.Println("Parsed filePath:", req.FilePath)

	err := entities.StoreInS3(req)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed to upload to s3",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "File uploaded to S3!",
	})
}

// CreateBulkOrder godoc
// @Summary Trigger bulk order creation via S3
// @Description Validates S3 path and pushes message to SQS for processing CSV orders
// @Tags Orders
// @Accept json
// @Produce json
// @Param input body entities.BulkOrderRequest true "S3 Path to CSV File"
// @Success 200 {object} map[string]string "message: Valid Path to s3 !"
// @Failure 400 {object} map[string]string "error: Invalid path or S3 bucket missing"
// @Router /orders/bulkorder [post]
func CreateBulkOrder(c *gin.Context) {
	var req = &entities.BulkOrderRequest{}
	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed Parse request",
		})
		return
	}
	err = entities.ValidateAndPushToSQS(req)

	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid  path to s3 or s3 bucket dont exits, first try creatring one and retry",
		})
		return
	}

	log.Println("Valid Path to s3 !")
	log.Println("Pushing to sqs !")
	c.JSON(200, gin.H{
		"message": "Valid Path to s3 !",
	})
}
