package middleware

import (
	"errors"
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
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
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
