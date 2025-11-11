package unit

import (
	"testing"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateAccountRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		req         models.CreateAccountRequest
		expectError error
	}{
		{
			name: "Valid request",
			req: models.CreateAccountRequest{
				AccountID:      1,
				InitialBalance: 100.0,
			},
			expectError: nil,
		},
		{
			name: "Valid request with zero balance",
			req: models.CreateAccountRequest{
				AccountID:      1,
				InitialBalance: 0.0,
			},
			expectError: nil,
		},
		{
			name: "Invalid account ID - zero",
			req: models.CreateAccountRequest{
				AccountID:      0,
				InitialBalance: 100.0,
			},
			expectError: models.ErrInvalidAccountID,
		},
		{
			name: "Invalid account ID - negative",
			req: models.CreateAccountRequest{
				AccountID:      -1,
				InitialBalance: 100.0,
			},
			expectError: models.ErrInvalidAccountID,
		},
		{
			name: "Negative initial balance",
			req: models.CreateAccountRequest{
				AccountID:      1,
				InitialBalance: -100.0,
			},
			expectError: models.ErrNegativeBalance,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateTransactionRequest_Validate(t *testing.T) {
	tests := []struct {
		name        string
		req         models.CreateTransactionRequest
		expectError error
	}{
		{
			name: "Valid request",
			req: models.CreateTransactionRequest{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               100.0,
			},
			expectError: nil,
		},
		{
			name: "Invalid source account ID",
			req: models.CreateTransactionRequest{
				SourceAccountID:      0,
				DestinationAccountID: 2,
				Amount:               100.0,
			},
			expectError: models.ErrInvalidAccountID,
		},
		{
			name: "Invalid destination account ID",
			req: models.CreateTransactionRequest{
				SourceAccountID:      1,
				DestinationAccountID: 0,
				Amount:               100.0,
			},
			expectError: models.ErrInvalidAccountID,
		},
		{
			name: "Same source and destination",
			req: models.CreateTransactionRequest{
				SourceAccountID:      1,
				DestinationAccountID: 1,
				Amount:               100.0,
			},
			expectError: models.ErrSameAccount,
		},
		{
			name: "Zero amount",
			req: models.CreateTransactionRequest{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               0.0,
			},
			expectError: models.ErrInvalidAmount,
		},
		{
			name: "Negative amount",
			req: models.CreateTransactionRequest{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               -100.0,
			},
			expectError: models.ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.expectError != nil {
				assert.ErrorIs(t, err, tt.expectError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
