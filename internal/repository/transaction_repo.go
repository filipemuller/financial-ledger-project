package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/lib/pq"
)

// TransactionRepository handles transaction database operations
type TransactionRepository struct {
	db *sql.DB
}

// NewTransactionRepository creates a new transaction repository
func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create inserts a new transaction record within a database transaction
func (r *TransactionRepository) Create(ctx context.Context, tx *sql.Tx, transaction *models.Transaction) error {
	query := `
		INSERT INTO transactions (id, source_account_id, destination_account_id, amount, status, idempotency_key, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	_, err := tx.ExecContext(
		ctx,
		query,
		transaction.ID,
		transaction.SourceAccountID,
		transaction.DestinationAccountID,
		transaction.Amount,
		transaction.Status,
		transaction.IdempotencyKey,
	)

	if err != nil {
		// Check for unique constraint violation on idempotency key
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return models.ErrDuplicateIdempotency
		}
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetByIdempotencyKey retrieves a transaction by its idempotency key
func (r *TransactionRepository) GetByIdempotencyKey(ctx context.Context, key string) (*models.Transaction, error) {
	query := `
		SELECT id, source_account_id, destination_account_id, amount, status, idempotency_key, created_at
		FROM transactions
		WHERE idempotency_key = $1
	`

	var transaction models.Transaction
	var idempotencyKey sql.NullString

	err := r.db.QueryRowContext(ctx, query, key).Scan(
		&transaction.ID,
		&transaction.SourceAccountID,
		&transaction.DestinationAccountID,
		&transaction.Amount,
		&transaction.Status,
		&idempotencyKey,
		&transaction.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	if idempotencyKey.Valid {
		transaction.IdempotencyKey = &idempotencyKey.String
	}

	return &transaction, nil
}
