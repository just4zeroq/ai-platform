-- +goose Up
CREATE TABLE balances (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE,
    balance DECIMAL(20,4) NOT NULL DEFAULT 0,
    total_recharged DECIMAL(20,4) NOT NULL DEFAULT 0,
    total_consumed DECIMAL(20,4) NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_balances_user_id ON balances(user_id);

-- +goose Down
DROP TABLE IF EXISTS balances;
