-- +goose Up
CREATE TABLE api_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    key VARCHAR(255) NOT NULL UNIQUE,
    status INT NOT NULL DEFAULT 1,
    expire_time TIMESTAMP,
    model_limits TEXT NOT NULL DEFAULT '',
    model_limits_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    group_name VARCHAR(64) NOT NULL DEFAULT 'default',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX idx_api_keys_key ON api_keys(key);

-- +goose Down
DROP TABLE IF EXISTS api_keys;
