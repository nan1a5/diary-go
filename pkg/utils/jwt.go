package utils

import (
	"time"

	"diary/config"
	"github.com/golang-jwt/jwt/v5"
)

// CreateJWTToken 生成 JWT Token
func CreateJWTToken(userID uint, cfg *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * time.Duration(cfg.JWTExpireHours)).Unix(),
		"iat": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// ParseJWTToken 解析 JWT Token
func ParseJWTToken(tokenStr string, cfg *config.Config) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWTSecret), nil
	})
}

// GetUserIDFromToken 从 Token 中获取用户 ID
func GetUserIDFromToken(tokenStr string, cfg *config.Config) (uint, error) {
	token, err := ParseJWTToken(tokenStr, cfg)
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if sub, ok := claims["sub"].(float64); ok {
			return uint(sub), nil
		}
	}

	return 0, jwt.ErrInvalidKey
}

