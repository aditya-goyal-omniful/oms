package controllers

import (
	"bytes"
	"io"
	"log"

	"github.com/aditya-goyal-omniful/oms/pkg/entities"
	"github.com/gin-gonic/gin"
)

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
