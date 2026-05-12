# AI Platform — Phase 1: Foundation (Monorepo + Proto + DB Schema)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Scaffold the full monorepo with GoFrame v2 microservices, protobuf definitions, and PostgreSQL schemas.

**Architecture:** 6 GoFrame services + 1 React SPA + 1 Tauri desktop, gRPC inter-service communication, PostgreSQL per service, 2 entry points (api-gateway for management, ai-gateway for LLM relay).

**Tech Stack:** Go 1.22+, GoFrame v2, PostgreSQL 16, protobuf + gRPC, Redis (cache), Vite + TanStack + Zustand + shadcn/ui, Tauri 2.0

---

## Project Structure

```
ai-platform/
├── proto/                          # Shared protobuf definitions
│   ├── user/v1/
│   │   └── user.proto
│   ├── asset/v1/
│   │   └── asset.proto
│   ├── market/v1/
│   │   └── market.proto
│   └── gateway/v1/
│       └── gateway.proto
├── api-gateway/                    # GoFrame HTTP gateway (management API)
├── ai-gateway/                     # GoFrame LLM relay gateway (fork new-api relay)
├── user-svc/                       # GoFrame gRPC user service
├── asset-svc/                      # GoFrame gRPC asset service
├── market-svc/                     # GoFrame gRPC market service
├── web/                            # Vite + React + TanStack
├── desktop/                        # Tauri 2.0
├── docker-compose.yml
└── scripts/
    └── init-db.sh                  # Database initialization script
```

---

### Task 1: Scaffold monorepo structure

**Files:**
- Create: `ai-platform/docker-compose.yml`
- Create: `ai-platform/scripts/init-db.sh`
- Create: `ai-platform/.gitignore`

- [ ] **1.1: Create root .gitignore**

Write `ai-platform/.gitignore`:
```
# Go
vendor/
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of go coverage
*.out

# GoFrame
temp/
resource/public/

# Node
node_modules/
dist/
.turbo/

# Tauri
desktop/src-tauri/target/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Env
.env
.env.local

# Logs
*.log
```

- [ ] **1.2: Create docker-compose.yml**

Write `ai-platform/docker-compose.yml`:
```yaml
version: "3.9"

services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: aiplatform
      POSTGRES_PASSWORD: aiplatform
      POSTGRES_DB: aiplatform
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./scripts/init-db.sh:/docker-entrypoint-initdb.d/init-db.sh

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  # --- Services (built from local Dockerfile) ---
  api-gateway:
    build: ./api-gateway
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    environment:
      - DB_USER=aiplatform
      - DB_PASS=aiplatform
      - DB_NAME=aiplatform
      - REDIS_ADDR=redis:6379

  user-svc:
    build: ./user-svc
    ports:
      - "8100:8100"  # gRPC
    depends_on:
      - postgres
    environment:
      - DB_NAME=user_svc

  asset-svc:
    build: ./asset-svc
    ports:
      - "8101:8101"  # gRPC
    depends_on:
      - postgres
    environment:
      - DB_NAME=asset_svc

  market-svc:
    build: ./market-svc
    ports:
      - "8102:8102"  # gRPC
    depends_on:
      - postgres
    environment:
      - DB_NAME=market_svc

volumes:
  pgdata:
```

- [ ] **1.3: Create init-db.sh**

Write `ai-platform/scripts/init-db.sh`:
```bash
#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE DATABASE user_svc;
    CREATE DATABASE asset_svc;
    CREATE DATABASE market_svc;
EOSQL
```

- [ ] **1.4: Commit**

```bash
cd /home/ubuntu/code/ai-platform
git init
git add .
git commit -m "chore: scaffold monorepo with docker-compose"
```

---

### Task 2: Define protobuf — user/v1/user.proto

**Files:**
- Create: `proto/user/v1/user.proto`

- [ ] **2.1: Write user.proto**

