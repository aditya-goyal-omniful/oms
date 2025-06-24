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

			orderID, _ := uuid.Parse(row[colIdx["order_id"]])
			skuID, _ := uuid.Parse(row[colIdx["sku_id"]])
			hubID, _ := uuid.Parse(row[colIdx["hub_id"]])
			sellerID, _ := uuid.Parse(row[colIdx["seller_id"]])
			tenantID, _ := uuid.Parse(row[colIdx["tenant_id"]])
			price, _ := strconv.ParseFloat(row[colIdx["price"]], 64)
			quantity, _ := strconv.Atoi(row[colIdx["quantity"]])

			order := models.Order{
				OrderID:  orderID,
				SKUID:    skuID,
				HubID:    hubID,
				SellerID: sellerID,
				TenantID: tenantID,
				Price:    price,
				Quantity: quantity,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			if err := ValidateOrder(ctx, &order); err != nil {
				log.Warnf(i18n.Translate(ctx, "Validation failed: %v"), err)
				invalid = append(invalid, row)
				continue
			}

			if err := saveOrder(ctx, &order, collection); err != nil {
				log.Errorf(i18n.Translate(ctx, "Save failed: %v"), err)
				invalid = append(invalid, row)
				continue
			}

			services.PublishOrder(&order, tenantID.String())
		}
	}

	if len(invalid) > 0 {
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
			log.Errorf(i18n.Translate(ctx, "failed to create CSV writer: %v"), err)
			return err
		}
		defer writer.Close(ctx)

		if err := writer.Initialize(); err != nil {
			log.Errorf(i18n.Translate(ctx, "failed to initialize CSV writer: %v"), err)
			return err
		}

		if err := writer.WriteNextBatch(invalid); err != nil {
			log.Errorf(i18n.Translate(ctx, "failed to write invalid rows: %v"), err)
			return err
		}

		log.Infof(i18n.Translate(ctx, "Invalid rows saved to CSV at: %s"), filePath)

		publicURL := fmt.Sprintf("http://localhost:8082/%s", filePath)

		log.Infof(i18n.Translate(ctx, "Download invalid CSV here: %s"), publicURL)
	}
	return nil
}