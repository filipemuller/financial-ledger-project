package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/filipe/financial-ledger-project/internal/database"
	"github.com/filipe/financial-ledger-project/internal/handler"
	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/filipe/financial-ledger-project/internal/repository"
	"github.com/filipe/financial-ledger-project/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouter creates a test router with all dependencies
func setupTestRouter(t *testing.T) (*chi.Mux, func()) {
	cfg := database.Config{
		Host:     "localhost",
		Port:     "5432",
		User:     "ledger_user",
		Password: "ledger_pass",
		DBName:   "financial_ledger",
		SSLMode:  "disable",
	}

	db, err := database.NewPostgresDB(cfg)
	require.NoError(t, err, "Failed to connect to test database")

	_, err = db.Exec("TRUNCATE accounts, transactions CASCADE")
	require.NoError(t, err, "Failed to truncate tables")

	accountRepo := repository.NewAccountRepository(db)
	transactionRepo := repository.NewTransactionRepository(db)

	accountService := service.NewAccountService(accountRepo)
	transferService := service.NewTransferService(db, accountRepo, transactionRepo)

	accountHandler := handler.NewAccountHandler(accountService)
	transactionHandler := handler.NewTransactionHandler(transferService)

	r := chi.NewRouter()
	r.Post("/accounts", accountHandler.CreateAccount)
	r.Get("/accounts/{account_id}", accountHandler.GetAccount)
	r.Post("/transactions", transactionHandler.CreateTransaction)

	cleanup := func() {
		db.Close()
	}

	return r, cleanup
}

func TestAPI_CreateAccount(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "Valid account creation",
			body:           `{"account_id": 1, "initial_balance": 100.50}`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Invalid JSON",
			body:           `{"account_id": 1, "initial_balance":}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Negative balance",
			body:           `{"account_id": 2, "initial_balance": -100}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/accounts", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAPI_GetAccount(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	createReq := httptest.NewRequest("POST", "/accounts",
		bytes.NewBufferString(`{"account_id": 1, "initial_balance": 100.50}`))
	createReq.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, createReq)
	require.Equal(t, http.StatusCreated, w.Code)

	getReq := httptest.NewRequest("GET", "/accounts/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, getReq)

	assert.Equal(t, http.StatusOK, w.Code)

	var response models.AccountResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, int64(1), response.AccountID)
	assert.Equal(t, 100.50, response.Balance)
}

func TestAPI_Transfer(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	accounts := []string{
		`{"account_id": 1, "initial_balance": 1000.00}`,
		`{"account_id": 2, "initial_balance": 500.00}`,
	}

	for _, acc := range accounts {
		req := httptest.NewRequest("POST", "/accounts", bytes.NewBufferString(acc))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	transferBody := `{
		"source_account_id": 1,
		"destination_account_id": 2,
		"amount": 250.50
	}`

	req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(transferBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var txnResponse models.TransactionResponse
	err := json.NewDecoder(w.Body).Decode(&txnResponse)
	require.NoError(t, err)

	assert.NotEmpty(t, txnResponse.TransactionID)
	assert.Equal(t, int64(1), txnResponse.SourceAccountID)
	assert.Equal(t, int64(2), txnResponse.DestinationAccountID)
	assert.Equal(t, 250.50, txnResponse.Amount)
	assert.Equal(t, "completed", txnResponse.Status)

	getReq := httptest.NewRequest("GET", "/accounts/1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, getReq)
	var acc1 models.AccountResponse
	json.NewDecoder(w.Body).Decode(&acc1)
	assert.Equal(t, 749.50, acc1.Balance)

	getReq = httptest.NewRequest("GET", "/accounts/2", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, getReq)
	var acc2 models.AccountResponse
	json.NewDecoder(w.Body).Decode(&acc2)
	assert.Equal(t, 750.50, acc2.Balance)
}

func TestAPI_InsufficientFunds(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/accounts",
		bytes.NewBufferString(`{"account_id": 1, "initial_balance": 50.00}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	req = httptest.NewRequest("POST", "/accounts",
		bytes.NewBufferString(`{"account_id": 2, "initial_balance": 100.00}`))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	transferBody := `{
		"source_account_id": 1,
		"destination_account_id": 2,
		"amount": 100.00
	}`

	req = httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(transferBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestAPI_Idempotency(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	accounts := []string{
		`{"account_id": 1, "initial_balance": 1000.00}`,
		`{"account_id": 2, "initial_balance": 500.00}`,
	}

	for _, acc := range accounts {
		req := httptest.NewRequest("POST", "/accounts", bytes.NewBufferString(acc))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

	transferBody := `{
		"source_account_id": 1,
		"destination_account_id": 2,
		"amount": 100.00
	}`

	idempotencyKey := "test-key-123"

	req1 := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(transferBody))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Idempotency-Key", idempotencyKey)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusCreated, w1.Code)

	var txn1 models.TransactionResponse
	json.NewDecoder(w1.Body).Decode(&txn1)

	req2 := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(transferBody))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", idempotencyKey)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusCreated, w2.Code)

	var txn2 models.TransactionResponse
	json.NewDecoder(w2.Body).Decode(&txn2)

	assert.Equal(t, txn1.TransactionID, txn2.TransactionID)

	getReq := httptest.NewRequest("GET", "/accounts/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, getReq)
	var acc models.AccountResponse
	json.NewDecoder(w.Body).Decode(&acc)
	assert.Equal(t, 900.00, acc.Balance)
}
