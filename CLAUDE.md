# AI Platform — CLAUDE.md

## Project Overview

AI capability platform monorepo. 6 Go services + React SPA + Tauri desktop app.

## Project Structure

```
ai-platform/
├── server/               ← Go services + shared infrastructure
│   ├── api/              ← Shared proto module (.pb.go)
│   ├── user-svc/         ← gRPC :8100
│   ├── asset-svc/        ← gRPC :8101
│   ├── market-svc/       ← gRPC :8102
│   ├── api-gateway/      ← HTTP :8080
│   ├── ai-gateway/       ← HTTP :8081
│   ├── proto/            ← Proto source definitions
│   ├── scripts/          ← Database init scripts
│   └── docker/           ← Dockerfiles, compose, entrypoint, migrations
├── app/
│   ├── web/              ← React SPA (Vite)
│   └── desktop/          ← Tauri desktop app
├── CLAUDE.md
└── .gitignore
```

## Module Dependency

```
server/api/     ← shared proto module (compiled .pb.go files)
│
├── server/user-svc    → replace api => ../api
├── server/asset-svc   → replace api => ../api
├── server/market-svc  → replace api => ../api
├── server/api-gateway → replace api => ../api
└── server/ai-gateway  → replace api => ../api
```

Each service is an independent Go module with its own `go.mod`. Proto-generated code lives in `server/api/` and is shared via `replace` directives.

## Proto Definitions

### Source Organization

```
server/proto/
├── user/v1/user.proto        → go_package "ai-platform/api/user/v1;userv1"
├── asset/v1/asset.proto      → go_package "ai-platform/api/asset/v1;assetv1"
├── market/v1/market.proto    → go_package "ai-platform/api/market/v1;marketv1"
└── gateway/v1/gateway.proto   → go_package "ai-platform/api/gateway/v1;gatewayv1"
```

Compiled output lands in `api/<svc>/v1/*.pb.go` (shared Go module). All services import from here — never copy proto files into service directories.

### Service RPC Definitions

**user.v1.UserService** (port 8100, gRPC server — `user-svc` owns this)
```protobuf
rpc Register(RegisterReq) returns (RegisterRes);
rpc Login(LoginReq) returns (LoginRes);
rpc ValidateToken(ValidateTokenReq) returns (ValidateTokenRes);
rpc CreateApiKey(CreateApiKeyReq) returns (CreateApiKeyRes);
rpc ListApiKeys(ListApiKeysReq) returns (ListApiKeysRes);
rpc DeleteApiKey(DeleteApiKeyReq) returns (DeleteApiKeyRes);
rpc GetUser(GetUserReq) returns (GetUserRes);
```
**Import:** `userv1 "api/user/v1"`
- Server: user-svc implements all RPCs
- Client: api-gateway (register/login/profile/keys), ai-gateway (ValidateToken)

**asset.v1.AssetService** (port 8101, gRPC server — `asset-svc` owns this)
```protobuf
rpc CreateOrder(CreateOrderReq) returns (CreateOrderRes);
rpc GetBalance(GetBalanceReq) returns (GetBalanceRes);
rpc ReportUsage(ReportUsageReq) returns (ReportUsageRes);
rpc ListTransactions(ListTransactionsReq) returns (ListTransactionsRes);
```
**Import:** `assetv1 "api/asset/v1"`
- Server: asset-svc implements all RPCs
- Client: api-gateway (balance/transactions), user-svc (GetBalance on registration), ai-gateway (ReportUsage after LLM calls)

**market.v1.MarketService** (port 8102, gRPC server — `market-svc` owns this)
```protobuf
rpc CreateListing(CreateListingReq) returns (CreateListingRes);
rpc ListListings(ListListingsReq) returns (ListListingsRes);
rpc BuyProduct(BuyProductReq) returns (BuyProductRes);
rpc ListTrades(ListTradesReq) returns (ListTradesRes);
```
**Import:** `marketv1 "api/market/v1"`

**gateway.v1.GatewayService** (ai-gateway internal RPC)
```protobuf
rpc ValidateKey(ValidateKeyReq) returns (ValidateKeyRes);
rpc ReportUsage(ReportUsageReq) returns (ReportUsageRes);
```
**Import:** `gatewayv1 "api/gateway/v1"`

### Cross-Service Dependency Map

```
user-svc  (8100)  ──gRPC──▶  asset-svc  (GetBalance → auto-create balance)
api-gateway (8080) ──gRPC──▶  user-svc   (register/login/profile/keys)
api-gateway (8080) ──gRPC──▶  asset-svc  (balance/transactions)
ai-gateway  (8081) ──gRPC──▶  user-svc   (ValidateToken/ValidateApiKey)
ai-gateway  (8081) ──gRPC──▶  asset-svc  (ReportUsage after LLM calls)
```

