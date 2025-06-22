package controllers

import (
	"bytes"
	"io"
	"log"

	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/gin-gonic/gin"
)

func ServeHome(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "OMS Service",
	})
}

func StoreInS3(c *gin.Context) {
	var req = &models.StoreCSV{}

	body, _ := c.GetRawData() // 👈 get raw body
	log.Println("Raw Body:", string(body))

	// 👇 Put body back for binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	if err := c.ShouldBindJSON(req); err != nil {
		log.Println("Bind Error:", err) // 👈 log bind error
		c.JSON(400, gin.H{
			"error": "Failed Parse request",
		})
		return
	}

	log.Println("Parsed filePath:", req.FilePath)

	err := models.StoreInS3(req)
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


func CreateBulkOrder(c *gin.Context) {
	var req = &models.BulkOrderRequest{}
	err := c.ShouldBindBodyWithJSON(&req)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Failed Parse request",
		})
		return
	}
	err = models.ValidateS3Path_PushToSQS(req)

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
