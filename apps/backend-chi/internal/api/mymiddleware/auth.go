package mymiddleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/util"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

func UserIDFromContext(ctx context.Context) (int32, bool) {
	v := ctx.Value(UserIDContextKey)
	userID, ok := v.(int32)
	return userID, ok
}

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			if !strings.HasPrefix(bearer, "Bearer ") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(bearer, "Bearer ")

			userID, err := util.ValidateToken(token, secret)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
