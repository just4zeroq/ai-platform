# AI Platform — CLAUDE.md

## Project Overview

AI capability platform monorepo. 6 Go services + React SPA + Tauri desktop app.

## Architecture

```
Service          Port   Type     Description
─────────────────────────────────────────────────
user-svc         8100   gRPC     User auth (register/login/JWT), API key CRUD
asset-svc        8101   gRPC     Balance, transactions, usage reporting, orders
market-svc       8102   gRPC     Product listings, marketplace
api-gateway      8080   HTTP     Public-facing HTTP API, proxies to gRPC services
ai-gateway       8081   HTTP     LLM proxy gateway, token validation, model routing
web              5173   Vite     React SPA (TanStack Router + Zustand)
desktop          —      Tauri    Electron alternative
```

## Module Dependency

```
api/            ← shared proto module (compiled .pb.go files)
│
├── user-svc    → replace api => ../api    (server: api/user/v1, client: api/asset/v1)
├── asset-svc   → replace api => ../api    (server: api/asset/v1)
├── market-svc  → replace api => ../api    (server: api/market/v1)
├── api-gateway → replace api => ../api    (client: api/user/v1 + api/asset/v1)
└── ai-gateway  → replace api => ../api    (client: api/user/v1)
```

Each service is an independent Go module with its own `go.mod`. Proto-generated code lives in `api/` and is shared via `replace` directives.

## Proto Definitions

### Source Organization

```
proto/
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

1. Create `proto/<svc>/v1/<svc>.proto` with `package <svc>.v1` and `go_package "ai-platform/api/<svc>/v1;<svc>v1"`
2. Run protoc from repo root:
   ```bash
   protoc --go_out=. --go-grpc_out=. proto/<svc>/v1/<svc>.proto
   ```
3. Add database setup in `scripts/init-db.sh` if new service
4. Import in any service via `api/<svc>/v1`

### Updating an Existing Proto

1. Edit `proto/<svc>/v1/<svc>.proto`
2. Re-generate:
   ```bash
   protoc --go_out=. --go-grpc_out=. proto/<svc>/v1/<svc>.proto
   ```
3. Rebuild all dependent services — no file copying needed:
   ```bash
   go build ./...             # all services that import api/<svc>/v1
   ```

## Development Commands

```bash
# Go services
cd <service-dir> && go build ./...
cd <service-dir> && go run .                # start service

# Web frontend
cd web && npm install && npm run dev         # start Vite dev server

# Database (PostgreSQL 16)
bash scripts/init-db.sh                      # create databases
```

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
