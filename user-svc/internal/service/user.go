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
		// Verify user still exists and is enabled
		user, err2 := g.DB().Model("users").Ctx(ctx).Where("id", claims.UserId).One()
		if err2 != nil {
			return nil, err2
		}
		if user == nil {
			return nil, errors.New("user not found")
		}
		if user["status"].Int() != 1 {
			return nil, errors.New("account is disabled")
		}
		return &userv1.ValidateTokenRes{
			UserId: claims.UserId, UserStatus: int32(user["status"].Int()),
			Group: user["group_name"].String(), HasToken: false,
			UserRole: claims.Role, TenantId: user["tenant_id"].Int64(),
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
		if et != nil && et.Time.Before(time.Now()) {
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

// Stub methods for API keys -- to be implemented in Task 2.
func (s *UserService) CreateApiKey(ctx context.Context, req *userv1.CreateApiKeyReq) (*userv1.CreateApiKeyRes, error) {
	return nil, errors.New("not implemented")
}

func (s *UserService) ListApiKeys(ctx context.Context, req *userv1.ListApiKeysReq) (*userv1.ListApiKeysRes, error) {
	return nil, errors.New("not implemented")
}

func (s *UserService) DeleteApiKey(ctx context.Context, req *userv1.DeleteApiKeyReq) (*userv1.DeleteApiKeyRes, error) {
	return nil, errors.New("not implemented")
}