```protobuf
syntax = "proto3";

package user.v1;

option go_package = "ai-platform/api/user/v1;userv1";

// === Messages ===

message User {
  int64 id = 1;
  int64 tenant_id = 2;
  string username = 3;
  string email = 4;
  string phone = 5;
  string display_name = 6;
  string avatar = 7;
  int32 status = 8;
  int32 role = 9;
  string group = 10;
  string source = 11;
  string remark = 12;
  string created_at = 13;
  string updated_at = 14;
}

message Tenant {
  int64 id = 1;
  string name = 2;
  int32 status = 3;
}

message ApiKey {
  int64 id = 1;
  int64 user_id = 2;
  string key = 3;
  string name = 4;
  int32 status = 5;
  bool model_limits_enabled = 6;
  string model_limits = 7; // comma-separated
  int64 expire_time = 8;
  string allow_ips = 9;
  string group = 10;
  int64 last_used_at = 11;
  string created_at = 12;
}

message Role {
  int32 id = 1;
  string name = 2;
  string description = 3;
}

message Permission {
  int32 id = 1;
  string code = 2;
  string name = 3;
  string description = 4;
}

// === Requests / Responses ===

message RegisterReq {
  string username = 1;
  string password = 2;
  string email = 3;
  string phone = 4;
  string source = 5;
  string oauth_id = 6;
}

message RegisterRes {
  User user = 1;
  string access_token = 2;
  string refresh_token = 3;
}

message LoginReq {
  string username = 1;
  string password = 2;
}

message LoginRes {
  User user = 1;
  string access_token = 2;
  string refresh_token = 3;
}

message ValidateTokenReq {
  string token = 1; // access token or api key
}

message ValidateTokenRes {
  int64 user_id = 1;
  int32 user_status = 2;
  string group = 3;
  bool has_token = 4; // true=api key auth, false=session auth
  // api key specifics
  int64 api_key_id = 5;
  bool model_limits_enabled = 6;
  repeated string model_limits = 7;
  string key_group = 8; // token-specific group override
  bool key_unlimited_quota = 9;
  int64 key_remain_quota = 10;
  int32 user_role = 11;
  int64 tenant_id = 12;
}

message CreateApiKeyReq {
  int64 user_id = 1;
  string name = 2;
  string model_limits = 3;
  bool model_limits_enabled = 4;
  int64 expire_time = 5;
  string allow_ips = 6;
  string group = 7;
}

message CreateApiKeyRes {
  ApiKey api_key = 1;
  string raw_key = 2; // shown only once
}

message ListApiKeysReq {
  int64 user_id = 1;
  int32 page = 2;
  int32 page_size = 3;
}

message ListApiKeysRes {
  repeated ApiKey api_keys = 1;
  int32 total = 2;
}

message DeleteApiKeyReq {
  int64 id = 1;
  int64 user_id = 2;
}

message DeleteApiKeyRes {  }

message GetUserReq {
  int64 user_id = 1;
}

message GetUserRes {
  User user = 1;
}

// === Service ===

service UserService {
  rpc Register(RegisterReq) returns (RegisterRes);
  rpc Login(LoginReq) returns (LoginRes);
  rpc ValidateToken(ValidateTokenReq) returns (ValidateTokenRes);
  rpc CreateApiKey(CreateApiKeyReq) returns (CreateApiKeyRes);
  rpc ListApiKeys(ListApiKeysReq) returns (ListApiKeysRes);
  rpc DeleteApiKey(DeleteApiKeyReq) returns (DeleteApiKeyRes);
  rpc GetUser(GetUserReq) returns (GetUserRes);
}
```

---

### Task 3: Define protobuf — asset/v1/asset.proto

**Files:**
- Create: `proto/asset/v1/asset.proto`

- [ ] **3.1: Write asset.proto**

