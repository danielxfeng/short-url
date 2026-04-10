package router

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/auth"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/dto"
	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/db"
	stateStore "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/inmemory"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/models"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"golang.org/x/oauth2"
)

type mockOauthHandler struct {
	config           auth.OauthConfig
	exchangeErr      error
	expectedVerifier string
	nextState        string
	authURLBase      string
}

func newMockOauthHandler() *mockOauthHandler {
	name := "Mock Google"
	pic := "https://example.com/mock-google.png"

	return &mockOauthHandler{
		expectedVerifier: "mock-verifier",
		nextState:        "mock-state",
		authURLBase:      "https://accounts.google.com/o/oauth2/v2/auth",
		config: auth.OauthConfig{
			Config: oauth2.Config{ClientID: "mock-google-client-id"},
			GetUserInfo: func(client *http.Client) (*models.UpsertUserParams, error) {
				_ = client
				return &models.UpsertUserParams{
					Provider:    models.ProviderEnumGOOGLE,
					ProviderID:  "google-sub-123",
					DisplayName: &name,
					ProfilePic:  &pic,
				}, nil
			},
		},
	}
}

func (m *mockOauthHandler) GetConfigForProvider(provider string) (*auth.OauthConfig, bool) {
	if strings.EqualFold(strings.TrimSpace(provider), "google") {
		cfg := m.config
		return &cfg, true
	}
	return nil, false
}

func (m *mockOauthHandler) GetOauthAuthURL(opt *oauth2.Config, store models.StateStoreRepository) string {
	_ = opt
	store.Add(m.nextState, m.expectedVerifier)
	return m.authURLBase + "?state=" + url.QueryEscape(m.nextState)
}

func (m *mockOauthHandler) ExchangeCodeAndGetClient(ctx context.Context, opt *oauth2.Config, code string, verifier string) (*http.Client, error) {
	_ = ctx
	_ = opt
	_ = code
	if m.exchangeErr != nil {
		return nil, m.exchangeErr
	}
	if verifier != m.expectedVerifier {
		return nil, errors.New("unexpected verifier")
	}
	return &http.Client{}, nil
}

type userTestApp struct {
	dep     *dep.Dep
	q       *db.Queries
	oauth   *mockOauthHandler
	store   models.StateStoreRepository
	repo    models.Repository
	handler http.Handler
}

func newUserIntegrationSetup(t *testing.T) *userTestApp {
	t.Helper()

	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)

	cfg := &dep.Config{
		AppMode:             dep.EnvTest,
		JWTSecret:           "test-secret",
		JWTExpiry:           time.Hour,
		BackendPublicURL:    "https://api.example.com",
		FrontendRedirectURL: "https://frontend.example.com/auth/callback",
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dep.NewDep(cfg, logger, sharedPool)
	helper := newMockOauthHandler()

	store := stateStore.NewMemoryStateStore()
	repo := models.NewRepository(q, q, store)
	return &userTestApp{
		dep:     d,
		q:       q,
		oauth:   helper,
		store:   store,
		repo:    repo,
		handler: UserRouter(d, repo, helper),
	}
}

func (a *userTestApp) rebuildHandler() {
	a.repo = models.NewRepository(a.q, a.q, a.store)
	a.handler = UserRouter(a.dep, a.repo, a.oauth)
}

func (a *userTestApp) withState(state string, verifier string) *userTestApp {
	a.store.Add(state, verifier)
	a.rebuildHandler()
	return a
}

func (a *userTestApp) request(method string, path string) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.handler.ServeHTTP(rr, httptest.NewRequest(method, path, nil))
	return rr
}

func TestUserRouter_AuthStart(t *testing.T) {
	testCases := []struct {
		name     string
		provider string
	}{
		{name: "unsupported provider", provider: "unsupported"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := newUserIntegrationSetup(t)

			rr := app.request(http.MethodGet, "/auth/"+tc.provider)

			if rr.Code != http.StatusFound {
				t.Fatalf("expected %d, got %d", http.StatusFound, rr.Code)
			}

			location := rr.Header().Get("Location")
			if location == "" {
				t.Fatalf("expected redirect location")
			}

			parsed, err := url.Parse(location)
			if err != nil {
				t.Fatalf("parse redirect location: %v", err)
			}
			if parsed.Scheme+"://"+parsed.Host+parsed.Path != app.dep.Cfg.FrontendRedirectURL {
				t.Fatalf("expected redirect base %q, got %q", app.dep.Cfg.FrontendRedirectURL, parsed.Scheme+"://"+parsed.Host+parsed.Path)
			}
			if parsed.Query().Get("error") != "unsupported provider" {
				t.Fatalf("expected unsupported provider error, got %q", parsed.Query().Get("error"))
			}
		})
	}
}

