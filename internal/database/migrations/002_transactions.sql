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

CREATE INDEX IF NOT EXISTS idx_transactions_source ON transactions(source_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_destination ON transactions(destination_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_idempotency ON transactions(idempotency_key)
WHERE idempotency_key IS NOT NULL;
