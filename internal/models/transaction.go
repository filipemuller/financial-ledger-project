package models

import (
	"encoding/json"
	"time"
)

type Transaction struct {
	ID                   string    `db:"id"`
	SourceAccountID      int64     `db:"source_account_id"`
	DestinationAccountID int64     `db:"destination_account_id"`
	Amount               int64     `db:"amount"`
	Status               string    `db:"status"`
	IdempotencyKey       *string   `db:"idempotency_key"`
	CreatedAt            time.Time `db:"created_at"`
}

type TransactionResponse struct {
	TransactionID        string    `json:"transaction_id"`
	SourceAccountID      int64     `json:"source_account_id"`
	DestinationAccountID int64     `json:"destination_account_id"`
	Amount               float64   `json:"amount"`
	Status               string    `json:"status"`
	CreatedAt            time.Time `json:"created_at"`
}

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

type CreateTransactionRequest struct {
	SourceAccountID      int64   `json:"source_account_id"`
	DestinationAccountID int64   `json:"destination_account_id"`
	Amount               float64 `json:"amount"`
}

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

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.ToResponse())
}
