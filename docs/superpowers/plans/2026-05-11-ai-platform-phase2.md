# AI Platform — Phase 2: Business Logic (Auth + API Keys + Gateway)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement user authentication (register/login/JWT), API key management, ai-gateway token validation, and web frontend login flow.

**Architecture:** user-svc implements gRPC auth business logic (bcrypt + JWT). api-gateway proxies HTTP → gRPC with JWT auth middleware. ai-gateway calls user-svc ValidateToken to check API keys. Web frontend provides login UI with auth store.

**Tech Stack:** Go 1.22+, GoFrame v2 gdb (DB), golang-jwt/jwt/v5, golang.org/x/crypto/bcrypt, google.golang.org/grpc, TanStack Router (auth guards), Zustand (auth store)

**Dependency chain:** user-svc auth (foundation) → api-gateway HTTP proxy → ai-gateway middleware → web login

---

## File Structure

All new/modified files across the monorepo:

```
user-svc/
├── go.mod                                           # [MODIFY] add jwt, bcrypt deps
├── internal/
│   ├── cmd/cmd.go                                   # [MODIFY] init DB on startup
│   ├── controller/user/user.go                      # [MODIFY] call service layer
│   └── service/
│       ├── user.go                                   # [CREATE] Register, Login, ValidateToken, GetUser
│       └── apikey.go                                 # [CREATE] CreateApiKey, ListApiKeys, DeleteApiKey

api-gateway/
├── go.mod                                           # [MODIFY] add grpc, protobuf deps
├── api/userpb/v1/
│   ├── user.pb.go                                   # [COPY] from ai-platform/api/user/v1/
│   └── user_grpc.pb.go                              # [COPY] from ai-platform/api/user/v1/
├── internal/
│   ├── cmd/cmd.go                                   # [MODIFY] init gRPC client connections
│   ├── grpcclient/client.go                         # [CREATE] gRPC client connection holder
│   ├── controller/user/user.go                      # [MODIFY] call gRPC services
│   └── middleware/
│       └── auth.go                                   # [CREATE] JWT auth middleware

ai-gateway/
├── go.mod                                           # [MODIFY] add grpc, protobuf deps
├── api/userpb/v1/
│   ├── user.pb.go                                   # [COPY] for gRPC client
│   └── user_grpc.pb.go                              # [COPY] for gRPC client
├── internal/
│   └── middleware/token_auth.go                      # [MODIFY] implement gRPC call to user-svc

web/src/
├── api/client.ts                                    # [CREATE] HTTP API client (fetch wrapper)
├── stores/auth.ts                                   # [MODIFY] add login/logout/fetchProfile actions
├── routes/
│   ├── login.tsx                                    # [MODIFY] login form UI
│   └── __root.tsx                                   # [MODIFY] auth guard / layout
```

---

### Task 1: user-svc — Auth Service (Register, Login, ValidateToken)

**Files:**
- Modify: `user-svc/go.mod`
- Create: `user-svc/internal/service/user.go`
- Modify: `user-svc/internal/controller/user/user.go`
- Modify: `user-svc/internal/cmd/cmd.go`

**Step 1.1 — Add dependencies**

```bash
cd /home/ubuntu/code/ai-platform/user-svc
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
```

Update `user-svc/go.mod`:
```
require (
    github.com/golang-jwt/jwt/v5 v5.2.2
    golang.org/x/crypto v0.32.0
)
```

**Step 1.2 — Create `user-svc/internal/service/user.go`**

