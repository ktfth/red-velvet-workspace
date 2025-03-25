 # Red Velvet Workspace

This project is a digital banking application that uses Kafka for event streaming and PostgreSQL for data persistence. The application supports checking accounts, credit cards, and PIX transfers.

## Features

- Checking Account Management
  - Create accounts
  - Deposit and withdraw funds
  - Check balance
- Credit Card Operations
  - Issue new credit cards
  - Make purchases
  - Process payments
- PIX Transfers
  - Register PIX keys
  - Send and receive instant transfers
- Transaction History
  - Track all account movements
  - Filter by transaction type

## Prerequisites

- Docker and Docker Compose
- Kubernetes cluster (Docker Desktop Kubernetes, Minikube, or Kind)
- Tilt
- Go 1.21 or higher
- k6 (for load testing)

## Running with Docker Compose

1. Start all services:
```bash
docker-compose up -d
```

2. Check the services:
- Banking API: http://localhost:8080
- Kafka UI: http://localhost:8090
- PostgreSQL: localhost:5432
- pgAdmin: http://localhost:5050

3. Stop all services:
```bash
docker-compose down
```

## Database Management

### PostgreSQL Configuration

The application uses PostgreSQL with the following default settings:
- Database: banco_digital
- User: admin
- Password: admin123
- Port: 5432

### pgAdmin Access

The project includes pgAdmin 4 for database management through a web interface.

1. Access pgAdmin at http://localhost:5050
2. Login credentials:
   - Email: admin@admin.com
   - Password: admin123

3. To connect to the PostgreSQL database:
   1. Click "Add New Server"
   2. In "General" tab:
      - Name: Banco Digital (or any name you prefer)
   3. In "Connection" tab:
      - Host: postgres
      - Port: 5432
      - Database: banco_digital
      - Username: admin
      - Password: admin123

Available features in pgAdmin:
- Database structure visualization
- SQL query execution
- Table management
- Backup and restore
- Performance monitoring

## Running with Kubernetes (Tilt)

1. Make sure your Kubernetes cluster is running:
```bash
kubectl cluster-info
```

2. Start the application with Tilt:
```bash
tilt up
```

3. Access the Tilt UI at http://localhost:10350

4. To stop the application:
```bash
tilt down
```

## API Endpoints

### Account Operations
- POST /accounts - Create a new account
- POST /accounts/credit-cards - Issue a new credit card
- POST /accounts/pix-keys - Register a PIX key
- POST /accounts/transactions - Make a transaction

## Load Testing

We use k6 for load testing. The test scripts are in the `tests/k6` directory.

1. Install k6:
```bash
# Windows (Chocolatey)
choco install k6

# macOS
brew install k6
```

2. Run the load tests:
```bash
k6 run tests/k6/load-test.js
```

The load test simulates:
- Account creation
- Credit card issuance
- PIX key registration
- Deposits
- Credit card purchases
- PIX transfers

## API Testing with Postman

We use Postman to test the API endpoints.

```bash
postman login --with-api-key $POSTMAN_API_KEY
postman collection run 6394192-bd1fe6cf-24a1-4514-b211-c52c62ba9c9e
```

## Project Structure

```
.
├── cmd/
│   └── main.go                # Application entry point
├── internal/
│   ├── domain/
│   │   └── models/           # Domain models
│   ├── application/
│   │   └── services/         # Business logic
│   ├── infrastructure/
│   │   └── database/         # Database configuration
│   └── interfaces/
│       └── http/
│           └── handlers/     # HTTP handlers
├── k8s/                      # Kubernetes manifests
├── tests/
│   └── k6/                   # Load test scripts
├── docker-compose.yml        # Docker Compose configuration
├── Dockerfile               # Docker build configuration
├── Tiltfile                # Tilt configuration
└── README.md               # This file
```

## API Documentation

The API documentation is available at http://localhost:8080/swagger/index.html when the application is running.

## Monitoring

- Kafka UI: http://localhost:8090 (when running with Docker Compose)
- pgAdmin: http://localhost:5050 (when running with Docker Compose)
- Tilt UI: http://localhost:10350 (when running with Tilt)