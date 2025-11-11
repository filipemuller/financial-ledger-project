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

// TransferService handles money transfer business logic
type TransferService struct {
	db          *sql.DB
	accountRepo *repository.AccountRepository
	txnRepo     *repository.TransactionRepository
}

// NewTransferService creates a new transfer service
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

// Transfer executes a money transfer between two accounts
func (s *TransferService) Transfer(
	ctx context.Context,
	req models.CreateTransactionRequest,
	idempotencyKey string,
) (*models.TransactionResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check idempotency - if key exists, return existing transaction
	if idempotencyKey != "" {
		if existingTxn, err := s.txnRepo.GetByIdempotencyKey(ctx, idempotencyKey); err == nil {
			// Transaction already exists with this key, return it
			response := existingTxn.ToResponse()
			return &response, nil
		} else if !errors.Is(err, models.ErrTransactionNotFound) {
			// Unexpected error
			return nil, fmt.Errorf("failed to check idempotency: %w", err)
		}
		// Transaction not found, proceed with creation
	}

	// Convert amount from float to cents
	amountInCents := models.FloatToCents(req.Amount)
	if amountInCents <= 0 {
		return nil, models.ErrInvalidAmount
	}

	// Begin database transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Lock and retrieve source account
	sourceAccount, err := s.accountRepo.GetForUpdate(ctx, tx, req.SourceAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source account: %w", err)
	}

	// Lock and retrieve destination account
	destAccount, err := s.accountRepo.GetForUpdate(ctx, tx, req.DestinationAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get destination account: %w", err)
	}

	// Check sufficient funds
	if sourceAccount.Balance < amountInCents {
		return nil, models.ErrInsufficientFunds
	}

	// Calculate new balances
	newSourceBalance := sourceAccount.Balance - amountInCents
	newDestBalance := destAccount.Balance + amountInCents

	// Update source account balance
	if err := s.accountRepo.UpdateBalance(ctx, tx, sourceAccount.ID, newSourceBalance); err != nil {
		return nil, fmt.Errorf("failed to update source balance: %w", err)
	}

	// Update destination account balance
	if err := s.accountRepo.UpdateBalance(ctx, tx, destAccount.ID, newDestBalance); err != nil {
		return nil, fmt.Errorf("failed to update destination balance: %w", err)
	}

	// Create transaction record
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

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return response (converts cents to float)
	response := transaction.ToResponse()
	return &response, nil
}