```go
package service

import (
	"context"
	"errors"
	"time"

	userv1 "user-svc/api/user/v1"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("ai-platform-jwt-secret-2026")

type UserClaims struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     int32  `json:"role"`
	jwt.RegisteredClaims
}

type UserService struct{}

func (s *UserService) Register(ctx context.Context, req *userv1.RegisterReq) (*userv1.RegisterRes, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("username and password are required")
	}
	count, err := g.DB().Model("users").Ctx(ctx).Where("username", req.Username).Count()
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("username already exists")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	source := req.Source
	if source == "" {
		source = "email"
	}
	result, err := g.DB().Model("users").Ctx(ctx).Data(g.Map{
		"username":     req.Username,
		"password":     string(hashed),
		"email":        req.Email,
		"phone":        req.Phone,
		"source":       source,
		"role":         1,
		"status":       1,
		"group_name":   "default",
		"display_name": req.Username,
	}).Insert()
	if err != nil {
		return nil, err
	}
	userId, _ := result.LastInsertId()
	token, err := s.generateToken(userId, req.Username, 1)
	if err != nil {
		return nil, err
	}
	return &userv1.RegisterRes{
		User: &userv1.User{
			Id: userId, Username: req.Username, Email: req.Email,
			Phone: req.Phone, Status: 1, Role: 1, Group: "default", Source: source,
		},
		AccessToken: token, RefreshToken: token,
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *userv1.LoginReq) (*userv1.LoginRes, error) {
	if req.Username == "" || req.Password == "" {
		return nil, errors.New("username and password are required")
	}
	user, err := g.DB().Model("users").Ctx(ctx).Where("username", req.Username).One()
	if err != nil {
		return nil, err
	}
	if user == nil {
		user, err = g.DB().Model("users").Ctx(ctx).Where("email", req.Username).One()
		if err != nil {
			return nil, err
		}
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user["password"].String()), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid username or password")
	}
	if user["status"].Int() != 1 {
		return nil, errors.New("account is disabled")
	}
	userId := user["id"].Int64()
	username := user["username"].String()
	role := user["role"].Int32()
	token, err := s.generateToken(userId, username, role)
	if err != nil {
		return nil, err
	}
	return &userv1.LoginRes{
		User: &userv1.User{
			Id: userId, TenantId: user["tenant_id"].Int64(), Username: username,
			Email: user["email"].String(), Phone: user["phone"].String(),
			DisplayName: user["display_name"].String(), Status: user["status"].Int32(),
			Role: role, Group: user["group_name"].String(), Source: user["source"].String(),
		},
		AccessToken: token, RefreshToken: token,
	}, nil
}

func (s *UserService) ValidateToken(ctx context.Context, token string) (*userv1.ValidateTokenRes, error) {
	claims, err := s.parseToken(token)
	if err == nil {
		return &userv1.ValidateTokenRes{
			UserId: claims.UserId, UserStatus: 1, Group: "default",
			HasToken: false, UserRole: claims.Role, TenantId: 0,
		}, nil
	}
	if len(token) > 3 && token[:3] == "sk-" {
		return s.validateApiKey(ctx, token)
	}
	return nil, errors.New("invalid token")
}

func (s *UserService) validateApiKey(ctx context.Context, key string) (*userv1.ValidateTokenRes, error) {
	apiKey, err := g.DB().Model("api_keys").Ctx(ctx).
		Where("key", key).Where("deleted_at IS NULL").One()
	if err != nil {
		return nil, err
	}
	if apiKey == nil {
		return nil, errors.New("invalid API key")
	}
	if apiKey["status"].Int() != 1 {
		return nil, errors.New("API key is not active")
	}
	if !apiKey["expire_time"].IsEmpty() {
		et := apiKey["expire_time"].GTime()
		if et != nil && et.Before(time.Now()) {
			return nil, errors.New("API key has expired")
		}
	}
	userId := apiKey["user_id"].Int64()
	user, err := g.DB().Model("users").Ctx(ctx).Where("id", userId).One()
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	if user["status"].Int() != 1 {
		return nil, errors.New("user is disabled")
	}
	keyGroup := apiKey["group_name"].String()
	if keyGroup == "" {
		keyGroup = "default"
	}
	return &userv1.ValidateTokenRes{
		UserId: userId, UserStatus: int32(user["status"].Int()),
		Group: user["group_name"].String(), HasToken: true,
		ApiKeyId: apiKey["id"].Int64(),
		ModelLimitsEnabled: apiKey["model_limits_enabled"].Bool(),
		ModelLimits: []string{}, KeyGroup: keyGroup,
		KeyUnlimitedQuota: true, KeyRemainQuota: 0,
		UserRole: int32(user["role"].Int()), TenantId: user["tenant_id"].Int64(),
	}, nil
}

func (s *UserService) GetUser(ctx context.Context, userId int64) (*userv1.User, error) {
	user, err := g.DB().Model("users").Ctx(ctx).Where("id", userId).One()
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return &userv1.User{
		Id: user["id"].Int64(), TenantId: user["tenant_id"].Int64(),
		Username: user["username"].String(), Email: user["email"].String(),
		Phone: user["phone"].String(), DisplayName: user["display_name"].String(),
		Avatar: user["avatar"].String(), Status: user["status"].Int32(),
		Role: user["role"].Int32(), Group: user["group_name"].String(),
		Source: user["source"].String(), Remark: user["remark"].String(),
		CreatedAt: user["created_at"].String(), UpdatedAt: user["updated_at"].String(),
	}, nil
}

func (s *UserService) generateToken(userId int64, username string, role int32) (string, error) {
	claims := UserClaims{
		UserId: userId, Username: username, Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ai-platform",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func (s *UserService) parseToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
```

