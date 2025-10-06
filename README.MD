# Product API Service

A RESTful Product API service implementing the OpenAPI specification for an e-commerce system. This service provides endpoints for creating and retrieving product information, with data stored in-memory using a thread-safe hashmap.


### Prerequisites
- Go 1.22 or higher
- Docker (for containerization)

### Running Locally

1. **Navigate to the src directory:**
   ```powershell
   cd src
   ```

2. **Download dependencies:**
   ```powershell
   go mod tidy
   ```

3. **Run the server:**
   ```powershell
   go run .
   ```
   
   The server will start on port 8080 (configurable via `PORT` environment variable).

4. **Run tests:**
   ```powershell
   go test ./...
   ```

### Running with Docker

1. **Build the Docker image:**
   ```powershell
   docker build -t product-api:latest .
   ```

2. **Run the container:**
   ```powershell
   docker run -p 8080:8080 product-api:latest
   ```

## 📖 API Documentation

### Endpoints

#### Health Check
```
GET /healthz
```
**Response:** `200 OK`
```json
{"status":"ok"}
```

#### Get Product by ID
```
GET /products/{productId}
```

**Parameters:**
- `productId` (path, required): Integer, minimum 1

**Responses:**
- `200 OK`: Product found
  ```json
  {
    "product_id": 1,
    "sku": "ABC-123-XYZ",
    "manufacturer": "Acme Corp",
    "category_id": 42,
    "weight": 1250,
    "some_other_id": 7
  }
  ```
- `404 Not Found`: Product does not exist
  ```json
  {
    "error": "PRODUCT_NOT_FOUND",
    "message": "The requested product does not exist"
  }
  ```
- `400 Bad Request`: Invalid product ID
- `500 Internal Server Error`: Server error

#### Add/Update Product Details
```
POST /products/{productId}/details
```

**Parameters:**
- `productId` (path, required): Integer, minimum 1

**Request Body:**
```json
{
  "product_id": 1,
  "sku": "ABC-123-XYZ",
  "manufacturer": "Acme Corp",
  "category_id": 42,
  "weight": 1250,
  "some_other_id": 7
}
```

**Field Constraints:**
- `product_id`: Must match the path parameter, positive integer
- `sku`: Required, 1-100 characters
- `manufacturer`: Required, 1-200 characters
- `category_id`: Positive integer
- `weight`: Non-negative integer (grams)
- `some_other_id`: Positive integer

**Responses:**
- `204 No Content`: Product successfully created/updated
- `400 Bad Request`: Invalid input data
  ```json
  {
    "error": "INVALID_INPUT",
    "message": "sku is required"
  }
  ```
- `500 Internal Server Error`: Server error

## 🧪 Testing Examples

### Using curl (PowerShell)

**1. Health check:**
```powershell
curl.exe -i http://localhost:8080/healthz
```

**2. Create a product (using file):**

Create `product.json`:
```powershell
@'
{
  "product_id": 1,
  "sku": "ABC-123-XYZ",
  "manufacturer": "Acme Corp",
  "category_id": 42,
  "weight": 1250,
  "some_other_id": 7
}
'@ > product.json
```

POST the product:
```powershell
curl.exe -i -X POST http://localhost:8080/products/1/details `
  -H "Content-Type: application/json" `
  --data-binary "@product.json"
```

**3. Retrieve the product:**
```powershell
curl.exe -i http://localhost:8080/products/1
```

**4. Test error cases:**

Missing product (404):
```powershell
curl.exe -i http://localhost:8080/products/999
```

Mismatched product_id (400):
```powershell
curl.exe -i -X POST http://localhost:8080/products/2/details `
  -H "Content-Type: application/json" `
  --data-binary "@product.json"
```

Invalid payload - empty sku (400):
```powershell
curl.exe -i -X POST http://localhost:8080/products/3/details `
  -H "Content-Type: application/json" `
  -d '{"product_id":3,"sku":"","manufacturer":"Test","category_id":1,"weight":100,"some_other_id":1}'
```

### Using curl (Bash/Linux)

**Create a product:**
```bash
curl -i -X POST http://localhost:8080/products/1/details \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": 1,
    "sku": "ABC-123-XYZ",
    "manufacturer": "Acme Corp",
    "category_id": 42,
    "weight": 1250,
    "some_other_id": 7
  }'
```

**Retrieve the product:**
```bash
curl -i http://localhost:8080/products/1
```

## 🏗️ Implementation Details

### Data Storage
- **Structure**: Concurrent-safe in-memory hashmap
- **Key**: Product ID (integer)
- **Value**: Product struct (all fields)
- **Concurrency**: Protected by `sync.RWMutex`
- **Persistence**: Data is volatile and cleared on restart

### Architecture
- **Framework**: go-chi/chi v5 for routing
- **Middleware**: Request ID, logging, panic recovery
- **Validation**: Strict input validation per OpenAPI spec
- **Error Handling**: Structured error responses with codes

### Container Image
- **Base**: Multi-stage build with distroless runtime
- **Size**: ~25MB (optimized for production)
- **Security**: Non-root, minimal attack surface
- **Platform**: linux/amd64

