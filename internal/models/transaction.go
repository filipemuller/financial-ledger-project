package models

import (
	"encoding/json"
	"time"
)

// Transaction represents a money transfer between accounts
// Amount is stored internally as int64 (cents) but exposed as float64 in JSON
type Transaction struct {
	ID                   string    `db:"id"`
	SourceAccountID      int64     `db:"source_account_id"`
	DestinationAccountID int64     `db:"destination_account_id"`
	Amount               int64     `db:"amount"` // stored in cents
	Status               string    `db:"status"`
	IdempotencyKey       *string   `db:"idempotency_key"`
	CreatedAt            time.Time `db:"created_at"`
}

// TransactionResponse is the JSON response structure for transactions
type TransactionResponse struct {
	TransactionID        string  `json:"transaction_id"`
	SourceAccountID      int64   `json:"source_account_id"`
	DestinationAccountID int64   `json:"destination_account_id"`
	Amount               float64 `json:"amount"` // exposed as float
	Status               string  `json:"status"`
	CreatedAt            time.Time `json:"created_at"`
}

// ToResponse converts internal Transaction model to API response
func (t *Transaction) ToResponse() TransactionResponse {
	return TransactionResponse{
		TransactionID:        t.ID,
		SourceAccountID:      t.SourceAccountID,
		DestinationAccountID: t.DestinationAccountID,
		Amount:               CentsToFloat(t.Amount),
		Status:               t.Status,
		CreatedAt:            t.CreatedAt,
	}
}

// CreateTransactionRequest represents the request to create a transfer
type CreateTransactionRequest struct {
	SourceAccountID      int64   `json:"source_account_id"`
	DestinationAccountID int64   `json:"destination_account_id"`
	Amount               float64 `json:"amount"`
}

// Validate validates the create transaction request
func (r *CreateTransactionRequest) Validate() error {
	if r.SourceAccountID <= 0 || r.DestinationAccountID <= 0 {
		return ErrInvalidAccountID
	}
	if r.SourceAccountID == r.DestinationAccountID {
		return ErrSameAccount
	}
	if r.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

// MarshalJSON customizes JSON marshaling for Transaction
func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.ToResponse())
}
