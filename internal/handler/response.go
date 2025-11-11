package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/filipe/financial-ledger-project/internal/models"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func sendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
	}
}

func sendError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	errorMessage := "Internal server error"

	switch {
	case errors.Is(err, models.ErrAccountNotFound):
		statusCode = http.StatusNotFound
		errorMessage = "Account not found"
	case errors.Is(err, models.ErrAccountExists):
		statusCode = http.StatusConflict
		errorMessage = "Account already exists"
	case errors.Is(err, models.ErrInvalidAccountID):
		statusCode = http.StatusBadRequest
		errorMessage = "Invalid account ID"
	case errors.Is(err, models.ErrNegativeBalance):
		statusCode = http.StatusBadRequest
		errorMessage = "Balance cannot be negative"
	case errors.Is(err, models.ErrInsufficientFunds):
		statusCode = http.StatusUnprocessableEntity
		errorMessage = "Insufficient funds"
	case errors.Is(err, models.ErrInvalidAmount):
		statusCode = http.StatusBadRequest
		errorMessage = "Amount must be positive"
	case errors.Is(err, models.ErrSameAccount):
		statusCode = http.StatusBadRequest
		errorMessage = "Cannot transfer to same account"
	case errors.Is(err, models.ErrDuplicateIdempotency):
		statusCode = http.StatusConflict
		errorMessage = "Duplicate idempotency key"
	default:
		log.Printf("Unexpected error: %v", err)
	}

	sendJSON(w, statusCode, ErrorResponse{Error: errorMessage})
}
