package utils

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/aditya-goyal-omniful/oms/pkg/services"
	"github.com/google/uuid"
	"github.com/omniful/go_commons/csv"
	"github.com/omniful/go_commons/i18n"
	"github.com/omniful/go_commons/log"
	"go.mongodb.org/mongo-driver/mongo"
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

func extractOrderFromRow(row []string, colIdx map[string]int) (*models.Order, error) {
	orderID, err := uuid.Parse(row[colIdx["order_id"]])
	if err != nil {
		return nil, err
	}
	skuID, err := uuid.Parse(row[colIdx["sku_id"]])
	if err != nil {
		return nil, err
	}
	hubID, err := uuid.Parse(row[colIdx["hub_id"]])
	if err != nil {
		return nil, err
	}
	sellerID, err := uuid.Parse(row[colIdx["seller_id"]])
	if err != nil {
		return nil, err
	}
	tenantID, err := uuid.Parse(row[colIdx["tenant_id"]])
	if err != nil {
		return nil, err
	}
	price, err := strconv.ParseFloat(row[colIdx["price"]], 64)
	if err != nil {
		return nil, err
	}
	quantity, err := strconv.Atoi(row[colIdx["quantity"]])
	if err != nil {
		return nil, err
	}

	order := &models.Order{
		OrderID:   orderID,
		SKUID:     skuID,
		HubID:     hubID,
		SellerID:  sellerID,
		TenantID:  tenantID,
		Price:     price,
		Quantity:  quantity,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return order, nil
}

func writeInvalidCSV(ctx context.Context, headers []string, invalid csv.Records) error {
	timestamp := time.Now().Format("20060102_150405")
	filePath := fmt.Sprintf("public/invalid_orders_%s.csv", timestamp)

	dest := &csv.Destination{}
	dest.SetFileName(filePath)
	dest.SetUploadDirectory("public/")
	dest.SetRandomizedFileName(false)

	writer, err := csv.NewCommonCSVWriter(
		csv.WithWriterHeaders(headers),
		csv.WithWriterDestination(*dest),
	)
	if err != nil {
		return err
	}
	defer writer.Close(ctx)

	if err := writer.Initialize(); err != nil {
		return err
	}
	if err := writer.WriteNextBatch(invalid); err != nil {
		return err
	}

	log.Infof(i18n.Translate(ctx, "Invalid rows saved to CSV at: %s"), filePath)
	publicURL := fmt.Sprintf("http://localhost:8082/%s", filePath)
	log.Infof(i18n.Translate(ctx, "Download invalid CSV here: %s"), publicURL)
	return nil
}

func ParseCSV(tmpFile string, ctx context.Context, collection *mongo.Collection) error {
	csvReader, err := csv.NewCommonCSV(
		csv.WithBatchSize(100),
		csv.WithSource(csv.Local),
		csv.WithLocalFileInfo(tmpFile),
		csv.WithHeaderSanitizers(csv.SanitizeAsterisks, csv.SanitizeToLower),
		csv.WithDataRowSanitizers(csv.SanitizeSpace, csv.SanitizeToLower),
	)
	if err != nil {
		log.Errorf(i18n.Translate(ctx, "failed to create CSV reader: %v"), err)
		return err
	}

	if err := csvReader.InitializeReader(ctx); err != nil {
		log.Errorf(i18n.Translate(ctx, "failed to initialize CSV reader: %v"), err)
		return err
	}

	headers, err := csvReader.GetHeaders()
	if err != nil {
		log.Errorf(i18n.Translate(ctx, "failed to read CSV headers: %v"), err)
		return err
	}
	log.Infof(i18n.Translate(ctx, "CSV Headers: %v"), headers)

	colIdx := make(map[string]int)
	for i, col := range headers {
		colIdx[col] = i
	}

	var invalid csv.Records

	for !csvReader.IsEOF() {
		records, err := csvReader.ReadNextBatch()
		if err != nil {
			log.Errorf(i18n.Translate(ctx, "failed to read CSV batch: %v"), err)
			break
		}

		for _, row := range records {
			log.Infof(i18n.Translate(ctx, "CSV Row: %v"), row)

			order, err := extractOrderFromRow(row, colIdx)
			if err != nil {
				log.Warnf(i18n.Translate(ctx, "Failed to parse order row: %v"), err)
				invalid = append(invalid, row)
				continue
			}

			if err := validateAndSaveOrder(ctx, order, collection); err != nil {
				log.Warnf(i18n.Translate(ctx, "Validation or save failed: %v"), err)
				invalid = append(invalid, row)
				continue
			}

			services.PublishOrder(order, order.TenantID.String())
		}
	}

	if len(invalid) > 0 {
		if err := writeInvalidCSV(ctx, headers, invalid); err != nil {
			log.Errorf(i18n.Translate(ctx, "Error writing invalid CSV: %v"), err)
			return err
		}
	}

	return nil
}