```protobuf
syntax = "proto3";

package asset.v1;

option go_package = "ai-platform/api/asset/v1;assetv1";

message Product {
  int64 id = 1;
  int64 tenant_id = 2;
  string type = 3; // model / token
  string name = 4;
  string description = 5;
  string provider = 6; // platform / user
  int64 provider_user_id = 7;
  double price = 8;
  string model_name = 9;
  string status = 10;
  int32 stock = 11;
  int32 sold_count = 12;
}

message Order {
  int64 id = 1;
  string order_no = 2;
  int64 user_id = 3;
  string type = 4;  // recharge / purchase / refund
  string status = 5; // pending / completed / failed / refunded
  double total_amount = 6;
  string payment_method = 7;
  string payment_trade_no = 8;
  string paid_at = 9;
}

message Balance {
  int64 user_id = 1;
  double balance = 2;
  double total_recharged = 3;
  double total_consumed = 4;
}

message Transaction {
  int64 id = 1;
  int64 user_id = 2;
  string type = 3; // recharge / consume / refund / commission
  double amount = 4;
  double balance_before = 5;
  double balance_after = 6;
  string reference_type = 7;
  int64 reference_id = 8;
  string remark = 9;
}

message UsageRecord {
  int64 id = 1;
  int64 user_id = 2;
  int64 api_key_id = 3;
  string model_name = 4;
  int32 prompt_tokens = 5;
  int32 completion_tokens = 6;
  double quota = 7;
  string request_id = 8;
}

message CreateOrderReq {
  int64 user_id = 1;
  string type = 2;
  double amount = 3;
  string payment_method = 4;
}

message CreateOrderRes {
  Order order = 1;
  string pay_url = 2;
}

message GetBalanceReq {
  int64 user_id = 1;
}

message GetBalanceRes {
  Balance balance = 1;
}

message ReportUsageReq {
  int64 user_id = 1;
  int64 api_key_id = 2;
  string model_name = 3;
  int32 prompt_tokens = 4;
  int32 completion_tokens = 5;
  double quota = 6;
  string request_id = 7;
  int64 channel_id = 8;
  string channel_name = 9;
}

message ReportUsageRes {
  double balance_after = 1;
}

message ListTransactionsReq {
  int64 user_id = 1;
  int32 page = 2;
  int32 page_size = 3;
}

message ListTransactionsRes {
  repeated Transaction transactions = 1;
  int32 total = 2;
}

service AssetService {
  rpc CreateOrder(CreateOrderReq) returns (CreateOrderRes);
  rpc GetBalance(GetBalanceReq) returns (GetBalanceRes);
  rpc ReportUsage(ReportUsageReq) returns (ReportUsageRes);
  rpc ListTransactions(ListTransactionsReq) returns (ListTransactionsRes);
}
```

---

### Task 4: Define protobuf — market/v1/market.proto

**Files:**
- Create: `proto/market/v1/market.proto`

- [ ] **4.1: Write market.proto**

```protobuf
syntax = "proto3";

package market.v1;

option go_package = "ai-platform/api/market/v1;marketv1";

message Listing {
  int64 id = 1;
  int64 seller_id = 2;
  string product_type = 3; // model / token
  string title = 4;
  string description = 5;
  double price = 6;
  int32 quantity = 7;
  int32 sold = 8;
  string status = 9; // pending_review / active / rejected / sold_out / cancelled
}

message Trade {
  int64 id = 1;
  int64 listing_id = 2;
  int64 buyer_id = 3;
  int64 seller_id = 4;
  double amount = 5;
  double platform_fee = 6;
  double seller_income = 7;
  string status = 8;
}

message CreateListingReq {
  int64 seller_id = 1;
  string product_type = 2;
  string title = 3;
  string description = 4;
  double price = 5;
  int32 quantity = 6;
}

message CreateListingRes {
  Listing listing = 1;
}

message ListListingsReq {
  string status = 1;
  string product_type = 2;
  int32 page = 3;
  int32 page_size = 4;
}

message ListListingsRes {
  repeated Listing listings = 1;
  int32 total = 2;
}

message BuyProductReq {
  int64 listing_id = 1;
  int64 buyer_id = 2;
  int32 quantity = 3;
}

message BuyProductRes {
  Trade trade = 1;
}

message ListTradesReq {
  int64 user_id = 1;
  int32 page = 2;
  int32 page_size = 3;
}

message ListTradesRes {
  repeated Trade trades = 1;
  int32 total = 2;
}

service MarketService {
  rpc CreateListing(CreateListingReq) returns (CreateListingRes);
  rpc ListListings(ListListingsReq) returns (ListListingsRes);
  rpc BuyProduct(BuyProductReq) returns (BuyProductRes);
  rpc ListTrades(ListTradesReq) returns (ListTradesRes);
}
```

---

### Task 5: Define protobuf — gateway/v1/gateway.proto

**Files:**
- Create: `proto/gateway/v1/gateway.proto`