func TestUserRouter_Callback(t *testing.T) {
	testCases := []struct {
		name      string
		provider  string
		query     string
		setup     func(t *testing.T, app *userTestApp)
		verify    func(t *testing.T, app *userTestApp, rr *httptest.ResponseRecorder)
		wantError string
	}{
		{
			name:     "success creates user and returns auth token",
			provider: "google",
			query:    "code=mock-code&state=mock-state",
			setup: func(t *testing.T, app *userTestApp) {
				t.Helper()
				app.withState("mock-state", app.oauth.expectedVerifier)
			},
			verify: func(t *testing.T, app *userTestApp, rr *httptest.ResponseRecorder) {
				t.Helper()
				location := rr.Header().Get("Location")
				parsed, err := url.Parse(location)
				if err != nil {
					t.Fatalf("parse callback redirect: %v", err)
				}
				token := parsed.Query().Get("auth")
				if token == "" {
					t.Fatalf("expected auth token in callback redirect")
				}
				userID, err := auth.ValidateToken(token, app.dep.Cfg.JWTSecret)
				if err != nil {
					t.Fatalf("validate token: %v", err)
				}
				user, err := app.q.GetUserByID(context.Background(), userID)
				if err != nil {
					t.Fatalf("expected upserted user in db: %v", err)
				}
				if user.Provider != db.ProviderEnumGOOGLE {
					t.Fatalf("expected provider GOOGLE, got %q", user.Provider)
				}
				if user.ProviderID != "google-sub-123" {
					t.Fatalf("expected provider id google-sub-123, got %q", user.ProviderID)
				}
			},
		},
		{
			name:      "missing code",
			provider:  "google",
			query:     "state=mock-state",
			setup:     nil,
			wantError: "code not found in query",
		},
		{
			name:      "missing state",
			provider:  "google",
			query:     "code=mock-code",
			setup:     nil,
			wantError: "invalid state",
		},
		{
			name:      "invalid state",
			provider:  "google",
			query:     "code=mock-code&state=unknown-state",
			setup:     nil,
			wantError: "invalid state",
		},
		{
			name:     "unsupported provider",
			provider: "github",
			query:    "code=mock-code&state=mock-state",
			setup: func(t *testing.T, app *userTestApp) {
				t.Helper()
				app.withState("mock-state", app.oauth.expectedVerifier)
			},
			wantError: "unsupported provider",
		},
		{
			name:     "exchange token fails",
			provider: "google",
			query:    "code=mock-code&state=mock-state",
			setup: func(t *testing.T, app *userTestApp) {
				t.Helper()
				app.oauth.exchangeErr = errors.New("exchange failed")
				app.withState("mock-state", app.oauth.expectedVerifier)
			},
			wantError: "failed to exchange code for token",
		},
		{
			name:     "get user info fails",
			provider: "google",
			query:    "code=mock-code&state=mock-state",
			setup: func(t *testing.T, app *userTestApp) {
				t.Helper()
				app.oauth.config.GetUserInfo = func(client *http.Client) (*models.UpsertUserParams, error) {
					_ = client
					return nil, errors.New("userinfo failed")
				}
				app.withState("mock-state", app.oauth.expectedVerifier)
			},
			wantError: "failed to get user info",
		},
		{
			name:     "get user info returns nil",
			provider: "google",
			query:    "code=mock-code&state=mock-state",
			setup: func(t *testing.T, app *userTestApp) {
				t.Helper()
				app.oauth.config.GetUserInfo = func(client *http.Client) (*models.UpsertUserParams, error) {
					_ = client
					return nil, nil
				}
				app.withState("mock-state", app.oauth.expectedVerifier)
			},
			wantError: "failed to get user info",
		},
		{
			name:     "upsert user fails",
			provider: "google",
			query:    "code=mock-code&state=mock-state",
			setup: func(t *testing.T, app *userTestApp) {
				t.Helper()
				helpDisplayName := "Broken User"
				app.oauth.config.GetUserInfo = func(client *http.Client) (*models.UpsertUserParams, error) {
					_ = client
					return &models.UpsertUserParams{
						Provider:    models.ProviderEnum("INVALID_PROVIDER"),
						ProviderID:  "provider-id",
						DisplayName: &helpDisplayName,
					}, nil
				}
				app.withState("mock-state", app.oauth.expectedVerifier)
			},
			wantError: "failed to upsert user",
		},
		{
			name:     "state replay rejected",
			provider: "google",
			query:    "code=mock-code&state=mock-state",
			setup: func(t *testing.T, app *userTestApp) {
				t.Helper()
				app.withState("mock-state", app.oauth.expectedVerifier)
			},
			verify: func(t *testing.T, app *userTestApp, rr *httptest.ResponseRecorder) {
				t.Helper()
				firstLocation := rr.Header().Get("Location")
				if firstLocation == "" {
					t.Fatalf("expected first callback redirect location")
				}
				replayApp := newUserIntegrationSetup(t)
				replayApp.withState("mock-state", replayApp.oauth.expectedVerifier)
				firstRR := httptest.NewRecorder()
				replayApp.handler.ServeHTTP(firstRR, httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=mock-code&state=mock-state", nil))
				secondRR := httptest.NewRecorder()
				replayApp.handler.ServeHTTP(secondRR, httptest.NewRequest(http.MethodGet, "/auth/google/callback?code=mock-code&state=mock-state", nil))
				secondLocation := secondRR.Header().Get("Location")
				parsedSecond, err := url.Parse(secondLocation)
				if err != nil {
					t.Fatalf("parse second callback redirect: %v", err)
				}
				if parsedSecond.Query().Get("error") != "invalid state" {
					t.Fatalf("expected invalid state on replay, got %q", parsedSecond.Query().Get("error"))
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := newUserIntegrationSetup(t)
			if tc.setup != nil {
				tc.setup(t, app)
			}

			rr := app.request(http.MethodGet, "/auth/"+tc.provider+"/callback?"+tc.query)

			if rr.Code != http.StatusFound {
				t.Fatalf("expected %d, got %d", http.StatusFound, rr.Code)
			}

			location := rr.Header().Get("Location")
			if location == "" {
				t.Fatalf("expected callback redirect location")
			}

			if tc.verify != nil {
				tc.verify(t, app, rr)
				return
			}

			parsed, err := url.Parse(location)
			if err != nil {
				t.Fatalf("parse callback redirect: %v", err)
			}
			if parsed.Scheme+"://"+parsed.Host+parsed.Path != app.dep.Cfg.FrontendRedirectURL {
				t.Fatalf("expected redirect base %q, got %q", app.dep.Cfg.FrontendRedirectURL, parsed.Scheme+"://"+parsed.Host+parsed.Path)
			}
			if parsed.Query().Get("error") != tc.wantError {
				t.Fatalf("expected error %q, got %q", tc.wantError, parsed.Query().Get("error"))
			}
		})
	}
}

func TestUserRouter_GetMe(t *testing.T) {
	testCases := []struct {
		name           string
		tokenFactory   func(t *testing.T, d *dep.Dep, q *db.Queries) string
		wantStatusCode int
		wantBodySubstr string
		verify         func(t *testing.T, q *db.Queries, seeded *db.User, rr *httptest.ResponseRecorder)
	}{
		{
			name: "success returns user dto",
			tokenFactory: func(t *testing.T, d *dep.Dep, q *db.Queries) string {
				t.Helper()
				seedName := "Seed User"
				seedPic := "https://example.com/seed.png"
				seeded, err := q.UpsertUser(context.Background(), db.UpsertUserParams{
					Provider:    db.ProviderEnumGOOGLE,
					ProviderID:  "seed-google-get",
					DisplayName: &seedName,
					ProfilePic:  &seedPic,
				})
				if err != nil {
					t.Fatalf("seed user: %v", err)
				}
				token, err := auth.GenerateToken(seeded.ID, d.Cfg.JWTSecret, d.Cfg.JWTExpiry)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			wantStatusCode: http.StatusOK,
			verify: func(t *testing.T, q *db.Queries, seeded *db.User, rr *httptest.ResponseRecorder) {
				t.Helper()
				_ = q
				var got dto.UserResponse
				if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
					t.Fatalf("decode user response: %v", err)
				}
				if got.ID == 0 {
					t.Fatalf("expected non-zero id")
				}
				if got.Provider != models.ProviderEnumGOOGLE {
					t.Fatalf("expected provider %q, got %q", models.ProviderEnumGOOGLE, got.Provider)
				}
				if got.ProviderID != "seed-google-get" {
					t.Fatalf("expected provider_id %q, got %q", "seed-google-get", got.ProviderID)
				}
				if got.DisplayName == nil || *got.DisplayName != "Seed User" {
					t.Fatalf("expected display_name %q, got %v", "Seed User", got.DisplayName)
				}
				if got.ProfilePic == nil || *got.ProfilePic != "https://example.com/seed.png" {
					t.Fatalf("expected profile_pic %q, got %v", "https://example.com/seed.png", got.ProfilePic)
				}
			},
		},
		{
			name: "missing bearer token",
			tokenFactory: func(t *testing.T, d *dep.Dep, q *db.Queries) string {
				t.Helper()
				_ = d
				_ = q
				return ""
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBodySubstr: "Unauthorized",
		},
		{
			name: "invalid bearer token",
			tokenFactory: func(t *testing.T, d *dep.Dep, q *db.Queries) string {
				t.Helper()
				_ = d
				_ = q
				return "not-a-jwt"
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBodySubstr: "Unauthorized",
		},
		{
			name: "user not found",
			tokenFactory: func(t *testing.T, d *dep.Dep, q *db.Queries) string {
				t.Helper()
				_ = q
				token, err := auth.GenerateToken(987654321, d.Cfg.JWTSecret, d.Cfg.JWTExpiry)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			wantStatusCode: http.StatusNotFound,
			wantBodySubstr: "user not found",
		},
		{
			name: "already deleted user",
			tokenFactory: func(t *testing.T, d *dep.Dep, q *db.Queries) string {
				t.Helper()
				seedName := "Deleted User"
				seeded, err := q.UpsertUser(context.Background(), db.UpsertUserParams{
					Provider:    db.ProviderEnumGOOGLE,
					ProviderID:  "seed-google-delete-already-deleted",
					DisplayName: &seedName,
				})
				if err != nil {
					t.Fatalf("seed user: %v", err)
				}

				token, err := auth.GenerateToken(seeded.ID, d.Cfg.JWTSecret, d.Cfg.JWTExpiry)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}

				if _, err := q.DeleteUser(context.Background(), seeded.ID); err != nil {
					t.Fatalf("pre-delete user: %v", err)
				}

				return token
			},
			wantStatusCode: http.StatusNotFound,
			wantBodySubstr: "user not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := newUserIntegrationSetup(t)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/me/", nil)
			token := tc.tokenFactory(t, app.dep, app.q)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			app.handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatusCode {
				t.Fatalf("expected %d, got %d", tc.wantStatusCode, rr.Code)
			}
			if tc.wantBodySubstr != "" && !strings.Contains(rr.Body.String(), tc.wantBodySubstr) {
				t.Fatalf("expected response body containing %q, got %q", tc.wantBodySubstr, rr.Body.String())
			}
			if tc.verify != nil {
				tc.verify(t, app.q, nil, rr)
			}
		})
	}
}

func TestUserRouter_DeleteMe(t *testing.T) {
	testCases := []struct {
		name           string
		tokenFactory   func(t *testing.T, d *dep.Dep, q *db.Queries) string
		wantStatusCode int
		wantBodySubstr string
		verify         func(t *testing.T, q *db.Queries, rr *httptest.ResponseRecorder)
	}{
		{
			name: "success deletes user",
			tokenFactory: func(t *testing.T, d *dep.Dep, q *db.Queries) string {
				t.Helper()
				seedName := "Seed User"
				seedPic := "https://example.com/seed.png"
				seeded, err := q.UpsertUser(context.Background(), db.UpsertUserParams{
					Provider:    db.ProviderEnumGOOGLE,
					ProviderID:  "seed-google-delete",
					DisplayName: &seedName,
					ProfilePic:  &seedPic,
				})
				if err != nil {
					t.Fatalf("seed user: %v", err)
				}
				token, err := auth.GenerateToken(seeded.ID, d.Cfg.JWTSecret, d.Cfg.JWTExpiry)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			wantStatusCode: http.StatusNoContent,
			verify: func(t *testing.T, q *db.Queries, rr *httptest.ResponseRecorder) {
				t.Helper()
				_ = rr
				_, err := q.GetUserByID(context.Background(), 1)
				if err == nil {
					t.Fatalf("expected seeded user to be deleted")
				}
			},
		},
		{
			name: "missing bearer token",
			tokenFactory: func(t *testing.T, d *dep.Dep, q *db.Queries) string {
				t.Helper()
				_ = d
				_ = q
				return ""
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBodySubstr: "Unauthorized",
		},
		{
			name: "invalid bearer token",
			tokenFactory: func(t *testing.T, d *dep.Dep, q *db.Queries) string {
				t.Helper()
				_ = d
				_ = q
				return "not-a-jwt"
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBodySubstr: "Unauthorized",
		},
		{
			name: "user not found",
			tokenFactory: func(t *testing.T, d *dep.Dep, q *db.Queries) string {
				t.Helper()
				_ = q
				token, err := auth.GenerateToken(987654321, d.Cfg.JWTSecret, d.Cfg.JWTExpiry)
				if err != nil {
					t.Fatalf("generate token: %v", err)
				}
				return token
			},
			wantStatusCode: http.StatusNotFound,
			wantBodySubstr: "user not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := newUserIntegrationSetup(t)

			token := tc.tokenFactory(t, app.dep, app.q)

			rr := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/me/", nil)
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}
			app.handler.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatusCode {
				t.Fatalf("expected %d, got %d", tc.wantStatusCode, rr.Code)
			}
			if tc.wantBodySubstr != "" && !strings.Contains(rr.Body.String(), tc.wantBodySubstr) {
				t.Fatalf("expected response body containing %q, got %q", tc.wantBodySubstr, rr.Body.String())
			}
			if tc.verify != nil {
				tc.verify(t, app.q, rr)
			}
		})
	}
}
