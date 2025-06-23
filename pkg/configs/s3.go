package configs

import (
	"context"

	"github.com/aditya-goyal-omniful/oms/pkg/utils"
	awsS3 "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/s3"
)

var s3Client *awsS3.Client

func ConnectS3(c context.Context) {
	//s3client creation

	log.Infof(i18n.Translate(c, "Connecting to s3"))
	s3Client, err = s3.NewDefaultAWSS3Client()
	if err != nil {
		log.Panicf(i18n.Translate(c, "Error connecting to s3:"), err)
		return
	}
	log.Infof(i18n.Translate(c, "Successfully Connected to s3"))
}

func GetS3Client() *awsS3.Client {
	return s3Client
}

func GetLocalCSV(filepath string) []byte {
	fileBytes, _ := utils.GetLocalCSV(filepath)
	return fileBytes
}