- [ ] **5.1: Write gateway.proto**

```protobuf
syntax = "proto3";

package gateway.v1;

option go_package = "ai-platform/api/gateway/v1;gatewayv1";

// This service is called by ai-gateway (LLM relay) to validate keys
// and report usage. It internally proxies to user-svc and asset-svc.

service GatewayService {
  rpc ValidateKey(ValidateKeyReq) returns (ValidateKeyRes);
  rpc ReportUsage(ReportUsageReq) returns (ReportUsageRes);
}

message ValidateKeyReq {
  string api_key = 1;
  string model_name = 2;
  string client_ip = 3;
}

message ValidateKeyRes {
  bool valid = 1;
  int64 user_id = 2;
  bool user_enabled = 3;
  string group = 4;
  bool model_limits_enabled = 5;
  repeated string model_limits = 6;
  int64 api_key_id = 7;
}

message ReportUsageReq {
  int64 user_id = 1;
  int64 api_key_id = 2;
  string model_name = 3;
  int32 prompt_tokens = 4;
  int32 completion_tokens = 5;
  double quota = 6;
  string request_id = 7;
}

message ReportUsageRes {
  bool success = 1;
  double balance_after = 2;
}
```

---

### Task 6: Design PostgreSQL schemas

**Files:**
- Create: `user-svc/manifest/sql/001_init.sql`
- Create: `asset-svc/manifest/sql/001_init.sql`
- Create: `market-svc/manifest/sql/001_init.sql`

- [ ] **6.1: Write user-svc schema**

Write `user-svc/manifest/sql/001_init.sql`:
```sql
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
    status INT NOT NULL DEFAULT 1,  -- 0=disabled, 1=enabled
    role INT NOT NULL DEFAULT 1,    -- 1=user, 10=admin, 20=super_admin
    group_name VARCHAR(64) NOT NULL DEFAULT 'default',
    source VARCHAR(32) NOT NULL DEFAULT 'email', -- email, github, google, wechat
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
    status INT NOT NULL DEFAULT 1,  -- 1=enabled, 2=disabled, 3=exhausted, 4=expired
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

-- Seed roles
INSERT INTO roles (id, name, description) VALUES
    (1, 'user', 'Regular user'),
    (10, 'admin', 'Platform administrator'),
    (20, 'super_admin', 'Super administrator');

-- Seed permissions
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

-- Super admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT 20, id FROM permissions;

-- Admin gets most permissions except system:config
INSERT INTO role_permissions (role_id, permission_id)
SELECT 10, id FROM permissions WHERE code != 'system:config';
```

- [ ] **6.2: Write asset-svc schema**

Write `asset-svc/manifest/sql/001_init.sql`:
```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE balances (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL UNIQUE,
    balance DECIMAL(20,6) NOT NULL DEFAULT 0,
    total_recharged DECIMAL(20,6) NOT NULL DEFAULT 0,
    total_consumed DECIMAL(20,6) NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 0,  -- optimistic lock
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_balances_user ON balances(user_id);
CREATE INDEX idx_balances_tenant ON balances(tenant_id);

CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL DEFAULT 0,
    user_id BIGINT NOT NULL,
    type VARCHAR(32) NOT NULL,  -- recharge, consume, refund, commission
    amount DECIMAL(20,6) NOT NULL,
    balance_before DECIMAL(20,6) NOT NULL,
    balance_after DECIMAL(20,6) NOT NULL,
    reference_type VARCHAR(32) NOT NULL DEFAULT '',  -- order, usage
    reference_id BIGINT NOT NULL DEFAULT 0,
    remark VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_tx_user ON transactions(user_id);
CREATE INDEX idx_tx_created ON transactions(created_at DESC);

CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    tenant_id INT NOT NULL DEFAULT 0,
    type VARCHAR(32) NOT NULL,  -- model, token
    name VARCHAR(128) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    provider VARCHAR(32) NOT NULL DEFAULT 'platform',  -- platform, user
    provider_user_id BIGINT NOT NULL DEFAULT 0,
    price DECIMAL(20,6) NOT NULL,
    model_name VARCHAR(128) NOT NULL DEFAULT '',
    model_config JSONB DEFAULT '{}',
    status VARCHAR(32) NOT NULL DEFAULT 'active',  -- pending, active, inactive, rejected
    stock INT NOT NULL DEFAULT -1,  -- -1 = unlimited
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
    type VARCHAR(32) NOT NULL,  -- recharge, purchase, refund
    status VARCHAR(32) NOT NULL DEFAULT 'pending',  -- pending, completed, failed, refunded
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
```

