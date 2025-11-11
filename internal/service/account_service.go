package service

import (
	"context"
	"fmt"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/filipe/financial-ledger-project/internal/repository"
)

// AccountService handles account business logic
type AccountService struct {
	accountRepo *repository.AccountRepository
}

// NewAccountService creates a new account service
func NewAccountService(accountRepo *repository.AccountRepository) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
	}
}

// CreateAccount creates a new account with the given initial balance
func (s *AccountService) CreateAccount(ctx context.Context, req models.CreateAccountRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return err
	}

	// Convert balance from float to cents
	balanceInCents := models.FloatToCents(req.InitialBalance)

	// Create account model
	account := &models.Account{
		ID:      req.AccountID,
		Balance: balanceInCents,
	}

	// Save to database
	if err := s.accountRepo.Create(ctx, account); err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetAccountBalance retrieves the balance for a given account
func (s *AccountService) GetAccountBalance(ctx context.Context, accountID int64) (*models.AccountResponse, error) {
	// Validate account ID
	if accountID <= 0 {
		return nil, models.ErrInvalidAccountID
	}

	// Get account from database
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Convert to response (converts cents to float)
	response := account.ToResponse()
	return &response, nil
}