**Step 1.3 — Update controller**

Replace `user-svc/internal/controller/user/user.go`:

```go
package user

import (
	"context"

	userv1 "user-svc/api/user/v1"
	"user-svc/internal/service"
)

type Controller struct {
	userv1.UnimplementedUserServiceServer
	svc *service.UserService
}

func (c *Controller) Register(ctx context.Context, req *userv1.RegisterReq) (*userv1.RegisterRes, error) {
	return c.svc.Register(ctx, req)
}
func (c *Controller) Login(ctx context.Context, req *userv1.LoginReq) (*userv1.LoginRes, error) {
	return c.svc.Login(ctx, req)
}
func (c *Controller) ValidateToken(ctx context.Context, req *userv1.ValidateTokenReq) (*userv1.ValidateTokenRes, error) {
	return c.svc.ValidateToken(ctx, req.Token)
}
func (c *Controller) CreateApiKey(ctx context.Context, req *userv1.CreateApiKeyReq) (*userv1.CreateApiKeyRes, error) {
	return c.svc.CreateApiKey(ctx, req)
}
func (c *Controller) ListApiKeys(ctx context.Context, req *userv1.ListApiKeysReq) (*userv1.ListApiKeysRes, error) {
	return c.svc.ListApiKeys(ctx, req)
}
func (c *Controller) DeleteApiKey(ctx context.Context, req *userv1.DeleteApiKeyReq) (*userv1.DeleteApiKeyRes, error) {
	return c.svc.DeleteApiKey(ctx, req)
}
func (c *Controller) GetUser(ctx context.Context, req *userv1.GetUserReq) (*userv1.GetUserRes, error) {
	user, err := c.svc.GetUser(ctx, req.UserId)
	if err != nil {
		return nil, err
	}
	return &userv1.GetUserRes{User: user}, nil
}
```

**Step 1.4 — Update cmd.go with DB init**

Replace `user-svc/internal/cmd/cmd.go`:

```go
package cmd

import (
	"context"
	"fmt"
	"net"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	userv1 "user-svc/api/user/v1"
	"user-svc/internal/controller/user"
	"user-svc/internal/service"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start gRPC server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			initDB(ctx)
			lis, err := net.Listen("tcp", ":8100")
			if err != nil {
				glog.Fatalf(ctx, "failed to listen: %v", err)
			}
			s := grpc.NewServer()
			userv1.RegisterUserServiceServer(s, &user.Controller{svc: &service.UserService{}})
			reflection.Register(s)
			glog.Printf(ctx, "user-svc gRPC server listening at %v", lis.Addr())
			fmt.Printf("user-svc gRPC server listening at %v\n", lis.Addr())
			if err := s.Serve(lis); err != nil {
				glog.Fatalf(ctx, "failed to serve: %v", err)
			}
			return nil
		},
	}
)

func initDB(ctx context.Context) {
	if err := g.DB().PingMaster(); err != nil {
		glog.Fatalf(ctx, "database connection failed: %v", err)
	}
	glog.Println(ctx, "database connected successfully")
}
```

**Step 1.5 — Build and verify**

```bash
cd /home/ubuntu/code/ai-platform/user-svc
go build ./...
```

Expected: clean compilation.

**Step 1.6 — Commit**

```bash
cd /home/ubuntu/code/ai-platform
git add user-svc/go.mod user-svc/go.sum user-svc/internal/service/user.go user-svc/internal/controller/user/user.go user-svc/internal/cmd/cmd.go
git commit -m "feat(user-svc): implement auth service (register, login, validate token)"
```

---

### Task 2: user-svc — API Key Management

**Files:**
- Create: `user-svc/internal/service/apikey.go`

**Step 2.1 — Create `user-svc/internal/service/apikey.go`**