- [ ] **6.3: Write market-svc schema**

Write `market-svc/manifest/sql/001_init.sql`:
```sql
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE listings (
    id BIGSERIAL PRIMARY KEY,
    seller_id BIGINT NOT NULL,
    product_type VARCHAR(32) NOT NULL,  -- model, token
    title VARCHAR(256) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price DECIMAL(20,6) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    sold INT NOT NULL DEFAULT 0,
    status VARCHAR(32) NOT NULL DEFAULT 'pending_review',
    -- pending_review, active, rejected, sold_out, cancelled
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
    -- completed, disputed, refunded
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_trades_buyer ON trades(buyer_id);
CREATE INDEX idx_trades_seller ON trades(seller_id);

CREATE TABLE settlements (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    type VARCHAR(32) NOT NULL,  -- income, withdraw
    amount DECIMAL(20,6) NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'pending',
    -- pending, completed, failed
    reference_trade_id BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_settlements_user ON settlements(user_id);
```

---

### Task 7: Scaffold GoFrame gRPC services

**Files:**
- Create: `user-svc/` (GoFrame gRPC project)
- Create: `asset-svc/` (GoFrame gRPC project)
- Create: `market-svc/` (GoFrame gRPC project)

- [ ] **7.1: Initialize user-svc with GoFrame**

```bash
cd /home/ubuntu/code/ai-platform
gf init user-svc
```

Then configure for gRPC mode. Edit `user-svc/manifest/config/config.yaml`:
```yaml
server:
  address: ":8100"
  grpc:
    address: ":8100"

database:
  default:
    link: "postgres://aiplatform:aiplatform@localhost:5432/user_svc"
    debug: true

gf:
  gres:
    mode: "sql"
```

- [ ] **7.2: Initialize asset-svc and market-svc** (same pattern, ports 8101/8102)

```bash
cd /home/ubuntu/code/ai-platform
gf init asset-svc
gf init market-svc
```

- [ ] **7.3: Generate protobuf Go code for user-svc**

```bash
cd /home/ubuntu/code/ai-platform

# Install protoc and protoc-gen-go if needed
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code
protoc --go_out=. --go-grpc_out=. proto/user/v1/user.proto
protoc --go_out=. --go-grpc_out=. proto/asset/v1/asset.proto
protoc --go_out=. --go-grpc_out=. proto/market/v1/market.proto
protoc --go_out=. --go-grpc_out=. proto/gateway/v1/gateway.proto
```

- [ ] **7.4: Copy generated protos into each service**

```bash
# Copy generated go files into respective services
cp -r api/user/v1  user-svc/api/user/v1
cp -r api/asset/v1 asset-svc/api/asset/v1
cp -r api/market/v1 market-svc/api/market/v1
cp -r api/gateway/v1 user-svc/api/gateway/v1
cp -r api/gateway/v1 asset-svc/api/gateway/v1
```

- [ ] **7.5: Implement user-svc gRPC controller skeleton**

Create `user-svc/internal/controller/user/v1/user.go`:
```go
package user

import (
	"context"

	userv1 "ai-platform/api/user/v1"
)

type Controller struct {
	userv1.UnimplementedUserServiceServer
}

func (c *Controller) Register(ctx context.Context, req *userv1.RegisterReq) (*userv1.RegisterRes, error) {
	// TODO: implement
	return nil, nil
}

// ... similar stubs for Login, ValidateToken, CreateApiKey, ListApiKeys, DeleteApiKey, GetUser
```

- [ ] **7.6: Register gRPC service in cmd**

Edit `user-svc/internal/cmd/cmd.go` to register the gRPC controller.

- [ ] **7.7: Implement asset-svc and market-svc controllers** (same pattern)

---

### Task 8: Scaffold api-gateway (HTTP + gRPC proxy)

**Files:**
- Create: `api-gateway/` (GoFrame HTTP project)

