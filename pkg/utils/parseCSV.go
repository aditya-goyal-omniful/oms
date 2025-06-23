package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	nethttp "net/http"

	localContext "github.com/aditya-goyal-omniful/oms/context"
	"github.com/aditya-goyal-omniful/oms/pkg/models"
	"github.com/aditya-goyal-omniful/oms/pkg/services"
	"github.com/google/uuid"
	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/csv"
	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/log"
	"go.mongodb.org/mongo-driver/mongo"
)

type ValidationResponse struct {
	IsValid bool   `json:"is_valid"`
	Error   string `json:"error"`
}

var client *http.Client
var err error

func init() {
	transport := &nethttp.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}

	ctx := localContext.GetContext()

	serviceName := config.GetString(ctx, "client.serviceName")
	baseURL := config.GetString(ctx, "client.baseURL")
	timeout := config.GetDuration(ctx, "http.timeout")
	client, err = http.NewHTTPClient(
		serviceName,
		baseURL,
		transport,
		http.WithTimeout(timeout),
	)
}

func ValidateWithIMS(hubID, skuID uuid.UUID) bool {
	req := &http.Request{
		Url: fmt.Sprintf("validators/validate_order/%s/%s", hubID, skuID),
		Headers: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Timeout: 5 * time.Second,
	}

	var response ValidationResponse
	_, err := client.Get(req, &response)
	if err != nil {
		log.Errorf("Failed to call IMS validate API: %v", err)
		return false
	}

	return response.IsValid
}

func ValidateOrder(order *models.Order) error {
	if order.OrderID == uuid.Nil {
		return errors.New("invalid OrderID")
	}
	if order.SKUID == uuid.Nil {
		return errors.New("invalid SKUID")
	}
	if order.HubID == uuid.Nil {
		return errors.New("invalid HubID")
	}
	if order.SellerID == uuid.Nil {
		return errors.New("invalid SellerID")
	}
	if order.Quantity <= 0 {
		return errors.New("invalid Quantity")
	}
	if order.Price < 0 {
		return errors.New("invalid Price")
	}

	valid := ValidateWithIMS(order.HubID, order.SKUID)
	if !valid {
		return errors.New("invalid HubID or SKUID")
	}

	return nil
}

func saveOrder(ctx context.Context, order *models.Order, collection *mongo.Collection) error {
	log.Infof("Attempting to insert order into DB: %+v", order)
	order.Status = "on_hold"
	_, err := collection.InsertOne(ctx, order)
	if err != nil {
		log.Errorf("Mongo insert error: %v", err)
		return fmt.Errorf("failed to insert order: %w", err)
	}
	log.Infof("Order successfully inserted: %v", order.OrderID)
	return nil
}

func ParseCSV(tmpFile string, ctx context.Context, logger *log.Logger, collection *mongo.Collection) error {
	csvReader, err := csv.NewCommonCSV(
		csv.WithBatchSize(100),
		csv.WithSource(csv.Local),
		csv.WithLocalFileInfo(tmpFile),
		csv.WithHeaderSanitizers(csv.SanitizeAsterisks, csv.SanitizeToLower),
		csv.WithDataRowSanitizers(csv.SanitizeSpace, csv.SanitizeToLower),
	)
	if err != nil {
		logger.Errorf("failed to create CSV reader: %v", err)
		return err
	}

	if err := csvReader.InitializeReader(ctx); err != nil {
		logger.Errorf("failed to initialize CSV reader: %v", err)
		return err
	}

	headers, err := csvReader.GetHeaders()
	if err != nil {
		logger.Errorf("failed to read CSV headers: %v", err)
		return err
	}
	logger.Infof("CSV Headers: %v", headers)

	colIdx := make(map[string]int)
	for i, col := range headers {
		colIdx[col] = i
	}

	var invalid csv.Records

	for !csvReader.IsEOF() {
		records, err := csvReader.ReadNextBatch()
		if err != nil {
			logger.Errorf("failed to read CSV batch: %v", err)
			break
		}

		for _, row := range records {
			logger.Infof("CSV Row: %v", row)

			orderID, _ := uuid.Parse(row[colIdx["order_id"]])
			skuID, _ := uuid.Parse(row[colIdx["sku_id"]])
			hubID, _ := uuid.Parse(row[colIdx["hub_id"]])
			sellerID, _ := uuid.Parse(row[colIdx["seller_id"]])
			price, _ := strconv.ParseFloat(row[colIdx["price"]], 64)
			quantity, _ := strconv.Atoi(row[colIdx["quantity"]])

			order := models.Order{
				OrderID:  orderID,
				SKUID:    skuID,
				HubID:    hubID,
				SellerID: sellerID,
				Price:    price,
				Quantity: quantity,
			}

			if err := ValidateOrder(&order); err != nil {
				logger.Warnf("Validation failed: %v", err)
				invalid = append(invalid, row)
				continue
			}

			if err := saveOrder(ctx, &order, collection); err != nil {
				logger.Errorf("Save failed: %v", err)
				invalid = append(invalid, row)
				continue
			}

			services.PublishOrder(&order)
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
			logger.Errorf("failed to create CSV writer: %v", err)
			return err
		}
		defer writer.Close(ctx)

		if err := writer.Initialize(); err != nil {
			logger.Errorf("failed to initialize CSV writer: %v", err)
			return err
		}

		if err := writer.WriteNextBatch(invalid); err != nil {
			logger.Errorf("failed to write invalid rows: %v", err)
			return err
		}

		logger.Infof("Invalid rows saved to CSV at: %s", filePath)

		publicURL := fmt.Sprintf("http://localhost:8082/%s", filePath)

		logger.Infof("Download invalid CSV here: %s", publicURL)
	}
	return nil
}