### Import Naming Convention

| proto source | Go package | Import path | Alias |
|---|---|---|---|
| `user/v1/user.proto` | `userv1` | `"api/user/v1"` | `userv1` |
| `asset/v1/asset.proto` | `assetv1` | `"api/asset/v1"` | `assetv1` |
| `market/v1/market.proto` | `marketv1` | `"api/market/v1"` | `marketv1` |
| `gateway/v1/gateway.proto` | `gatewayv1` | `"api/gateway/v1"` | `gatewayv1` |

In code, always use the proto-shortened import alias to avoid conflicts with GoFrame API structs in api-gateway:

```go
import (
    assetv1 "api/asset/v1"    // proto types (messages, client, server interfaces)
    userv1 "api/user/v1"
)

// Register gRPC server
assetv1.RegisterAssetServiceServer(s, &asset.Controller{svc: svc})

// Call gRPC client
pbRes, err := grpcclient.UserSvc.GetUser(ctx, &userv1.GetUserReq{UserId: id})
```

### Adding a New Proto

1. Create `server/proto/<svc>/v1/<svc>.proto` with `package <svc>.v1` and `go_package "ai-platform/api/<svc>/v1;<svc>v1"`
2. Run protoc from repo root:
   ```bash
   protoc --go_out=. --go-grpc_out=. server/proto/<svc>/v1/<svc>.proto
   ```
3. Add database setup in `scripts/init-db.sh` if new service
4. Import in any service via `api/<svc>/v1`

### Updating an Existing Proto

1. Edit `server/proto/<svc>/v1/<svc>.proto`
2. Re-generate:
   ```bash
   protoc --go_out=. --go-grpc_out=. server/proto/<svc>/v1/<svc>.proto
   ```
3. Rebuild dependent services:
   ```bash
   cd server/<svc> && go build ./...
   ```

## Development Commands

```bash
# Go services (all paths relative to server/)
cd server/<service-dir> && go build ./...
cd server/<service-dir> && go run .          # start service

# Web frontend
cd app/web && npm install && npm run dev     # start Vite dev server

# Database (PostgreSQL 16)
bash server/scripts/init-db.sh               # create databases (local)
```

## Docker

Dockerfiles are in `server/docker/`. Per-service compose files in each service root for independent deployment.

**Config approach:** Each Docker image bakes the original `manifest/config/config.yaml`. At container startup, `docker/entrypoint.sh` substitutes hostnames from environment variables — no config copies needed.

```bash
# One-click build & start (top-level, all services)
cd server/docker && bash start.sh

# Per-service (standalone, each includes its own postgres):
cd server/user-svc && docker compose up -d       # postgres + user-svc + asset-svc
cd server/asset-svc && docker compose up -d      # postgres + asset-svc
cd server/market-svc && docker compose up -d     # postgres + market-svc
cd server/api-gateway && docker compose up -d    # postgres + user-svc + asset-svc + api-gateway
cd server/ai-gateway && docker compose up -d     # postgres + user-svc + asset-svc + ai-gateway
```

**Env vars available** (set in docker-compose `environment:`):
| Variable | Default | Override in Docker |
|----------|---------|-------------------|
| `DB_HOST` | `localhost` → `postgres` | Database server hostname |
| `USER_SVC_ADDR` | `localhost:8100` → `user-svc:8100` | user-svc gRPC address |
| `ASSET_SVC_ADDR` | `localhost:8101` → `asset-svc:8101` | asset-svc gRPC address |

## Database Migrations

