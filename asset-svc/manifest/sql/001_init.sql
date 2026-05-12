CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE balances (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL UNIQUE,
    balance DECIMAL(20,6) NOT NULL DEFAULT 0,
    total_recharged DECIMAL(20,6) NOT NULL DEFAULT 0,
    total_consumed DECIMAL(20,6) NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_balances_user ON balances(user_id);
CREATE INDEX idx_balances_tenant ON balances(tenant_id);

CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    type VARCHAR(32) NOT NULL,
    amount DECIMAL(20,6) NOT NULL,
    balance_before DECIMAL(20,6) NOT NULL,
    balance_after DECIMAL(20,6) NOT NULL,
    reference_type VARCHAR(32) NOT NULL DEFAULT '',
    reference_id BIGINT NOT NULL DEFAULT 0,
    remark VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tx_user ON transactions(user_id);
CREATE INDEX idx_tx_created ON transactions(created_at DESC);

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL DEFAULT 0,
    type VARCHAR(32) NOT NULL,
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    provider VARCHAR(32) NOT NULL DEFAULT 'platform',
    provider_user_id BIGINT NOT NULL DEFAULT 0,
    price DECIMAL(20,6) NOT NULL,
    model_name VARCHAR(128) NOT NULL DEFAULT '',
    model_config JSONB DEFAULT '{}',
    status VARCHAR(32) NOT NULL DEFAULT 'active',
    stock INT NOT NULL DEFAULT -1,
    sold_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_products_type ON products(type);
CREATE INDEX idx_products_status ON products(status);

CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    order_no VARCHAR(64) NOT NULL UNIQUE,
    tenant_id INT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    type VARCHAR(32) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(20,6) NOT NULL,
    payment_method VARCHAR(32) NOT NULL DEFAULT '',
    payment_trade_no VARCHAR(128) NOT NULL DEFAULT '',
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_orders_user ON orders(user_id);
CREATE INDEX idx_orders_trade_no ON orders(payment_trade_no);

CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    product_id BIGINT NOT NULL,
    product_type VARCHAR(32) NOT NULL,
    product_name VARCHAR(128) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    unit_price DECIMAL(20,6) NOT NULL,
    subtotal DECIMAL(20,6) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE usage_records (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    api_key_id BIGINT NOT NULL DEFAULT 0,
    model_name VARCHAR(128) NOT NULL,
    prompt_tokens INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    quota DECIMAL(20,6) NOT NULL DEFAULT 0,
    request_id VARCHAR(64) NOT NULL DEFAULT '',
    channel_id INT NOT NULL DEFAULT 0,
    channel_name VARCHAR(128) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_usage_user ON usage_records(user_id);
CREATE INDEX idx_usage_created ON usage_records(created_at DESC);
CREATE INDEX idx_usage_request ON usage_records(request_id);
