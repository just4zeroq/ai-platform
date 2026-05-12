-- +goose Up
CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(32) NOT NULL,
    amount DECIMAL(20,4) NOT NULL,
    balance_before DECIMAL(20,4) NOT NULL,
    balance_after DECIMAL(20,4) NOT NULL,
    reference_type VARCHAR(64) NOT NULL DEFAULT '',
    reference_id BIGINT NOT NULL DEFAULT 0,
    remark TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_type ON transactions(type);

-- +goose Down
DROP TABLE IF EXISTS transactions;
