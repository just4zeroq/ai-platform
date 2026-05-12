CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(128) NOT NULL DEFAULT 'default',
    status INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO tenants (id, name) VALUES (0, 'Default Tenant');

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL DEFAULT 0 REFERENCES tenants(id),
    username VARCHAR(64) NOT NULL UNIQUE,
    password VARCHAR(256) NOT NULL,
    email VARCHAR(128) NOT NULL DEFAULT '',
    phone VARCHAR(32) NOT NULL DEFAULT '',
    display_name VARCHAR(128) NOT NULL DEFAULT '',
    avatar VARCHAR(512) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    role INT NOT NULL DEFAULT 1,
    group_name VARCHAR(64) NOT NULL DEFAULT 'default',
    source VARCHAR(32) NOT NULL DEFAULT 'email',
    remark VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_email ON users(email);

CREATE TABLE api_keys (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL REFERENCES users(id),
    key VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL DEFAULT '',
    status INT NOT NULL DEFAULT 1,
    model_limits_enabled BOOLEAN NOT NULL DEFAULT FALSE,
    model_limits TEXT NOT NULL DEFAULT '',
    expire_time TIMESTAMPTZ,
    allow_ips TEXT NOT NULL DEFAULT '',
    group_name VARCHAR(64) NOT NULL DEFAULT '',
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);
CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_key ON api_keys(key);

CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(64) NOT NULL UNIQUE,
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE permissions (
    id SERIAL PRIMARY KEY,
    code VARCHAR(128) NOT NULL UNIQUE,
    name VARCHAR(128) NOT NULL DEFAULT '',
    description VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE role_permissions (
    role_id INT NOT NULL REFERENCES roles(id),
    permission_id INT NOT NULL REFERENCES permissions(id),
    PRIMARY KEY (role_id, permission_id)
);

INSERT INTO roles (id, name, description) VALUES
    (1, 'user', 'Regular user'),
    (10, 'admin', 'Platform administrator'),
    (20, 'super_admin', 'Super administrator');

INSERT INTO permissions (code, name, description) VALUES
    ('user:read', 'Read users', 'View user list'),
    ('user:write', 'Write users', 'Create/edit users'),
    ('user:ban', 'Ban users', 'Disable/enable users'),
    ('asset:read', 'Read assets', 'View asset/products'),
    ('asset:write', 'Write assets', 'Create/edit products'),
    ('asset:approve', 'Approve assets', 'Approve C2C listings'),
    ('order:read', 'Read orders', 'View order list'),
    ('order:refund', 'Refund orders', 'Process refunds'),
    ('system:config', 'System config', 'Modify system settings');

INSERT INTO role_permissions (role_id, permission_id)
SELECT 20, id FROM permissions;

INSERT INTO role_permissions (role_id, permission_id)
SELECT 10, id FROM permissions WHERE code != 'system:config';
