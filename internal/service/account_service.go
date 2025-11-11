package service

import (
	"context"
	"fmt"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/filipe/financial-ledger-project/internal/repository"
)

type AccountService struct {
	accountRepo *repository.AccountRepository
}

func NewAccountService(accountRepo *repository.AccountRepository) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, req models.CreateAccountRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	balanceInCents := models.FloatToCents(req.InitialBalance)

	account := &models.Account{
		ID:      req.AccountID,
		Balance: balanceInCents,
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

func (s *AccountService) GetAccountBalance(ctx context.Context, accountID int64) (*models.AccountResponse, error) {
	if accountID <= 0 {
		return nil, models.ErrInvalidAccountID
	}

	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	response := account.ToResponse()
	return &response, nil
}
