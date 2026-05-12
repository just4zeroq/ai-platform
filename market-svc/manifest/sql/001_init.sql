CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE listings (
    id BIGSERIAL PRIMARY KEY,
    seller_id BIGINT NOT NULL,
    product_type VARCHAR(32) NOT NULL,
    title VARCHAR(256) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price DECIMAL(20,6) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    sold INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending_review',
    review_remark TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_listings_status ON listings(status);
CREATE INDEX idx_listings_seller ON listings(seller_id);

CREATE TABLE trades (
    id BIGSERIAL PRIMARY KEY,
    listing_id BIGINT NOT NULL REFERENCES listings(id),
    buyer_id BIGINT NOT NULL,
    seller_id BIGINT NOT NULL,
    amount DECIMAL(20,6) NOT NULL,
    platform_fee DECIMAL(20,6) NOT NULL DEFAULT 0,
    seller_income DECIMAL(20,6) NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'completed',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_trades_buyer ON trades(buyer_id);
CREATE INDEX idx_trades_seller ON trades(seller_id);

CREATE TABLE settlements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(32) NOT NULL,
    amount DECIMAL(20,6) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    reference_trade_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_settlements_user ON settlements(user_id);
