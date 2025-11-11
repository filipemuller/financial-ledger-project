package handler

import (
	"encoding/json"
	"net/http"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/filipe/financial-ledger-project/internal/service"
)

type TransactionHandler struct {
	transferService *service.TransferService
}

func NewTransactionHandler(transferService *service.TransferService) *TransactionHandler {
	return &TransactionHandler{
		transferService: transferService,
	}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid JSON"})
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")

	transaction, err := h.transferService.Transfer(r.Context(), req, idempotencyKey)
	if err != nil {
		sendError(w, err)
		return
	}

	sendJSON(w, http.StatusCreated, transaction)
}
