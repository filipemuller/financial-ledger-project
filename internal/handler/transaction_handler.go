package handler

import (
	"encoding/json"
	"net/http"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/filipe/financial-ledger-project/internal/service"
)

// TransactionHandler handles transaction-related HTTP requests
type TransactionHandler struct {
	transferService *service.TransferService
}

// NewTransactionHandler creates a new transaction handler
func NewTransactionHandler(transferService *service.TransferService) *TransactionHandler {
	return &TransactionHandler{
		transferService: transferService,
	}
}

// CreateTransaction handles POST /transactions
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransactionRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid JSON"})
		return
	}

	// Extract idempotency key from header
	idempotencyKey := r.Header.Get("Idempotency-Key")

	// Execute transfer
	transaction, err := h.transferService.Transfer(r.Context(), req, idempotencyKey)
	if err != nil {
		sendError(w, err)
		return
	}

	// Return 201 Created with transaction data
	sendJSON(w, http.StatusCreated, transaction)
}
