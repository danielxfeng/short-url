package mymiddleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/apierror"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/auth"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/util"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

func UserIDFromContext(ctx context.Context) (int32, bool) {
	v := ctx.Value(UserIDContextKey)
	userID, ok := v.(int32)
	return userID, ok
}

func MustUserIDFromContext(ctx context.Context) int32 {
	userID, ok := UserIDFromContext(ctx)
	if !ok {
		panic("user id missing from context")
	}
	return userID
}

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearer := r.Header.Get("Authorization")
			if !strings.HasPrefix(bearer, "Bearer ") {
				util.SendError(w, apierror.NewApiError(http.StatusUnauthorized, "Unauthorized", nil))
				return
			}

			token := strings.TrimPrefix(bearer, "Bearer ")

			userID, err := auth.ValidateToken(token, secret)
			if err != nil {
				util.SendError(w, apierror.NewApiError(http.StatusUnauthorized, "Unauthorized", nil))
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