Uses [Goose](https://github.com/pressly/goose) for SQL migrations. Migration files in `server/<svc>/migrations/`.

```bash
# Install goose CLI
go install github.com/pressly/goose/v3/cmd/goose@latest

# Run all migrations (needs postgres running)
cd server/docker && bash migrate.sh

# Or per service:
goose -dir server/user-svc/migrations postgres "postgres://aiplatform:aiplatform@localhost:5432/user_svc?sslmode=disable" up
```

Available migrations:
- `user-svc/migrations/` — users, api_keys
- `asset-svc/migrations/` — balances, transactions, usage_records, orders
- `ai-gateway/migrations/` — channels, abilities
- `market-svc/migrations/` — (empty, pending implementation)

## DAO Code Generation

GoFrame `gf gen dao` generates type-safe DAO/DO/Entity code from database tables. Each service's `server/<svc>/hack/config.yaml` now points to PostgreSQL.

**Workflow:**
1. Start PostgreSQL, run migrations (tables must exist)
2. Generate code:
   ```bash
   cd server/user-svc && gf gen dao
   ```
3. Output: `server/<svc>/internal/dao/`, `server/<svc>/internal/model/do/`, `server/<svc>/internal/model/entity/`

After generation, you can migrate from `g.DB().Model("table")` to `dao.Table.Ctx(ctx)` for type-safe DB access.

## Code Conventions

### Go Services

- **Framework**: GoFrame v2.7.1 (`github.com/gogf/gf/v2`)
- **DB access**: `g.DB().Model("<table>").Ctx(ctx)` — global DB instance from config
- **Config**: `g.Cfg().MustGet(ctx, "key")` — reads YAML from `manifest/config/config.yaml`
- **Logging**: `g.Log().Info(ctx, ...)` (api-gateway/ai-gateway) or `glog.Printf(ctx, ...)` (gRPC services)
- **HTTP context**: `g.RequestFromCtx(ctx)` to get request vars like `user_id`
- **Errors**: Use `gerror.NewCode(gcode.CodeNotAuthorized, ...)` for API errors

**Controller pattern** — thin delegation to service layer:
```go
type Controller struct {
    svc *service.SomeService
    assetv1.UnimplementedAssetServiceServer   // embed for gRPC
}
func New(svc *service.SomeService) *Controller { return &Controller{svc: svc} }
```

**Service pattern** — business logic:
```go
type SomeService struct{}
func (s *SomeService) Method(ctx, req) (res, err) {
    // use g.DB() for direct database access
}
```

**Import aliases for shared api:**
```go
userv1 "api/user/v1"          // user proto types
assetv1 "api/asset/v1"        // asset proto types
marketv1 "api/market/v1"      // market proto types
gatewayv1 "api/gateway/v1"    // gateway proto types
```

### gRPC Client Pattern (cross-service calls)

```go
// internal/grpcclient/client.go
package grpcclient
var UserSvc userpb.UserServiceClient  // populated in cmd.go initGrpcClients

// Usage in service layer:
grpcclient.UserSvc.SomeMethod(ctx, &userpb.SomeReq{...})
```

### Frontend (React + TypeScript)

- **Framework**: React 19, TanStack Router (file-based routing), TanStack Query
- **State**: Zustand with localStorage persistence for auth
- **Styling**: Tailwind CSS v4 + `class-variance-authority`
- **API**: `apiPost<T>(path, body, token?)` / `apiGet<T>(path, token)` in `src/api/client.ts`
- **Route files**: `src/routes/<name>.tsx` with `createFileRoute('/<name>')`
- **API base**: `http://localhost:8080/api/v1`

### Auth Flow

1. Login/Register → user-svc → returns JWT (HS256, 24h expiry)
2. Web stores JWT in `localStorage` via Zustand
3. api-gateway validates JWT in middleware, injects `user_id`/`username`/`role` into context
4. ai-gateway validates API keys (sk-xxx prefix) via user-svc gRPC

### Database

PostgreSQL 16. One database per service:
- `user_svc` — users, api_keys
- `asset_svc` — balances, transactions, usage_records, products, orders, order_items
- `market_svc` — listings, reviews
- `ai_gateway` — model configs, channels

**Optimistic locking** for balance updates:
```go
// 3-retry loop with version column
balance:   current - quota,
version:   gdb.Raw("version + 1"),
total_consumed: gdb.Raw("total_consumed + " + quotaStr),
```

## Key Conventions

| Rule | Why |
|------|-----|
| Use `gcode.CodeNotAuthorized` (not `CodeUnauthorized`) | GoFrame v2.7.1 constant |
| All gRPC clients in dedicated `internal/grpcclient` package | Consistent dependency pattern |
| Controllers embed `UnimplementedXxxServiceServer` | gRPC forward compatibility |
| Proto imports via `api/<svc>/v1`, never local copies | Single source of truth |
| Use `strings.Join` not `g.FormatFloat` for SQL literals | g.FormatFloat doesn't exist in GoFrame v2 |

---

# Behavioral Guidelines

Behavioral guidelines to reduce common LLM coding mistakes. Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed. For trivial tasks, use judgment.

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- No error handling for impossible scenarios.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior engineer say this is overcomplicated?" If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- If you notice unrelated dead code, mention it - don't delete it.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure tests pass before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently. Weak criteria ("make it work") require constant clarification.

---

**These guidelines are working if:** fewer unnecessary changes in diffs, fewer rewrites due to overcomplication, and clarifying questions come before implementation rather than after mistakes.
