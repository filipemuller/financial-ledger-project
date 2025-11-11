package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConcurrentTransfers_BalanceConsistency verifies that concurrent transfers
// maintain data consistency - no money is created or lost
func TestConcurrentTransfers_BalanceConsistency(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	initialBalance := 10000.00
	totalSystemBalance := initialBalance * 3.0

	for i := 1; i <= 3; i++ {
		body := fmt.Sprintf(`{"account_id": %d, "initial_balance": %.2f}`, i, initialBalance)
		req := httptest.NewRequest("POST", "/accounts", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	var wg sync.WaitGroup
	numGoroutines := 100
	transferAmount := 10.00

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			source := (i % 3) + 1
			dest := ((i + 1) % 3) + 1

			body := fmt.Sprintf(`{
				"source_account_id": %d,
				"destination_account_id": %d,
				"amount": %.2f
			}`, source, dest, transferAmount)

			req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		}(i)
	}

	wg.Wait()

	totalBalance := 0.0
	for i := 1; i <= 3; i++ {
		req := httptest.NewRequest("GET", fmt.Sprintf("/accounts/%d", i), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var acc models.AccountResponse
		json.NewDecoder(w.Body).Decode(&acc)
		totalBalance += acc.Balance
	}

	assert.Equal(t, totalSystemBalance, totalBalance,
		"Money was created or lost due to race condition!")
}

// TestConcurrentTransfers_FromSameAccount tests that concurrent transfers
// from the same account are handled correctly
func TestConcurrentTransfers_FromSameAccount(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	accounts := []string{
		`{"account_id": 1, "initial_balance": 1000.00}`,
		`{"account_id": 2, "initial_balance": 0.00}`,
	}

	for _, acc := range accounts {
		req := httptest.NewRequest("POST", "/accounts", bytes.NewBufferString(acc))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	var wg sync.WaitGroup
	successCount := int32(0)
	transferAmount := 25.00
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			body := fmt.Sprintf(`{
				"source_account_id": 1,
				"destination_account_id": 2,
				"amount": %.2f
			}`, transferAmount)

			req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == 201 {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(40), successCount,
		"Expected exactly 40 successful transfers")

	req := httptest.NewRequest("GET", "/accounts/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var acc1 models.AccountResponse
	json.NewDecoder(w.Body).Decode(&acc1)
	assert.Equal(t, 0.00, acc1.Balance, "Source account should be empty")

	req = httptest.NewRequest("GET", "/accounts/2", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var acc2 models.AccountResponse
	json.NewDecoder(w.Body).Decode(&acc2)
	assert.Equal(t, 1000.00, acc2.Balance, "Destination should have all money")
}

// TestConcurrentTransfers_WithIdempotency tests that idempotency works
// correctly under concurrent load
func TestConcurrentTransfers_WithIdempotency(t *testing.T) {
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

	var wg sync.WaitGroup
	numGoroutines := 10
	idempotencyKey := "unique-key-concurrent-test"
	transferAmount := 100.00

	var transactionIDs []string
	var mu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			body := fmt.Sprintf(`{
				"source_account_id": 1,
				"destination_account_id": 2,
				"amount": %.2f
			}`, transferAmount)

			req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Idempotency-Key", idempotencyKey)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == 201 {
				var txn models.TransactionResponse
				json.NewDecoder(w.Body).Decode(&txn)
				mu.Lock()
				transactionIDs = append(transactionIDs, txn.TransactionID)
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	require.NotEmpty(t, transactionIDs, "Should have at least one transaction ID")
	firstID := transactionIDs[0]
	for _, id := range transactionIDs {
		assert.Equal(t, firstID, id, "All requests should return same transaction ID")
	}

	req := httptest.NewRequest("GET", "/accounts/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	var acc models.AccountResponse
	json.NewDecoder(w.Body).Decode(&acc)
	assert.Equal(t, 900.00, acc.Balance, "Balance should only be deducted once")
}

// TestConcurrentCreates_DuplicateAccount tests that duplicate account
// creation is handled correctly
func TestConcurrentCreates_DuplicateAccount(t *testing.T) {
	router, cleanup := setupTestRouter(t)
	defer cleanup()

	var wg sync.WaitGroup
	numGoroutines := 10
	successCount := int32(0)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			body := `{"account_id": 1, "initial_balance": 100.00}`
			req := httptest.NewRequest("POST", "/accounts", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code == 201 {
				atomic.AddInt32(&successCount, 1)
			}
		}()
	}

	wg.Wait()

	assert.Equal(t, int32(1), successCount,
		"Only one account creation should succeed")
}