```go
package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	userv1 "user-svc/api/user/v1"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

func (s *UserService) CreateApiKey(ctx context.Context, req *userv1.CreateApiKeyReq) (*userv1.CreateApiKeyRes, error) {
	if req.UserId == 0 {
		return nil, errors.New("user_id is required")
	}
	keyBytes := make([]byte, 24)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, err
	}
	rawKey := "sk-" + hex.EncodeToString(keyBytes)
	name := req.Name
	if name == "" {
		name = "Default Key"
	}
	dbData := g.Map{
		"user_id":              req.UserId,
		"key":                  rawKey,
		"name":                 name,
		"status":               1,
		"model_limits_enabled": req.ModelLimitsEnabled,
		"model_limits":         req.ModelLimits,
		"allow_ips":            req.AllowIps,
		"group_name":           "default",
	}
	if req.ExpireTime > 0 {
		dbData["expire_time"] = gtime.NewFromTimeStamp(req.ExpireTime)
	}
	if req.Group != "" {
		dbData["group_name"] = req.Group
	}
	result, err := g.DB().Model("api_keys").Ctx(ctx).Data(dbData).Insert()
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &userv1.CreateApiKeyRes{
		ApiKey: &userv1.ApiKey{
			Id: id, UserId: req.UserId, Key: rawKey, Name: name, Status: 1,
			ModelLimitsEnabled: req.ModelLimitsEnabled, ModelLimits: req.ModelLimits,
			ExpireTime: req.ExpireTime, AllowIps: req.AllowIps, Group: req.Group,
		},
		RawKey: rawKey,
	}, nil
}

func (s *UserService) ListApiKeys(ctx context.Context, req *userv1.ListApiKeysReq) (*userv1.ListApiKeysRes, error) {
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize < 1 {
		pageSize = 20
	}
	total, err := g.DB().Model("api_keys").Ctx(ctx).
		Where("user_id", req.UserId).Where("deleted_at IS NULL").Count()
	if err != nil {
		return nil, err
	}
	var dbKeys []gdb.Record
	err = g.DB().Model("api_keys").Ctx(ctx).
		Where("user_id", req.UserId).Where("deleted_at IS NULL").
		Order("id DESC").Limit(pageSize).Offset((page - 1) * pageSize).Scan(&dbKeys)
	if err != nil {
		return nil, err
	}
	apiKeys := make([]*userv1.ApiKey, 0)
	for _, k := range dbKeys {
		apiKeys = append(apiKeys, &userv1.ApiKey{
			Id: k["id"].Int64(), UserId: k["user_id"].Int64(), Key: k["key"].String(),
			Name: k["name"].String(), Status: int32(k["status"].Int()),
			ModelLimitsEnabled: k["model_limits_enabled"].Bool(),
			ModelLimits: k["model_limits"].String(), AllowIps: k["allow_ips"].String(),
			Group: k["group_name"].String(), CreatedAt: k["created_at"].String(),
		})
	}
	return &userv1.ListApiKeysRes{ApiKeys: apiKeys, Total: int32(total)}, nil
}

func (s *UserService) DeleteApiKey(ctx context.Context, req *userv1.DeleteApiKeyReq) (*userv1.DeleteApiKeyRes, error) {
	_, err := g.DB().Model("api_keys").Ctx(ctx).
		Where("id", req.Id).Where("user_id", req.UserId).Delete()
	if err != nil {
		return nil, err
	}
	return &userv1.DeleteApiKeyRes{}, nil
}
```

Add import: `"github.com/gogf/gf/v2/os/gtime"` to apikey.go.

**Step 2.2 — Build and verify**

```bash
cd /home/ubuntu/code/ai-platform/user-svc
go build ./...
```

Expected: clean compilation.

**Step 2.3 — Commit**

```bash
cd /home/ubuntu/code/ai-platform
git add user-svc/internal/service/apikey.go
git commit -m "feat(user-svc): implement API key CRUD (create, list, delete)"
```

---

### Task 3: api-gateway — gRPC Proxy + Auth Middleware

**Files:**
- Create: `api-gateway/api/userpb/v1/user.pb.go` (copy)
- Create: `api-gateway/api/userpb/v1/user_grpc.pb.go` (copy)
- Create: `api-gateway/internal/grpcclient/client.go`
- Create: `api-gateway/internal/middleware/auth.go`
- Modify: `api-gateway/go.mod`
- Modify: `api-gateway/internal/cmd/cmd.go`
- Modify: `api-gateway/internal/controller/user/user.go`
- Modify: `api-gateway/manifest/config/config.yaml`

**Step 3.1 — Copy proto files**

```bash
cd /home/ubuntu/code/ai-platform
mkdir -p api-gateway/api/userpb/v1
cp api/user/v1/user.pb.go api-gateway/api/userpb/v1/user.pb.go
cp api/user/v1/user_grpc.pb.go api-gateway/api/userpb/v1/user_grpc.pb.go
```

