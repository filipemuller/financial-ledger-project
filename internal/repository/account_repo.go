package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/filipe/financial-ledger-project/internal/models"
	"github.com/lib/pq"
)

// AccountRepository handles account database operations
type AccountRepository struct {
	db *sql.DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create inserts a new account into the database
func (r *AccountRepository) Create(ctx context.Context, account *models.Account) error {
	query := `
		INSERT INTO accounts (id, balance, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query, account.ID, account.Balance)
	if err != nil {
		// Check for unique constraint violation (duplicate account ID)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return models.ErrAccountExists
		}
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetByID retrieves an account by its ID
func (r *AccountRepository) GetByID(ctx context.Context, id int64) (*models.Account, error) {
	query := `
		SELECT id, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	var account models.Account
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// GetForUpdate retrieves an account and locks it for update within a transaction
// This prevents race conditions when updating balances
func (r *AccountRepository) GetForUpdate(ctx context.Context, tx *sql.Tx, id int64) (*models.Account, error) {
	query := `
		SELECT id, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`

	var account models.Account
	err := tx.QueryRowContext(ctx, query, id).Scan(
		&account.ID,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrAccountNotFound
		}
		return nil, fmt.Errorf("failed to get account for update: %w", err)
	}

	return &account, nil
}

// UpdateBalance updates an account's balance within a transaction
func (r *AccountRepository) UpdateBalance(ctx context.Context, tx *sql.Tx, id int64, newBalance int64) error {
	query := `
		UPDATE accounts
		SET balance = $1, updated_at = NOW()
		WHERE id = $2
	`

	result, err := tx.ExecContext(ctx, query, newBalance, id)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return models.ErrAccountNotFound
	}

	return nil
}