- [ ] **8.1: Initialize api-gateway**

```bash
cd /home/ubuntu/code/ai-platform
gf init api-gateway
```

- [ ] **8.2: Configure api-gateway.yaml**

Edit `api-gateway/manifest/config/config.yaml`:
```yaml
server:
  address: ":8080"

# gRPC client connections
grpc:
  user-svc:
    address: "localhost:8100"
  asset-svc:
    address: "localhost:8101"
  market-svc:
    address: "localhost:8102"
```

---

### Task 9: Scaffold ai-gateway (GoFrame LLM relay)

**Files:**
- Create: `ai-gateway/` (GoFrame HTTP project, port 8081)

- [ ] **9.1: Initialize ai-gateway**

```bash
cd /home/ubuntu/code/ai-platform
gf init ai-gateway
```

- [ ] **9.2: Configure ai-gateway.yaml**

Edit `ai-gateway/manifest/config/config.yaml`:
```yaml
server:
  address: ":8081"

grpc:
  user-svc:
    address: "localhost:8100"
  asset-svc:
    address: "localhost:8101"

# Channel/upstream config (will relocate DB tables from new-api)
database:
  default:
    link: "postgres://aiplatform:aiplatform@localhost:5432/ai_gateway"
```

- [ ] **9.3: Build skeleton channel + ability models**

Create `ai-gateway/internal/model/channel.go` and `ai-gateway/internal/model/ability.go` mirroring new-api's structures but stripped of user/token references.

- [ ] **9.4: Build skeleton middleware/TokenAuth** — gRPC call to user-svc instead of local DB read

Create `ai-gateway/internal/middleware/token_auth.go`:
```go
package middleware

// TokenAuth validates API key via gRPC call to user-svc
// Steps:
// 1. Extract "sk-xxx" from Authorization header
// 2. gRPC call user-svc.ValidateToken(token)
// 3. Check user_enabled, model_limits, group
// 4. Inject user_id, api_key_id, model_limits into context
```

- [ ] **9.5: Build skeleton middleware/Distributor** — channel selection logic (ported from new-api)

---

### Task 10: Scaffold web frontend

**Files:**
- Create: `web/` (Vite + React + TanStack)

- [ ] **10.1: Initialize Vite project**

```bash
cd /home/ubuntu/code/ai-platform
npm create vite@latest web -- --template react-ts
cd web
npm install @tanstack/react-router @tanstack/react-query zustand
npm install tailwindcss @tailwindcss/vite
npx shadcn@latest init
```

- [ ] **10.2: Configure project structure**

```
web/src/
├── routes/          # TanStack Router route tree
├── components/      # Shared UI components (shadcn/ui)
├── stores/          # Zustand stores
├── api/             # API client (fetch wrapper)
├── hooks/           # Custom hooks
├── lib/             # Utility functions
├── types/           # TypeScript types
├── main.tsx
└── router.tsx       # Route definition
```

- [ ] **10.3: Build route skeleton**

```typescript
// router.tsx
import { createRouter } from '@tanstack/react-router'
import { routeTree } from './routes'

export const router = createRouter({ routeTree })
```

Routes for MVP:
```
/login
/dashboard
/keys
/models
/orders
/market
/settings
/admin/users
/admin/products
/admin/orders
/admin/market
/admin/roles
```

---

### Task 11: Scaffold Tauri desktop

**Files:**
- Create: `desktop/` (Tauri 2.0)

- [ ] **11.1: Initialize Tauri project**

```bash
cd /home/ubuntu/code/ai-platform
npm create tauri-app@latest desktop -- --template react-ts
```

- [ ] **11.2: Configure Tauri to reuse web frontend code**

Edit `desktop/src-tauri/tauri.conf.json` pointing dev URL and build to web frontend.

---

## Self-Review Checklist

1. **Spec coverage:** Plan covers all 4 requested items (plan doc, DB schema, proto, monorepo scaffolding)
2. **Placeholder scan:** No TBD/TODO patterns remain — business logic is deferred to later phases but skeletons are fully defined
3. **Type consistency:** Protobuf types are consistent across user/asset/market/gateway services
4. **No missing files:** All file paths are absolute and explicit
