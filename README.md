# Financial Ledger System

An internal transfers system built with Go and PostgreSQL. Provides three core endpoints for account management and money transfers with strong consistency guarantees and precise financial calculations.

## Key Features

- **Precise Money Handling** - Stores amounts as cents (integers) to avoid floating-point errors
- **Strong Consistency** - Database transactions with row-level locking prevent race conditions
- **Idempotency** - Optional keys prevent duplicate transfers on network retries
- **Comprehensive Testing** - Unit, integration, and concurrency tests included

## How It Works

The system uses a layered architecture where amounts are stored as integers (cents) in PostgreSQL but exposed as decimals through the API:

```
Request: {"amount": 100.50}
   ↓
Handler: Validates input
   ↓
Service: Converts 100.50 → 10050 cents
   ↓
Repository: Stores 10050 in database (BIGINT)
   ↓
Database: UPDATE accounts SET balance = balance - 10050
   ↓
Response: {"amount": 100.50}
```

**Why cents?** Floating-point arithmetic is imprecise, so storing as integers guarantees exact calculations, which is something really important for financial systems.

## Quick Start

**Prerequisites:** Go 1.21+, Docker, Docker Compose

```bash
# 1. Start database
docker-compose up -d

# 2. Run migrations
go run cmd/migrate/main.go

# 3. Start API server
go run cmd/api/main.go
```

API runs at `http://localhost:8080`

**Using Makefile:**
```bash
make setup  # Start DB + run migrations
make run    # Start API server
make test   # Run all tests
```

## API Endpoints

### POST /accounts - Create Account
```bash
curl -X POST http://localhost:8080/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": 1000.50}'
```
Returns: `201 Created`

### GET /accounts/{id} - Get Balance
```bash
curl http://localhost:8080/accounts/1
```
Returns: `{"account_id": 1, "balance": 1000.50}`

### POST /transactions - Transfer Money
```bash
curl -X POST http://localhost:8080/transactions \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: unique-key-123" \
  -d '{
    "source_account_id": 1,
    "destination_account_id": 2,
    "amount": 250.25
  }'
```
Returns: `{"transaction_id": "...", "status": "completed", ...}`

**Idempotency:** Using the same `Idempotency-Key` in multiple requests returns the original transaction without re-executing the transfer.

**Error Codes:**
- `400` - Invalid input
- `404` - Account not found
- `409` - Duplicate idempotency key or account already exists
- `422` - Insufficient funds

See [QUICKSTART.md](QUICKSTART.md) for detailed testing workflow.

## Testing

```bash
go test -v ./...                    # Run all tests
go test -v ./internal/models/...    # Unit tests only
go test -v ./tests/integration/...  # Integration tests
go test -v -race ./...              # With race detection
```

Includes unit tests, integration tests with real PostgreSQL, and concurrency tests to verify no money is lost or created under parallel load.

## Project Structure

```
cmd/                    # Entry points
  ├── api/             # HTTP server
  └── migrate/         # Database migrations
internal/
  ├── models/          # Domain models (Account, Transaction)
  ├── service/         # Business logic layer
  ├── repository/      # Database operations
  ├── handler/         # HTTP handlers
  └── database/        # Connection pool + migrations
tests/
  ├── unit/            # Unit tests
  └── integration/     # Integration + concurrency tests
```

## Database Schema

```sql
-- Amounts stored as BIGINT (cents)
CREATE TABLE accounts (
    id BIGINT PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT positive_balance CHECK (balance >= 0)
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    source_account_id BIGINT REFERENCES accounts(id),
    destination_account_id BIGINT REFERENCES accounts(id),
    amount BIGINT NOT NULL,
    idempotency_key VARCHAR(255) UNIQUE,
    CONSTRAINT positive_amount CHECK (amount > 0),
    CONSTRAINT different_accounts CHECK (source_account_id != destination_account_id)
);
```

## Key Design Decisions

**1. Integer Cents Storage**

The most critical decision: store all amounts as integers (cents) in the database, not decimals or floats.

- Floats have precision errors: `0.1 + 0.2 = 0.30000000000000004`
- Financial calculations must be exact
- Solution: `100.50` (float) → `10050` (cents in DB) → `100.50` (float in API)
- PostgreSQL uses `BIGINT` which can store up to 92 trillion dollars

**2. Atomic Transfers with Row Locking**

Transfers acquire locks on both accounts to prevent race conditions:

```sql
BEGIN;
SELECT * FROM accounts WHERE id = 1 FOR UPDATE;  -- Locks account 1
SELECT * FROM accounts WHERE id = 2 FOR UPDATE;  -- Locks account 2
UPDATE accounts SET balance = balance - 10050 WHERE id = 1;
UPDATE accounts SET balance = balance + 10050 WHERE id = 2;
COMMIT;  -- Both updates or neither
```

This ensures no money is lost or created, even under high concurrency.

**3. Idempotency Keys**

Network failures can cause clients to retry requests. Without idempotency, a transfer could execute twice. The `Idempotency-Key` header + unique database constraint ensures duplicate requests return the original result without re-executing.

## Project Assumptions

- **Single currency** - All accounts use the same currency (no exchange rates)
- **No authentication** - Simplified for internal use (no user auth required)
- **Synchronous transfers** - Executes immediately, not queued/async
- **Pre-created accounts** - Accounts must exist before transfers
- **No overdrafts** - Balances cannot go negative

## Development Process

This system was designed and implemented with AI assistance (Claude):

- **System design** - Layered architecture, cents-based storage approach, and concurrency strategy
- **Database migrations** - Schema design with appropriate constraints and indexes
- **Test coverage** - Comprehensive test suite including edge cases and concurrency scenarios
- **Implementation** - Clean, production-ready Go code following best practices

The AI helped accelerate development while ensuring financial accuracy and data consistency were prioritized throughout.

## License

MIT License
