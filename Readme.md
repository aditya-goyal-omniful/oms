# 📟 Order Management Service (OMS)

Golang-based microservice to manage order intake, processing, storage, and webhook notification using Kafka, S3, SQS, Redis, and MongoDB.

---

## 🤩 Tech Stack

* **Language**: Go
* **Storage**: MongoDB
* **Cache**: Redis
* **Message Queues**: Kafka (orders), SQS (CSV ingestion)
* **File Storage**: S3 (via LocalStack)
* **Config**: go\_commons/config
* **HTTP Client**: go\_commons/httpclient
* **i18n**: go\_commons/i18n for internationalization support
* **Swagger**: Swagger UI via swaggo/gin-swagger

---

## 📂 Features

* Create order (validates SKU & Hub, status set to `on_hold`, pushes to Kafka)
* Bulk order upload via CSV → S3 → SQS → Parse → Validate → Save to MongoDB → Kafka
* Order retry worker that retries on `on_hold` orders
* RESTful APIs with multi-tenancy header support (`X-Tenant-ID`)
* Redis-backed validation caching for SKUs and Hubs
* Webhook registration and triggering on successful order creation
* Swagger docs hosted at `/swagger/index.html`

---

## 🤪 API Endpoints

| Method | Endpoint            | Description                         |
| ------ | ------------------- | ----------------------------------- |
| POST   | `/orders`           | Create single order (Kafka + Redis) |
| POST   | `/orders/bulkorder` | Trigger bulk order from S3 via SQS  |
| POST   | `/s3/filepath`      | Upload local CSV to S3              |
| GET    | `/orders`           | Filter orders by seller, date, etc. |
| POST   | `/webhooks`         | Register a webhook for a tenant     |
| GET    | `/webhooks`         | List all registered webhooks        |

---

## ⚙️ How It Works

### 1. **Upload CSV to S3**

* API: `POST /s3/filepath`
* Reads local CSV and uploads to S3 bucket via LocalStack

### 2. **Push S3 Path to SQS**

* API: `POST /orders/bulkorder`
* Sends S3 path to SQS queue for CSV processing

### 3. **CSV Processor (Consumer)**

* Downloads CSV → parses rows
* Validates fields via IMS
* Saves to MongoDB and pushes to Kafka

### 4. **Kafka Consumer (OMS)**

* Listens to topic `order.created`
* Calls IMS API to check inventory and updates status (`new_order` or keeps `on_hold`)
* If successful, triggers tenant's webhook (if registered)

### 5. **Order Retry Worker**

* Background cron worker retries `on_hold` orders every 2 minutes

---

## 🔔 Webhook Flow

* Tenants can register a webhook URL using `POST /webhooks`
* Upon successful order creation (status changed to `new_order`), OMS sends a POST payload to the tenant's webhook URL
* Webhook URLs are cached in Redis for performance

---

## 🐳 Docker Setup

```bash
docker-compose up -d
```

Services:

* Kafka + Zookeeper
* Redis
* LocalStack (S3, SQS)

---

## 🫣 Create S3 Bucket via LocalStack

```bash
aws --endpoint-url=http://localhost:4566 s3api create-bucket --bucket orders
```

---

## 🚀 Run the OMS Server (with LocalStack & Config Setup)

In PowerShell:

```powershell
$env:LOCAL_SQS_ENDPOINT = "http://localhost:4566"
$env:AWS_ACCESS_KEY_ID = "test"
$env:AWS_SECRET_ACCESS_KEY = "test"
$env:AWS_REGION = "us-east-1"
$env:AWS_S3_ENDPOINT = "http://localhost:4566"
$env:LOCAL_S3_BUCKET_URL = "localhost:4566"
$env:LOCALSTACK_ENDPOINT = "http://localhost:4566"
$env:CONFIG_SOURCE = "local"
go run cmd/main.go
```

---

## 📌 Swagger UI

> View Swagger docs at:

👉 [`http://localhost:8082/swagger/index.html`](http://localhost:8082/swagger/index.html)

---

## 📂 CSV Upload Format

```csv
order_id,sku_id,quantity,seller_id,hub_id,price,tenant_id
<uuid>,<uuid>,<int>,<uuid>,<uuid>,<float>,<uuid>
```

> `order_id` is optional. If not provided, a UUID is generated.

---

## 📁 Invalid Orders

* Invalid CSV rows are logged and saved to the `./public` directory as `invalid_orders_<timestamp>.csv`

---

## 📬 Kafka Topics

* **Producer**: `order.created`
* **Consumer**: Updates order status after IMS inventory check and sends webhooks

---

## 📦 Directory Structure

```
├── cmd/
│   └── main.go
├── context/
├── pkg/
│   ├── controllers/
│   ├── database/
│   ├── entities/
│   ├── helpers/
│   ├── models/
│   ├── routes/
│   └── services/
├── utils/
├── docker-compose.yml
├── README.md
├── localstack/
└── public/
```

---

## 📊 Future Improvements

* Add metrics and Prometheus integration
* Add test coverage
* Add retry tracking in DB

---

## 🧠 Developer Notes

* Redis is used to cache SKU/Hub validation from IMS and webhook URLs
* go\_commons provides reusable HTTP client, SQS, S3, Redis, config, i18n, and logging interfaces
* Swagger comments are generated using `swag init`

---

## 🔗 External Dependencies

* [go\_commons](https://github.com/omniful/go_commons)
* [MongoDB Driver](https://github.com/mongodb/mongo-go-driver)
* [AWS SDK v2 for Go](https://aws.github.io/aws-sdk-go-v2/)
* [LocalStack](https://github.com/localstack/localstack)
* [swaggo/swag](https://github.com/swaggo/swag)