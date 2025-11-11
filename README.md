# Financial Ledger System

A high-performance internal transfers system built with Go and PostgreSQL. This system provides three core endpoints for managing accounts and money transfers with strong consistency guarantees, idempotency support, and precise financial calculations.

## Key Features

- ✅ **Precise Money Handling**: Stores amounts as cents (integers) to avoid floating-point precision errors
- ✅ **Strong Consistency**: Database transactions with row-level locking prevent race conditions
- ✅ **Idempotency**: Optional idempotency keys prevent duplicate transfers
- ✅ **High Performance**: Connection pooling and optimized queries
- ✅ **Comprehensive Testing**: Unit, integration, and concurrency tests included

## Architecture Highlights

### Cents-Based Storage
- **Database**: Amounts stored as `BIGINT` (cents) - e.g., 10050 cents
- **API**: Amounts exposed as `float64` - e.g., 100.50
- **Why**: Avoids floating-point precision issues (`0.1 + 0.2 ≠ 0.3` in float arithmetic)

### Layered Design
```
HTTP Handler → Service Layer → Repository Layer → PostgreSQL
```

## Prerequisites

- **Go** 1.21 or higher
- **Docker** and **Docker Compose**
- **Make** (optional, for convenience commands)

## Quick Start

### 1. Clone and Setup

```bash
cd financial-ledger-project
cp .env.example .env
```

### 2. Start PostgreSQL Database

```bash
docker-compose up -d
```

Wait for the database to be ready:
```bash
docker-compose ps
# Wait until postgres shows "healthy"
```

### 3. Run Database Migrations

```bash
go run cmd/migrate/main.go
```

You should see:
```
Connected to database successfully
Running migration: 001_accounts.sql
✓ Migration 001_accounts.sql completed
Running migration: 002_transactions.sql
✓ Migration 002_transactions.sql completed
All migrations completed successfully!
```

### 4. Start the API Server

```bash
go run cmd/api/main.go
```

You should see:
```
Connected to database successfully
Starting API server on port 8080...
```

The API is now running at `http://localhost:8080`

## API Endpoints

### 1. Create Account

Creates a new account with an initial balance.

```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "account_id": 1,
    "initial_balance": 1000.50
  }'
```

**Response**: `201 Created` (empty body)

**Error Codes**:
- `400` - Invalid input (negative balance, invalid JSON)
- `409` - Account already exists

### 2. Get Account Balance

Retrieves the current balance for an account.

```bash
curl http://localhost:8080/accounts/1
```

**Response**: `200 OK`
```json
{
  "account_id": 1,
  "balance": 1000.50
}
```

**Error Codes**:
- `404` - Account not found

### 3. Transfer Money

Transfers money between two accounts.

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: transfer-2024-001" \
  -d '{
    "source_account_id": 1,
    "destination_account_id": 2,
    "amount": 250.25
  }'
```

**Response**: `201 Created`
```json
{
  "transaction_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "source_account_id": 1,
  "destination_account_id": 2,
  "amount": 250.25,
  "status": "completed",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Codes**:
- `400` - Invalid input (same account, negative amount, invalid JSON)
- `404` - Account not found
- `409` - Duplicate idempotency key
- `422` - Insufficient funds

### Idempotency

The `Idempotency-Key` header is **optional** but recommended for production use. If you send the same key multiple times, only the first request will execute - subsequent requests return the original transaction without modifying balances.

```bash
# First request - executes transfer
curl -X POST http://localhost:8080/transactions \
  -H "Idempotency-Key: unique-key-123" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": 100.00}'

# Second request - returns same transaction, no balance change
curl -X POST http://localhost:8080/transactions \
  -H "Idempotency-Key: unique-key-123" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": 100.00}'
```

## Complete Testing Workflow

### 1. Create Two Accounts

```bash
# Create account 1 with $1000.50
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": 1000.50}'

# Create account 2 with $500.00
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 2, "initial_balance": 500.00}'
```

### 2. Check Initial Balances

```bash
curl http://localhost:8080/accounts/1
# {"account_id":1,"balance":1000.50}

curl http://localhost:8080/accounts/2
# {"account_id":2,"balance":500.00}
```

### 3. Transfer Money

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: test-transfer-001" \
  -d '{
    "source_account_id": 1,
    "destination_account_id": 2,
    "amount": 250.25
  }'
```

### 4. Verify Updated Balances

```bash
curl http://localhost:8080/accounts/1
# {"account_id":1,"balance":750.25}

curl http://localhost:8080/accounts/2
# {"account_id":2,"balance":750.25}
```

### 5. Test Idempotency (Retry Same Transfer)

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: test-transfer-001" \
  -d '{
    "source_account_id": 1,
    "destination_account_id": 2,
    "amount": 250.25
  }'

# Returns same transaction, balances unchanged
curl http://localhost:8080/accounts/1
# {"account_id":1,"balance":750.25}  ← Still 750.25
```

