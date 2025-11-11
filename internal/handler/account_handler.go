package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/filipe/financial-ledger-project/internal/service"
	"github.com/go-chi/chi/v5"
)

// AccountHandler handles account-related HTTP requests
type AccountHandler struct {
	accountService *service.AccountService
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

// CreateAccount handles POST /accounts
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest

	// Decode request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid JSON"})
		return
	}

	// Create account
	if err := h.accountService.CreateAccount(r.Context(), req); err != nil {
		sendError(w, err)
		return
	}

	// Return 201 Created with empty response
	w.WriteHeader(http.StatusCreated)
}

// GetAccount handles GET /accounts/{account_id}
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	// Parse account ID from URL
	accountIDStr := chi.URLParam(r, "account_id")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid account ID"})
		return
	}

	// Get account balance
	account, err := h.accountService.GetAccountBalance(r.Context(), accountID)
	if err != nil {
		sendError(w, err)
		return
	}

	// Return account data
	sendJSON(w, http.StatusOK, account)
}
