-- +goose Up
CREATE TABLE abilities (
    group_name VARCHAR(64) NOT NULL,
    model VARCHAR(255) NOT NULL,
    channel_id INT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    priority BIGINT NOT NULL DEFAULT 0,
    weight INT NOT NULL DEFAULT 0,
    tag VARCHAR(255) NOT NULL DEFAULT '',
    PRIMARY KEY (group_name, model, channel_id)
);

CREATE INDEX idx_abilities_tag ON abilities(tag);

-- +goose Down
DROP TABLE IF EXISTS abilities;
