package mymiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/util"
)

func TestAuth(t *testing.T) {
	const secret = "test-secret"
	const userID int32 = 42

	validToken, err := util.GenerateToken(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("generate valid token: %v", err)
	}

	invalidToken, err := util.GenerateToken(userID, "wrong-secret", time.Hour)
	if err != nil {
		t.Fatalf("generate invalid token: %v", err)
	}

	tests := []struct {
		name          string
		authHeader    string
		wantStatus    int
		wantNext      bool
		wantHasUserID bool
		wantUserID    int32
	}{
		{
			name:       "missing authorization",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "too short auth header like Bea",
			authHeader: "Bea",
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "bare bearer without token",
			authHeader: "Bearer",
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer " + invalidToken,
			wantStatus: http.StatusUnauthorized,
			wantNext:   false,
		},
		{
			name:          "valid token writes user id to context",
			authHeader:    "Bearer " + validToken,
			wantStatus:    http.StatusOK,
			wantNext:      true,
			wantHasUserID: true,
			wantUserID:    userID,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				gotUserID, ok := UserIDFromContext(r.Context())
				if ok != tc.wantHasUserID {
					t.Fatalf("expected has user id %v, got %v", tc.wantHasUserID, ok)
				}
				if tc.wantHasUserID && gotUserID != tc.wantUserID {
					t.Fatalf("expected user id %d, got %d", tc.wantUserID, gotUserID)
				}
				w.WriteHeader(http.StatusOK)
			})

			h := Auth(secret)(next)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, rr.Code)
			}
			if nextCalled != tc.wantNext {
				t.Fatalf("expected next called %v, got %v", tc.wantNext, nextCalled)
			}
		})
	}
}
