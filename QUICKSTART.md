# Quick Start Guide

## Setup (First Time)

```bash
# 1. Start database
docker-compose up -d

# 2. Wait for database (check health)
docker-compose ps

# 3. Run migrations
DATABASE_PORT=5433 go run cmd/migrate/main.go

# 4. Start API
DATABASE_PORT=5433 go run cmd/api/main.go
```

## Quick Test Commands

### Create Accounts
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": 1000.00}'

curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 2, "initial_balance": 500.00}'
```

### Check Balances
```bash
curl http://localhost:8080/accounts/1
curl http://localhost:8080/accounts/2
```

### Transfer Money
```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: transfer-001" \
  -d '{
    "source_account_id": 1,
    "destination_account_id": 2,
    "amount": 250.50
  }'
```

## Using Makefile

```bash
make setup              # Start DB + run migrations
make run                # Start API server
make test               # Run all tests
make test-unit          # Run unit tests only
make test-integration   # Run integration tests
make test-coverage      # Generate coverage report
make stop               # Stop database
make clean              # Remove all data
```

## Verify Implementation

### 1. Test Cents/Float Conversion
```bash
go test -v ./internal/models/...
```

### 2. Test Validation Logic
```bash
go test -v ./tests/unit/...
```

### 3. Test with Real Database
```bash
make setup
go test -v ./tests/integration/...
```

### 4. Test Concurrency
```bash
go test -v -race ./tests/integration/concurrency_test.go ./tests/integration/api_test.go
```

## Database Access

```bash
docker exec -it financial-ledger-db psql -U ledger_user -d financial_ledger

SELECT * FROM accounts;
SELECT * FROM transactions;
```

## Cleanup

```bash
make stop

make clean
```