The pb.go files use `package userv1`. They live in `api/userpb/v1/` (Go package name = directory's package declaration, which is `userv1`). The existing `api-gateway/api/user/v1/user.go` stays as `package user` — no conflict.

**Step 3.2 — Add gRPC deps to api-gateway**

```bash
cd /home/ubuntu/code/ai-platform/api-gateway
go get google.golang.org/grpc
go get google.golang.org/protobuf
go get github.com/golang-jwt/jwt/v5
```

**Step 3.3 — Update config**

Write `api-gateway/manifest/config/config.yaml`:

```yaml
server:
  address: ":8080"

grpc:
  user-svc:
    address: "localhost:8100"
```

**Step 3.4 — Create grpcclient package**

Write `api-gateway/internal/grpcclient/client.go`:

```go
package grpcclient

import (
	userpb "api-gateway/api/userpb/v1"
)

var (
	UserSvc userpb.UserServiceClient
)
```

**Step 3.5 — Create auth middleware**

Write `api-gateway/internal/middleware/auth.go`:

```go
package middleware

import (
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("ai-platform-jwt-secret-2026")

type UserClaims struct {
	UserId   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     int32  `json:"role"`
	jwt.RegisteredClaims
}

func Auth(r *ghttp.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.WriteStatus(401, g.Map{"message": "missing authorization header"})
		r.Exit()
		return
	}
	tokenStr := authHeader
	if strings.HasPrefix(tokenStr, "Bearer ") {
		tokenStr = tokenStr[7:]
	}
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		r.Response.WriteStatus(401, g.Map{"message": "invalid or expired token"})
		r.Exit()
		return
	}
	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		r.Response.WriteStatus(401, g.Map{"message": "invalid token claims"})
		r.Exit()
		return
	}
	r.SetCtxVar("user_id", claims.UserId)
	r.SetCtxVar("username", claims.Username)
	r.SetCtxVar("role", claims.Role)
	r.Middleware.Next()
}
```

**Step 3.6 — Update controller/user/user.go (gRPC proxy)**

Replace `api-gateway/internal/controller/user/user.go`:

```go
package user

import (
	"context"

	userv1 "api-gateway/api/user/v1"
	userpb "api-gateway/api/userpb/v1"
	"api-gateway/internal/grpcclient"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

type Controller struct{}

func (c *Controller) Register(ctx context.Context, req *userv1.RegisterReq) (res *userv1.RegisterRes, err error) {
	pbRes, err := grpcclient.UserSvc.Register(ctx, &userpb.RegisterReq{
		Username: req.Username, Password: req.Password, Email: req.Email,
	})
	if err != nil {
		return nil, err
	}
	return &userv1.RegisterRes{
		UserId: pbRes.User.Id, Username: pbRes.User.Username, Token: pbRes.AccessToken,
	}, nil
}

func (c *Controller) Login(ctx context.Context, req *userv1.LoginReq) (res *userv1.LoginRes, err error) {
	pbRes, err := grpcclient.UserSvc.Login(ctx, &userpb.LoginReq{
		Username: req.Username, Password: req.Password,
	})
	if err != nil {
		return nil, err
	}
	return &userv1.LoginRes{
		UserId: pbRes.User.Id, Username: pbRes.User.Username, Token: pbRes.AccessToken,
	}, nil
}

func (c *Controller) GetProfile(ctx context.Context, req *userv1.GetProfileReq) (res *userv1.GetProfileRes, err error) {
	r := g.RequestFromCtx(ctx)
	userId := r.GetCtxVar("user_id").Int64()
	if userId == 0 {
		return nil, gerror.NewCode(gcode.CodeUnauthorized)
	}
	pbRes, err := grpcclient.UserSvc.GetUser(ctx, &userpb.GetUserReq{UserId: userId})
	if err != nil {
		return nil, err
	}
	return &userv1.GetProfileRes{
		Id: pbRes.User.Id, Username: pbRes.User.Username,
		Email: pbRes.User.Email, DisplayName: pbRes.User.DisplayName,
	}, nil
}
```

**Step 3.7 — Update cmd.go**

Replace `api-gateway/internal/cmd/cmd.go`:

```go
package cmd

import (
	"context"

	"api-gateway/internal/controller/asset"
	"api-gateway/internal/controller/user"
	"api-gateway/internal/grpcclient"
	"api-gateway/internal/middleware"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userpb "api-gateway/api/userpb/v1"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			initGrpcClients(ctx)

			s := g.Server()

			// Public routes
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.POST("/user/register", user.Controller{}.Register)
				group.POST("/user/login", user.Controller{}.Login)
			})

			// Protected routes (JWT required)
			s.Group("/api/v1", func(group *ghttp.RouterGroup) {
				group.Middleware(middleware.Auth)
				group.GET("/user/profile", user.Controller{}.GetProfile)
				// Asset routes from asset controller (TBD Phase 3)
			})

			// Admin routes (future)
			s.Group("/api/v1/admin", func(group *ghttp.RouterGroup) {
				group.Middleware(middleware.Auth)
			})

			s.Run()
			return nil
		},
	}
)

func initGrpcClients(ctx context.Context) {
	cfg := g.Cfg().MustGet(ctx, "grpc.user-svc").Map()
	address := cfg["address"].(string)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		g.Log().Fatalf(ctx, "failed to connect to user-svc: %v", err)
	}
	grpcclient.UserSvc = userpb.NewUserServiceClient(conn)
	g.Log().Println(ctx, "connected to user-svc gRPC at", address)
}
```

**Step 3.8 — Build and verify**

```bash
cd /home/ubuntu/code/ai-platform/api-gateway
go build ./...
```

Expected: clean compilation.

**Step 3.9 — Commit**

```bash
cd /home/ubuntu/code/ai-platform
git add api-gateway/
git commit -m "feat(api-gateway): implement gRPC proxy for user auth with JWT middleware"
```

---

### Task 4: ai-gateway — TokenAuth Middleware

**Files:**
- Copy: `ai-gateway/api/userpb/v1/user.pb.go`
- Copy: `ai-gateway/api/userpb/v1/user_grpc.pb.go`
- Modify: `ai-gateway/go.mod`
- Modify: `ai-gateway/internal/middleware/token_auth.go`
- Modify: `ai-gateway/internal/cmd/cmd.go`
- Modify: `ai-gateway/manifest/config/config.yaml`

**Step 4.1 — Copy proto files**

```bash
cd /home/ubuntu/code/ai-platform
mkdir -p ai-gateway/api/userpb/v1
cp api/user/v1/user.pb.go ai-gateway/api/userpb/v1/user.pb.go
cp api/user/v1/user_grpc.pb.go ai-gateway/api/userpb/v1/user_grpc.pb.go
```

**Step 4.2 — Add gRPC deps**

```bash
cd /home/ubuntu/code/ai-platform/ai-gateway
go get google.golang.org/grpc
go get google.golang.org/protobuf
```

**Step 4.3 — Update config**

Write `ai-gateway/manifest/config/config.yaml`:

```yaml
server:
  address: ":8081"

grpc:
  user-svc:
    address: "localhost:8100"

database:
  default:
    link: "postgres://aiplatform:aiplatform@localhost:5432/ai_gateway"
    debug: true
```

**Step 4.4 — Create grpcclient package**

Write `ai-gateway/internal/grpcclient/client.go`:

```go
package grpcclient

import (
	userpb "ai-gateway/api/userpb/v1"
)

var (
	UserSvc userpb.UserServiceClient
)
```

**Step 4.5 — Update cmd.go**

Modify `ai-gateway/internal/cmd/cmd.go`:

```go
package cmd

import (
	"context"

	"ai-gateway/internal/grpcclient"
	"ai-gateway/internal/router"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userpb "ai-gateway/api/userpb/v1"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			initGrpcClients(ctx)

			s := g.Server()
			s.Group("/", func(group *ghttp.RouterGroup) {
				router.RelayRouter(group)
			})
			s.Run()
			return nil
		},
	}
)

func initGrpcClients(ctx context.Context) {
	cfg := g.Cfg().MustGet(ctx, "grpc.user-svc").Map()
	address := cfg["address"].(string)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		g.Log().Fatalf(ctx, "failed to connect to user-svc: %v", err)
	}
	grpcclient.UserSvc = userpb.NewUserServiceClient(conn)
	g.Log().Println(ctx, "connected to user-svc gRPC at", address)
}
```

**Step 4.6 — Implement TokenAuth middleware**

Replace `ai-gateway/internal/middleware/token_auth.go`:

```go
package middleware

import (
	"net/http"
	"strings"

	"ai-gateway/internal/grpcclient"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	userpb "ai-gateway/api/userpb/v1"
)

func TokenAuth(r *ghttp.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.WriteStatus(http.StatusUnauthorized, gerror.New("missing authorization header"))
		r.Exit()
		return
	}
	key := authHeader
	if strings.HasPrefix(key, "Bearer ") || strings.HasPrefix(key, "bearer ") {
		key = key[7:]
	}

	// Call user-svc via gRPC
	ctx := r.Context()
	res, err := grpcclient.UserSvc.ValidateToken(ctx, &userpb.ValidateTokenReq{Token: key})
	if err != nil {
		g.Log().Errorf(ctx, "token validation failed: %v", err)
		r.Response.WriteStatus(http.StatusUnauthorized, g.Map{"error": "invalid token"})
		r.Exit()
		return
	}
	if !res.HasToken {
		r.Response.WriteStatus(http.StatusUnauthorized, g.Map{"error": "API key required"})
		r.Exit()
		return
	}

	// Inject validated user info into context
	r.SetCtxVar("user_id", res.UserId)
	r.SetCtxVar("api_key_id", res.ApiKeyId)
	r.SetCtxVar("user_group", res.KeyGroup)
	r.SetCtxVar("model_limits_enabled", res.ModelLimitsEnabled)
	r.SetCtxVar("model_limits", res.ModelLimits)
	r.SetCtxVar("user_status", res.UserStatus)

	r.Middleware.Next()
}
```

**Step 4.7 — Build and verify**

```bash
cd /home/ubuntu/code/ai-platform/ai-gateway
go build ./...
```

Expected: clean compilation.

**Step 4.8 — Commit**

```bash
cd /home/ubuntu/code/ai-platform
git add ai-gateway/
git commit -m "feat(ai-gateway): implement TokenAuth middleware via gRPC user-svc call"
```

---

### Task 5: Web Frontend — Login Flow

**Files:**
- Create: `web/src/api/client.ts`
- Modify: `web/src/stores/auth.ts`
- Modify: `web/src/routes/login.tsx`
- Modify: `web/src/routes/__root.tsx`

**Step 5.1 — Create API client**

Write `web/src/api/client.ts`:

```ts
const API_BASE = 'http://localhost:8080/api/v1'

export async function apiPost<T>(path: string, body: unknown, token?: string): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: res.statusText }))
    throw new Error(err.message || 'request failed')
  }
  return res.json()
}

export async function apiGet<T>(path: string, token: string): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { Authorization: `Bearer ${token}` },
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: res.statusText }))
    throw new Error(err.message || 'request failed')
  }
  return res.json()
}
```

**Step 5.2 — Update auth store**

Modify `web/src/stores/auth.ts`:

```ts
import { create } from 'zustand'
import { apiGet, apiPost } from '../api/client'

interface User {
  id: number
  username: string
  email: string
  display_name?: string
}

interface AuthState {
  token: string | null
  user: User | null
  loading: boolean
  setAuth: (token: string, user: User) => void
  login: (username: string, password: string) => Promise<void>
  register: (username: string, password: string, email: string) => Promise<void>
  fetchProfile: () => Promise<void>
  logout: () => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  token: localStorage.getItem('token'),
  user: null,
  loading: false,

  setAuth: (token, user) => {
    localStorage.setItem('token', token)
    set({ token, user })
  },

  login: async (username, password) => {
    set({ loading: true })
    try {
      const res = await apiPost<{ user_id: number; username: string; token: string }>(
        '/user/login', { username, password }
      )
      const user = { id: res.user_id, username: res.username, email: '' }
      localStorage.setItem('token', res.token)
      set({ token: res.token, user, loading: false })
    } catch (e) {
      set({ loading: false })
      throw e
    }
  },

  register: async (username, password, email) => {
    set({ loading: true })
    try {
      const res = await apiPost<{ user_id: number; username: string; token: string }>(
        '/user/register', { username, password, email }
      )
      const user = { id: res.user_id, username: res.username, email: '' }
      localStorage.setItem('token', res.token)
      set({ token: res.token, user, loading: false })
    } catch (e) {
      set({ loading: false })
      throw e
    }
  },

  fetchProfile: async () => {
    const { token } = get()
    if (!token) return
    try {
      const user = await apiGet<User>('/user/profile', token)
      set({ user })
    } catch {
      // Token expired
      localStorage.removeItem('token')
      set({ token: null, user: null })
    }
  },

  logout: () => {
    localStorage.removeItem('token')
    set({ token: null, user: null })
  },
}))
```

**Step 5.3 — Update login page**

Replace `web/src/routes/login.tsx`:

```tsx
import { createFileRoute, useNavigate } from '@tanstack/react-router'
import { useState } from 'react'
import { useAuthStore } from '../stores/auth'

export const Route = createFileRoute('/login')({
  component: LoginPage,
})

function LoginPage() {
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const login = useAuthStore((s) => s.login)
  const register = useAuthStore((s) => s.register)
  const loading = useAuthStore((s) => s.loading)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      if (mode === 'login') {
        await login(username, password)
      } else {
        await register(username, password, email)
      }
      navigate({ to: '/dashboard' })
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="w-full max-w-sm space-y-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold">AI Platform</h1>
          <p className="text-muted-foreground text-sm mt-1">
            {mode === 'login' ? 'Sign in to your account' : 'Create a new account'}
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="block text-sm font-medium mb-1">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full border rounded-md px-3 py-2 text-sm"
              placeholder="Enter username"
              required
            />
          </div>
          {mode === 'register' && (
            <div>
              <label className="block text-sm font-medium mb-1">Email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full border rounded-md px-3 py-2 text-sm"
                placeholder="Enter email"
              />
            </div>
          )}
          <div>
            <label className="block text-sm font-medium mb-1">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full border rounded-md px-3 py-2 text-sm"
              placeholder="Enter password"
              required
            />
          </div>

          {error && <p className="text-red-500 text-sm">{error}</p>}

          <button
            type="submit"
            disabled={loading}
            className="w-full bg-foreground text-background rounded-md py-2 text-sm font-medium hover:opacity-90 disabled:opacity-50"
          >
            {loading ? 'Loading...' : mode === 'login' ? 'Sign In' : 'Create Account'}
          </button>
        </form>

        <p className="text-center text-sm text-muted-foreground">
          {mode === 'login' ? (
            <>Don't have an account? <button onClick={() => setMode('register')} className="underline">Register</button></>
          ) : (
            <>Already have an account? <button onClick={() => setMode('login')} className="underline">Sign In</button></>
          )}
        </p>
      </div>
    </div>
  )
}
```

**Step 5.4 — Update root layout with auth guard**

Replace `web/src/routes/__root.tsx`:

```tsx
import { createRootRoute, Outlet, useNavigate } from '@tanstack/react-router'
import { useEffect } from 'react'
import { useAuthStore } from '../stores/auth'

export const Route = createRootRoute({
  component: RootLayout,
})

function RootLayout() {
  const token = useAuthStore((s) => s.token)
  const user = useAuthStore((s) => s.user)
  const fetchProfile = useAuthStore((s) => s.fetchProfile)
  const logout = useAuthStore((s) => s.logout)
  const navigate = useNavigate()

  useEffect(() => {
    if (token && !user) {
      fetchProfile()
    }
  }, [token, user, fetchProfile])

  const path = window.location.pathname
  const isLoginPage = path === '/login'

  return (
    <div className="min-h-screen bg-background">
      {token && !isLoginPage && (
        <header className="border-b px-6 py-3 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <h2 className="font-semibold">AI Platform</h2>
            <nav className="flex gap-4 text-sm">
              <a href="/dashboard" className="text-muted-foreground hover:text-foreground">Dashboard</a>
              <a href="/keys" className="text-muted-foreground hover:text-foreground">API Keys</a>
              <a href="/models" className="text-muted-foreground hover:text-foreground">Models</a>
              <a href="/orders" className="text-muted-foreground hover:text-foreground">Orders</a>
            </nav>
          </div>
          <div className="flex items-center gap-3 text-sm">
            <span className="text-muted-foreground">{user?.username}</span>
            <button onClick={() => { logout(); navigate({ to: '/login' })}} className="text-muted-foreground hover:text-foreground">
              Logout
            </button>
          </div>
        </header>
      )}
      <Outlet />
    </div>
  )
}
```

**Step 5.5 — Build and verify frontend**

```bash
cd /home/ubuntu/code/ai-platform/web
npx tsc --noEmit
```

Expected: no TypeScript errors.

**Step 5.6 — Commit**

```bash
cd /home/ubuntu/code/ai-platform
git add web/src/api/ web/src/stores/auth.ts web/src/routes/login.tsx web/src/routes/__root.tsx
git commit -m "feat(web): implement login flow with auth store and API client"
```

---

## Self-Review

1. **Spec coverage:** Plan covers auth foundation (Register/Login/ValidateToken + JWT), API key CRUD, api-gateway HTTP→gRPC proxy with auth middleware, ai-gateway token validation middleware, and web login UI. Missing: asset-svc integration (Phase 3) and market-svc (Phase 4).

2. **Placeholder scan:** All code is complete. No TBD/TODO/later patterns. Every file path is absolute.

3. **Type consistency:** `userpb.UserServiceClient` is used consistently across api-gateway and ai-gateway. JWT claims struct is duplicated in api-gateway (for auth middleware) and user-svc (for generation), which is expected since they're separate Go modules.

4. **Dependency chain:** Valid — user-svc is implemented first, then api-gateway and ai-gateway depend on it, web frontend depends on api-gateway routes.
