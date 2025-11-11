package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/filipe/financial-ledger-project/internal/service"
	"github.com/go-chi/chi/v5"
)

type AccountHandler struct {
	accountService *service.AccountService
}

func NewAccountHandler(accountService *service.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}

func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAccountRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid JSON"})
		return
	}

	if err := h.accountService.CreateAccount(r.Context(), req); err != nil {
		sendError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	accountIDStr := chi.URLParam(r, "account_id")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid account ID"})
		return
	}

	account, err := h.accountService.GetAccountBalance(r.Context(), accountID)
	if err != nil {
		sendError(w, err)
		return
	}

	sendJSON(w, http.StatusOK, account)
}
