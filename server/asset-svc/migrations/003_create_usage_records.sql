-- +goose Up
CREATE TABLE usage_records (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    api_key_id BIGINT NOT NULL DEFAULT 0,
    model_name VARCHAR(255) NOT NULL DEFAULT '',
    prompt_tokens INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    quota DECIMAL(20,4) NOT NULL DEFAULT 0,
    request_id VARCHAR(255) NOT NULL DEFAULT '',
    channel_id BIGINT NOT NULL DEFAULT 0,
    channel_name VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_records_user_id ON usage_records(user_id);
CREATE INDEX idx_usage_records_request_id ON usage_records(request_id);

-- +goose Down
DROP TABLE IF EXISTS usage_records;
