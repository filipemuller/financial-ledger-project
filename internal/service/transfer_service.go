package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/filipe/financial-ledger-project/internal/repository"
	"github.com/google/uuid"
)

type TransferService struct {
	db          *sql.DB
	accountRepo *repository.AccountRepository
	txnRepo     *repository.TransactionRepository
}

func NewTransferService(
	db *sql.DB,
	accountRepo *repository.AccountRepository,
	txnRepo *repository.TransactionRepository,
) *TransferService {
	return &TransferService{
		db:          db,
		accountRepo: accountRepo,
		txnRepo:     txnRepo,
	}
}

func (s *TransferService) Transfer(
	ctx context.Context,
	req models.CreateTransactionRequest,
	idempotencyKey string,
) (*models.TransactionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	if idempotencyKey != "" {
		if existingTxn, err := s.txnRepo.GetByIdempotencyKey(ctx, idempotencyKey); err == nil {
			response := existingTxn.ToResponse()
			return &response, nil
		} else if !errors.Is(err, models.ErrTransactionNotFound) {
			return nil, fmt.Errorf("failed to check idempotency: %w", err)
		}
	}

	amountInCents := models.FloatToCents(req.Amount)
	if amountInCents <= 0 {
		return nil, models.ErrInvalidAmount
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	sourceAccount, err := s.accountRepo.GetForUpdate(ctx, tx, req.SourceAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source account: %w", err)
	}

	destAccount, err := s.accountRepo.GetForUpdate(ctx, tx, req.DestinationAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination account: %w", err)
	}

	if sourceAccount.Balance < amountInCents {
		return nil, models.ErrInsufficientFunds
	}

	newSourceBalance := sourceAccount.Balance - amountInCents
	newDestBalance := destAccount.Balance + amountInCents

	if err := s.accountRepo.UpdateBalance(ctx, tx, sourceAccount.ID, newSourceBalance); err != nil {
		return nil, fmt.Errorf("failed to update source balance: %w", err)
	}

	if err := s.accountRepo.UpdateBalance(ctx, tx, destAccount.ID, newDestBalance); err != nil {
		return nil, fmt.Errorf("failed to update destination balance: %w", err)
	}

	var idempotencyKeyPtr *string
	if idempotencyKey != "" {
		idempotencyKeyPtr = &idempotencyKey
	}

	transaction := &models.Transaction{
		ID:                   uuid.New().String(),
		SourceAccountID:      req.SourceAccountID,
		DestinationAccountID: req.DestinationAccountID,
		Amount:               amountInCents,
		Status:               "completed",
		IdempotencyKey:       idempotencyKeyPtr,
	}

	if err := s.txnRepo.Create(ctx, tx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	response := transaction.ToResponse()
	return &response, nil
}
