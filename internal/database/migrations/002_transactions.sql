-- Create transactions table
-- Amount is stored as BIGINT representing cents

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_account_id BIGINT NOT NULL,
    destination_account_id BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'completed',
    idempotency_key VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_source_account FOREIGN KEY (source_account_id) REFERENCES accounts(id),
    CONSTRAINT fk_destination_account FOREIGN KEY (destination_account_id) REFERENCES accounts(id),
    CONSTRAINT positive_amount CHECK (amount > 0),
    CONSTRAINT different_accounts CHECK (source_account_id != destination_account_id),
    CONSTRAINT unique_idempotency_key UNIQUE (idempotency_key)
);

-- Indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_transactions_source ON transactions(source_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_destination ON transactions(destination_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_idempotency ON transactions(idempotency_key)
WHERE idempotency_key IS NOT NULL;

-- Add comments for documentation
COMMENT ON COLUMN transactions.amount IS 'Transfer amount stored in cents (e.g., 25050 = $250.50)';
COMMENT ON COLUMN transactions.idempotency_key IS 'Optional key to prevent duplicate transactions';
