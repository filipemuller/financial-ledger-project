-- Create accounts table
-- Balance is stored as BIGINT representing cents (smallest unit)
-- Example: 10050 cents = 100.50 in currency

CREATE TABLE IF NOT EXISTS accounts (
    id BIGINT PRIMARY KEY,
    balance BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT positive_balance CHECK (balance >= 0)
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_accounts_id ON accounts(id);

-- Add comment for documentation
COMMENT ON COLUMN accounts.balance IS 'Account balance stored in cents (e.g., 10050 = $100.50)';
