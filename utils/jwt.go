package utils

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lyuangg/gadmin/config"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func InitJWT(cfg *config.Config) {
	jwtSecret = []byte(cfg.JWTSecret)
}

func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	if ctx == nil {
		return nil, false
	}
	val := ctx.Value("claims")
	if val == nil {
		return nil, false
	}
	claims, ok := val.(*Claims)
	return claims, ok && claims != nil
}

type Claims struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	Type         int    `json:"type"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	RoleIDs      []uint `json:"role_ids"`
	jwt.RegisteredClaims
}

// GenerateToken tokenVersion 用于退出时失效该用户所有 token
func GenerateToken(userID uint, username, nickname string, userType int, isSuperAdmin bool, roleIDs []uint, tokenVersion uint) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(24 * time.Hour)
	jti := fmt.Sprintf("%d:%d", userID, tokenVersion)

	claims := Claims{
		UserID:       userID,
		Username:     username,
		Nickname:     nickname,
		Type:         userType,
		IsSuperAdmin: isSuperAdmin,
		RoleIDs:      roleIDs,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(nowTime),
			Issuer:    "gadmin",
			ID:        jti,
		},
	}

	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func GetTokenVersion(claims *Claims) (uint, uint, error) {
	if claims.ID == "" {
		return 0, 0, errors.New("token中缺少jti字段")
	}

	parts := strings.Split(claims.ID, ":")
	if len(parts) != 2 {
		return 0, 0, errors.New("token的jti格式错误")
	}

	userID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		return 0, 0, errors.New("无法解析token中的user_id")
	}

	tokenVersion, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return 0, 0, errors.New("无法解析token中的token_version")
	}

	return uint(userID), uint(tokenVersion), nil
}
