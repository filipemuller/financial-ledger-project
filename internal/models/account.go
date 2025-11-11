package models

import (
	"encoding/json"
	"time"
)

type Account struct {
	ID        int64      `db:"id"`
	Balance   int64      `db:"balance"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

type AccountResponse struct {
	AccountID int64   `json:"account_id"`
	Balance   float64 `json:"balance"`
}

func (a *Account) ToResponse() AccountResponse {
	return AccountResponse{
		AccountID: a.ID,
		Balance:   CentsToFloat(a.Balance),
	}
}

type CreateAccountRequest struct {
	AccountID      int64   `json:"account_id"`
	InitialBalance float64 `json:"initial_balance"`
}

func (r *CreateAccountRequest) Validate() error {
	if r.AccountID <= 0 {
		return ErrInvalidAccountID
	}
	if r.InitialBalance < 0 {
		return ErrNegativeBalance
	}
	return nil
}

func (a *Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.ToResponse())
}
