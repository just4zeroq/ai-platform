-- +goose Up
CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    type INT NOT NULL DEFAULT 0,
    key TEXT,
    name VARCHAR(128),
    models TEXT,
    group_name VARCHAR(64) NOT NULL DEFAULT 'default',
    status INT NOT NULL DEFAULT 1,
    priority BIGINT NOT NULL DEFAULT 0,
    weight INT NOT NULL DEFAULT 0,
    model_mapping TEXT,
    base_url VARCHAR(255) NOT NULL DEFAULT '',
    created_at BIGINT NOT NULL DEFAULT 0,
    updated_at BIGINT NOT NULL DEFAULT 0,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_channels_group ON channels(group_name);

-- +goose Down
DROP TABLE IF EXISTS channels;
