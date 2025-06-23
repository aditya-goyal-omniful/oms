package utils

import (
	"os"

	"github.com/omniful/go_commons/log"
)

func GetLocalCSV(filepath string) ([]byte, error) {
	filePath := filepath
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Panic(err)
		return nil, err
	}
	return fileBytes, nil
}
