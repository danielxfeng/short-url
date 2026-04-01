package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateTokenAndValidateToken(t *testing.T) {
	const secret = "test-secret"
	const userID int32 = 123

	tests := []struct {
		name       string
		tokenFunc  func(t *testing.T) string
		secret     string
		wantUserID int32
		wantErr    bool
	}{
		{
			name: "valid token",
			tokenFunc: func(t *testing.T) string {
				t.Helper()
				token, err := GenerateToken(userID, secret, time.Hour)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			secret:     secret,
			wantUserID: userID,
			wantErr:    false,
		},
		{
			name: "wrong secret",
			tokenFunc: func(t *testing.T) string {
				t.Helper()
				token, err := GenerateToken(userID, secret, time.Hour)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name: "malformed token",
			tokenFunc: func(t *testing.T) string {
				t.Helper()
				return "not-a-jwt"
			},
			secret:  secret,
			wantErr: true,
		},
		{
			name: "expired token",
			tokenFunc: func(t *testing.T) string {
				t.Helper()
				token, err := GenerateToken(userID, secret, -1*time.Minute)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			secret:  secret,
			wantErr: true,
		},
		{
			name: "too early",
			tokenFunc: func(t *testing.T) string {
				t.Helper()
				now := time.Now()
				claims := Claims{
					UserID: userID,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(now.Add(2 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(now),
						NotBefore: jwt.NewNumericDate(now.Add(30 * time.Minute)),
					},
				}

				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenStr, err := token.SignedString([]byte(secret))
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return tokenStr
			},
			secret:  secret,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			token := tc.tokenFunc(t)

			gotUserID, err := ValidateToken(token, tc.secret)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotUserID != tc.wantUserID {
				t.Fatalf("expected user id %d, got %d", tc.wantUserID, gotUserID)
			}
		})
	}
}