### 6. Test Insufficient Funds

```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "source_account_id": 1,
    "destination_account_id": 2,
    "amount": 10000.00
  }'

# Response: 422 Unprocessable Entity
# {"error":"Insufficient funds"}
```

## Running Tests

### Unit Tests

Test business logic and conversions:
```bash
go test -v ./internal/models/... ./tests/unit/...
```

### Integration Tests

Test with real PostgreSQL database:
```bash
# Make sure database is running
docker-compose up -d

# Run integration tests
go test -v ./tests/integration/...
```

### All Tests with Coverage

```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

### Concurrency Tests

Test race conditions and data consistency:
```bash
go test -v -race ./tests/integration/concurrency_test.go ./tests/integration/api_test.go
```

## Project Structure

```
financial-ledger-project/
├── cmd/
│   ├── api/main.go              # API server entry point
│   └── migrate/main.go          # Database migration runner
├── internal/
│   ├── models/                  # Domain models
│   │   ├── account.go           # Account model (int64 balance)
│   │   ├── transaction.go       # Transaction model
│   │   ├── conversions.go       # Cents ↔ Float helpers
│   │   └── errors.go            # Domain errors
│   ├── service/                 # Business logic
│   │   ├── account_service.go
│   │   └── transfer_service.go
│   ├── repository/              # Database operations
│   │   ├── account_repo.go
│   │   └── transaction_repo.go
│   ├── handler/                 # HTTP handlers
│   │   ├── account_handler.go
│   │   └── transaction_handler.go
│   └── database/
│       ├── postgres.go          # Connection pool
│       └── migrations/          # SQL migrations
├── tests/
│   ├── unit/                    # Unit tests
│   └── integration/             # Integration & concurrency tests
├── docker-compose.yml           # PostgreSQL setup
├── .env.example                 # Environment config template
└── README.md
```

## Configuration

Environment variables (in `.env`):

```bash
# Database
DATABASE_HOST=localhost
DATABASE_PORT=5432
DATABASE_USER=ledger_user
DATABASE_PASSWORD=ledger_pass
DATABASE_NAME=financial_ledger
DATABASE_SSLMODE=disable

# API
API_PORT=8080
```

## Database Schema

### Accounts Table
```sql
CREATE TABLE accounts (
    id BIGINT PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0,  -- stored in cents
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT positive_balance CHECK (balance >= 0)
);
```

### Transactions Table
```sql
CREATE TABLE transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_account_id BIGINT NOT NULL REFERENCES accounts(id),
    destination_account_id BIGINT NOT NULL REFERENCES accounts(id),
    amount BIGINT NOT NULL,  -- stored in cents
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT positive_amount CHECK (amount > 0),
    CONSTRAINT different_accounts CHECK (source_account_id != destination_account_id)
);
```

## Development

### Stop Database
```bash
docker-compose down
```

### Clean Database (Remove All Data)
```bash
docker-compose down -v
```

### View Logs
```bash
# API server logs (if running in background)
docker-compose logs -f

# Database logs
docker-compose logs -f postgres
```

### Database Connection
```bash
# Connect to PostgreSQL CLI
docker exec -it financial-ledger-db psql -U ledger_user -d financial_ledger

# Example queries
SELECT * FROM accounts;
SELECT * FROM transactions;
```

## Key Design Decisions

### 1. Cents-Based Storage
- **Problem**: `float64` has precision errors (`0.1 + 0.2 = 0.30000000000000004`)
- **Solution**: Store as `BIGINT` cents, convert to `float64` only in API responses
- **Result**: Perfect precision for all financial calculations

### 2. Row-Level Locking
- **Problem**: Concurrent transfers could create race conditions
- **Solution**: Use `SELECT FOR UPDATE` to lock accounts during transfers
- **Result**: Strong consistency, no money lost or created

### 3. Database Transactions
- **Problem**: Balance updates must be atomic (both succeed or both fail)
- **Solution**: Wrap all updates in a database transaction
- **Result**: Guaranteed atomicity

### 4. Idempotency
- **Problem**: Network retries could duplicate transfers
- **Solution**: Optional `Idempotency-Key` header with unique constraint
- **Result**: Safe to retry requests

## Assumptions

1. **Single Currency**: All accounts use the same currency
2. **No Authentication**: No auth/authz required (as per spec)
3. **Synchronous Processing**: Transfers execute immediately (not queued)
4. **Account Pre-creation**: Accounts must exist before transferring
5. **Non-negative Balances**: Accounts cannot go negative

## License

MIT License
