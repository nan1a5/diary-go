package middleware

import (
	"context"
	"diary/config"
	"diary/internal/models"
	"diary/pkg/utils"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type key int

const userCtxKey key = 0

func AuthMiddleware(cfg *config.Config, db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				next.ServeHTTP(w, r)
				return
			}
			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			tokenStr := parts[1]
			token, err := utils.ParseJWTToken(tokenStr, cfg)
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}
			sub := claims["sub"]
			var uid uint64
			switch v := sub.(type) {
			case float64:
				uid = uint64(v)
			case string:
				uid, _ = strconv.ParseUint(v, 10, 64)
			default:
				http.Error(w, "invalid sub claim", http.StatusUnauthorized)
				return
			}
			var u models.User
			if err := db.First(&u, uid).Error; err != nil {
				http.Error(w, "user not found", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userCtxKey, &u)
			ctx = context.WithValue(ctx, "user_id", u.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func UserFromContext(r *http.Request) *models.User {
	u, _ := r.Context().Value(userCtxKey).(*models.User)
	return u
}
