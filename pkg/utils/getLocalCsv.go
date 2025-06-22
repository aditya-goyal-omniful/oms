package utils

import (
	"log"
	"os"
)

func GetLocalCSV(filepath string) ([]byte, error) {
	filePath := filepath
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return fileBytes, nil
}
