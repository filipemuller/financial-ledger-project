package models

import (
	"encoding/json"
	"time"
)

// Account represents a financial account
// Balance is stored internally as int64 (cents) but exposed as float64 in JSON
type Account struct {
	ID        int64     `db:"id"`
	Balance   int64     `db:"balance"` // stored in cents
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// AccountResponse is the JSON response structure for accounts
type AccountResponse struct {
	AccountID int64   `json:"account_id"`
	Balance   float64 `json:"balance"` // exposed as float
}

// ToResponse converts internal Account model to API response
func (a *Account) ToResponse() AccountResponse {
	return AccountResponse{
		AccountID: a.ID,
		Balance:   CentsToFloat(a.Balance),
	}
}

// CreateAccountRequest represents the request to create a new account
type CreateAccountRequest struct {
	AccountID      int64   `json:"account_id"`
	InitialBalance float64 `json:"initial_balance"`
}

// Validate validates the create account request
func (r *CreateAccountRequest) Validate() error {
	if r.AccountID <= 0 {
		return ErrInvalidAccountID
	}
	if r.InitialBalance < 0 {
		return ErrNegativeBalance
	}
	return nil
}

// MarshalJSON customizes JSON marshaling for Account
func (a *Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.ToResponse())
}