## 🌐 AWS Deployment

This service is designed for deployment to AWS ECS via Terraform.

### Environment Variables
- `PORT`: Server listen port (default: 8080)

### Health Endpoint
- Path: `/healthz`
- Use this for ECS health checks and load balancer target group health checks

### Example ECS Task Definition Health Check
```json
{
  "healthCheck": {
    "command": ["CMD-SHELL", "wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1"],
    "interval": 30,
    "timeout": 5,
    "retries": 3,
    "startPeriod": 60
  }
}
```

## 📝 Notes

- All timestamps and logs use UTC
- Product IDs must be positive integers
- The POST endpoint performs **upsert** (create if new, update if exists)
- No authentication/authorization is currently implemented
- Data is stored in-memory only (non-persistent)


## 📄 License

MIT License - Educational project for CS6650

## 🔗 References

- [OpenAPI Specification](https://swagger.io/docs/specification/v3_0/about/)
- [API Contract](../api.yaml)
- [HTTP Status Codes](https://http.cat/)

## Instructions

### Prepare Credentials

Retrieve you temporary credentials from Learner's Lab.
Enter your configuration when prompted:
```
aws configure
```

And set your session token:
```
aws configure set aws_session_token <YOUR-TEMP-SESSION-TOKEBN>
```

### Apply Infrastructure
```
cd terraform
terraform init
terraform apply -auto-approve
```

### Send Requests
Make sure you are in terraform folder

Get public ip address
```
aws ec2 describe-network-interfaces \
--network-interface-ids $(
    aws ecs describe-tasks \
    --cluster $(terraform output -raw ecs_cluster_name) \
    --tasks $(
        aws ecs list-tasks \
        --cluster $(terraform output -raw ecs_cluster_name) \
        --service-name $(terraform output -raw ecs_service_name) \
        --query 'taskArns[0]' --output text
    ) \
    --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" \
    --output text
) \
--query 'NetworkInterfaces[0].Association.PublicIp' \
--output text
```
powershell:
$cluster = terraform output -raw ecs_cluster_name
$service = terraform output -raw ecs_service_name
$taskArn = aws ecs list-tasks --cluster $cluster --service-name $service --query 'taskArns[0]' --output text
$eni = aws ecs describe-tasks --cluster $cluster --tasks $taskArn --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" --output text
aws ec2 describe-network-interfaces --network-interface-ids $eni --query 'NetworkInterfaces[0].Association.PublicIp' --output text

Send some requests:
```
curl http://<PUBLIC-IP-ADDRESS>:8080/healthz
curl -sS -X GET "http://<PUBLIC-IP>:8080/products" -H "Accept: application/json"
curl -sS -X POST "http://<PUBLIC-IP>:8080/products" \
  -H "Content-Type: application/json" \
  -d '{"id":"1","name":"Example Product","description":"demo","price":9.99}'
```

## Clean Up
```
terraform destroy -auto-approve
```

Load Testing:
```
locust -f locustfile.py --host http://<PUBLIC-IP>:8080
```

## Terraform Instructions

### Prepare Credentials

Retrieve you temporary credentials from Learner's Lab.
Enter your configuration when prompted:
```
aws configure
```

And set your session token:
```
aws configure set aws_session_token <YOUR-TEMP-SESSION-TOKEBN>
```

### Apply Infrastructure
```
cd terraform
terraform init
terraform apply -auto-approve
```

### Send Requests
Make sure you are in terraform folder

Get public ip address
```
aws ec2 describe-network-interfaces \
--network-interface-ids $(
    aws ecs describe-tasks \
    --cluster $(terraform output -raw ecs_cluster_name) \
    --tasks $(
        aws ecs list-tasks \
        --cluster $(terraform output -raw ecs_cluster_name) \
        --service-name $(terraform output -raw ecs_service_name) \
        --query 'taskArns[0]' --output text
    ) \
    --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" \
    --output text
) \
--query 'NetworkInterfaces[0].Association.PublicIp' \
--output text
```
powershell:
$cluster = terraform output -raw ecs_cluster_name
$service = terraform output -raw ecs_service_name
$taskArn = aws ecs list-tasks --cluster $cluster --service-name $service --query 'taskArns[0]' --output text
$eni = aws ecs describe-tasks --cluster $cluster --tasks $taskArn --query "tasks[0].attachments[0].details[?name=='networkInterfaceId'].value" --output text
aws ec2 describe-network-interfaces --network-interface-ids $eni --query 'NetworkInterfaces[0].Association.PublicIp' --output text

Send some requests:
```
curl http://<PUBLIC-IP-ADDRESS>:8080/healthz
curl -sS -X GET "http://<PUBLIC-IP>:8080/products" -H "Accept: application/json"
curl -sS -X POST "http://<PUBLIC-IP>:8080/products" \
  -H "Content-Type: application/json" \
  -d '{"id":"1","name":"Example Product","description":"demo","price":9.99}'
```

## Clean Up
```
terraform destroy -auto-approve
```

Load Testing:
```
locust -f locustfile.py --host http://<PUBLIC-IP>:8080
```