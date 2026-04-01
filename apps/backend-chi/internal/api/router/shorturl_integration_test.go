package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/auth"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/dto"
	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/db"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/models"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	sharedPool    *pgxpool.Pool
	sharedQueries *db.Queries
)

func loadTestEnv() {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}

	envPath := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../../.env"))
	_ = godotenv.Load(envPath)
}

func TestMain(m *testing.M) {
	loadTestEnv()

	testDbURL, err := dep.GetEnvStrOrError("TEST_DB_URL")
	if err != nil {
		log.Fatalf("setup test db: %v", err)
	}

	if err := db.MigrateDB(testDbURL); err != nil {
		log.Fatalf("setup test db: %v", err)
	}

	pool, err := db.NewPool(testDbURL)
	if err != nil {
		log.Fatalf("setup test db: %v", err)
	}

	sharedPool = pool
	sharedQueries = db.New(pool)
	dto.InitValidator()
	if err := sharedQueries.ResetDb(context.Background()); err != nil {
		log.Fatalf("setup test db: %v", err)
	}

	exitCode := m.Run()
	if err := sharedQueries.ResetDb(context.Background()); err != nil {
		log.Printf("cleanup test db: %v", err)
	}
	db.ClosePool(sharedPool)
	os.Exit(exitCode)
}

func runBeforeEachReset(t *testing.T, q *db.Queries) {
	t.Helper()
	if err := q.ResetDb(context.Background()); err != nil {
		t.Fatalf("setup test db: %v", err)
	}
}

func newShortURLIntegrationSetup(t *testing.T, userID int32) (*dep.Dep, *db.Queries, string) {
	t.Helper()

	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)

	cfg := &dep.Config{
		AppMode:      dep.EnvTest,
		JWTSecret:    "test-secret",
		JWTExpiry:    time.Hour,
		NotFoundPage: "https://example.com/not-found",
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	d := dep.NewDep(cfg, logger, sharedPool)

	token, err := auth.GenerateToken(userID, cfg.JWTSecret, cfg.JWTExpiry)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	return d, q, token
}

func newShortURLTestRepo(q *db.Queries) models.Repository {
	return models.NewRepository(q, q, nil)
}

type shortURLTestApp struct {
	dep     *dep.Dep
	q       *db.Queries
	repo    models.Repository
	handler http.Handler
}

func newShortURLTestApp(t *testing.T, userID int32) (*shortURLTestApp, string) {
	t.Helper()
	d, q, token := newShortURLIntegrationSetup(t, userID)
	repo := newShortURLTestRepo(q)
	return &shortURLTestApp{
		dep:     d,
		q:       q,
		repo:    repo,
		handler: ShortURLRouter(d, repo),
	}, token
}

func (a *shortURLTestApp) seedUser(t *testing.T, providerID string) db.User {
	return seedUser(t, a.q, providerID)
}

func (a *shortURLTestApp) seedLink(t *testing.T, userID int32, code string) db.Link {
	return seedLink(t, a.q, userID, code)
}

func (a *shortURLTestApp) serve(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.handler.ServeHTTP(rr, req)
	return rr
}

func seedUser(t *testing.T, q *db.Queries, providerID string) db.User {
	t.Helper()
	name := "name-" + providerID
	pic := "https://example.com/" + providerID

	u, err := q.UpsertUser(context.Background(), db.UpsertUserParams{
		Provider:    db.ProviderEnumGITHUB,
		ProviderID:  providerID,
		DisplayName: &name,
		ProfilePic:  &pic,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

func seedLink(t *testing.T, q *db.Queries, userID int32, code string) db.Link {
	t.Helper()
	l, err := q.CreateLink(context.Background(), db.CreateLinkParams{
		UserID:      userID,
		Code:        code,
		OriginalUrl: "https://example.com/" + code,
	})
	if err != nil {
		t.Fatalf("seed link: %v", err)
	}
	return l
}

func authedReq(method string, path string, token string, body []byte) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func TestShortURLRouter_GetRedirect(t *testing.T) {
	app, _ := newShortURLTestApp(t, 42)
	u := app.seedUser(t, "redirect-user")
	seed := app.seedLink(t, u.ID, "abc12345")
	deleted := app.seedLink(t, u.ID, "deleted01")
	if _, err := app.q.SetLinkDeleted(context.Background(), db.SetLinkDeletedParams{Code: deleted.Code, UserID: u.ID}); err != nil {
		t.Fatalf("seed deleted link: %v", err)
	}

	testCases := []struct {
		name          string
		path          string
		wantLocation  string
		verifyClicked bool
	}{
		{name: "existing code redirects to original url", path: "/" + seed.Code, wantLocation: seed.OriginalUrl, verifyClicked: true},
		{name: "soft deleted code redirects to not found page", path: "/" + deleted.Code, wantLocation: app.dep.Cfg.NotFoundPage},
		{name: "missing code redirects to not found page", path: "/missing", wantLocation: app.dep.Cfg.NotFoundPage},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := app.serve(httptest.NewRequest(http.MethodGet, tc.path, nil))

			if rr.Code != http.StatusFound {
				t.Fatalf("expected %d, got %d", http.StatusFound, rr.Code)
			}
			if got := rr.Header().Get("Location"); got != tc.wantLocation {
				t.Fatalf("expected redirect to %q, got %q", tc.wantLocation, got)
			}

			if !tc.verifyClicked {
				return
			}

			deadline := time.Now().Add(800 * time.Millisecond)
			for {
				got, err := app.q.GetLinkByCode(context.Background(), seed.Code)
				if err == nil && got.Clicks > seed.Clicks {
					break
				}
				if time.Now().After(deadline) {
					t.Fatalf("click increment did not happen, err=%v", err)
				}
				time.Sleep(20 * time.Millisecond)
			}
		})
	}
}

func TestShortURLRouter_AuthRequiredForProtectedEndpoints(t *testing.T) {
	app, _ := newShortURLTestApp(t, 42)

	testCases := []struct {
		name    string
		request *http.Request
	}{
		{name: "list endpoint requires auth", request: httptest.NewRequest(http.MethodGet, "/", nil)},
		{name: "create endpoint requires auth", request: httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(`{"original_url":"https://example.com/new"}`)))},
		{name: "delete endpoint requires auth", request: httptest.NewRequest(http.MethodDelete, "/any-code", nil)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := app.serve(tc.request)
			if rr.Code != http.StatusUnauthorized {
				t.Fatalf("expected %d, got %d", http.StatusUnauthorized, rr.Code)
			}
		})
	}
}

func TestShortURLRouter_ListRequiresAuthAndSupportsPagination(t *testing.T) {
	testCases := []struct {
		name            string
		prepare         func(t *testing.T, q *db.Queries, userID int32)
		query           string
		secondPage      bool
		wantCount       int
		wantHasMore     bool
		wantCursorNil   bool
		wantDescOrder   bool
		wantIDsLTCursor bool
	}{
		{
			name:  "empty list when no seed",
			query: "/?limit=2",
			prepare: func(t *testing.T, q *db.Queries, userID int32) {
				_ = q
				_ = userID
			},
			wantCount:     0,
			wantHasMore:   false,
			wantCursorNil: true,
		},
		{
			name:  "one page valid params no next",
			query: "/?limit=10",
			prepare: func(t *testing.T, q *db.Queries, userID int32) {
				l1 := seedLink(t, q, userID, "c1")
				l2 := seedLink(t, q, userID, "c2")
				l3 := seedLink(t, q, userID, "c3")
				l4 := seedLink(t, q, userID, "c4")
				_ = []int32{l1.ID, l2.ID, l3.ID, l4.ID}
			},
			wantCount:     4,
			wantHasMore:   false,
			wantCursorNil: true,
			wantDescOrder: true,
		},
		{
			name:  "invalid params normalized",
			query: "/?limit=abc&cursor=abc",
			prepare: func(t *testing.T, q *db.Queries, userID int32) {
				seedLink(t, q, userID, "c1")
				seedLink(t, q, userID, "c2")
				seedLink(t, q, userID, "c3")
				seedLink(t, q, userID, "c4")
			},
			wantCount:     4,
			wantHasMore:   false,
			wantCursorNil: true,
			wantDescOrder: true,
		},
		{
			name:  "several pages page 1 has next",
			query: "/?limit=2",
			prepare: func(t *testing.T, q *db.Queries, userID int32) {
				seedLink(t, q, userID, "c1")
				seedLink(t, q, userID, "c2")
				seedLink(t, q, userID, "c3")
				seedLink(t, q, userID, "c4")
			},
			wantCount:     2,
			wantHasMore:   true,
			wantCursorNil: false,
			wantDescOrder: true,
		},
		{
			name:       "several pages page 2 no next and correct order",
			query:      "/?limit=2",
			secondPage: true,
			prepare: func(t *testing.T, q *db.Queries, userID int32) {
				seedLink(t, q, userID, "c1")
				seedLink(t, q, userID, "c2")
				seedLink(t, q, userID, "c3")
				seedLink(t, q, userID, "c4")
			},
			wantCount:       2,
			wantHasMore:     false,
			wantCursorNil:   true,
			wantDescOrder:   true,
			wantIDsLTCursor: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d, q, _ := newShortURLIntegrationSetup(t, 42)
			h := ShortURLRouter(d, newShortURLTestRepo(q))
			u := seedUser(t, q, "list-user")
			other := seedUser(t, q, "list-other")
			token, err := auth.GenerateToken(u.ID, d.Cfg.JWTSecret, d.Cfg.JWTExpiry)
			if err != nil {
				t.Fatalf("generate token: %v", err)
			}

			if tc.prepare != nil {
				tc.prepare(t, q, u.ID)
			}
			seedLink(t, q, other.ID, "other")

			path := tc.query
			var page1Cursor int32
			if tc.secondPage {
				firstRR := httptest.NewRecorder()
				h.ServeHTTP(firstRR, authedReq(http.MethodGet, "/?limit=2", token, nil))
				if firstRR.Code != http.StatusOK {
					t.Fatalf("expected %d, got %d", http.StatusOK, firstRR.Code)
				}
				var first dto.LinksResponse
				if err := json.NewDecoder(firstRR.Body).Decode(&first); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if first.Cursor == nil {
					t.Fatalf("expected non-nil cursor for page 1")
				}
				page1Cursor = *first.Cursor
				path = fmt.Sprintf("/?limit=2&cursor=%d", page1Cursor)
			}

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, authedReq(http.MethodGet, path, token, nil))
			if rr.Code != http.StatusOK {
				t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
			}

			var resp dto.LinksResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if len(resp.Links) != tc.wantCount {
				t.Fatalf("expected %d links, got %d", tc.wantCount, len(resp.Links))
			}

			if tc.wantDescOrder {
				for i := 0; i < len(resp.Links)-1; i++ {
					if resp.Links[i].ID < resp.Links[i+1].ID {
						t.Fatalf("expected descending order, got %d then %d", resp.Links[i].ID, resp.Links[i+1].ID)
					}
				}
			}

			if tc.wantIDsLTCursor {
				for _, l := range resp.Links {
					if l.ID >= page1Cursor {
						t.Fatalf("expected page2 ids < cursor %d, got %d", page1Cursor, l.ID)
					}
				}
			}

			if resp.HasMore != tc.wantHasMore {
				t.Fatalf("expected has_more=%v got %v", tc.wantHasMore, resp.HasMore)
			}

			if tc.wantCursorNil && resp.Cursor != nil {
				t.Fatalf("expected nil cursor, got %v", *resp.Cursor)
			}
			if !tc.wantCursorNil && resp.Cursor == nil {
				t.Fatalf("expected non-nil cursor")
			}
		})
	}
}

func TestShortURLRouter_ListExcludesSoftDeletedLinks(t *testing.T) {
	d, q, _ := newShortURLIntegrationSetup(t, 42)
	h := ShortURLRouter(d, newShortURLTestRepo(q))
	u := seedUser(t, q, "list-soft-user")
	token, err := auth.GenerateToken(u.ID, d.Cfg.JWTSecret, d.Cfg.JWTExpiry)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	l1 := seedLink(t, q, u.ID, "alive-1")
	l2 := seedLink(t, q, u.ID, "alive-2")
	deleted := seedLink(t, q, u.ID, "deleted-x")

	if _, err := q.SetLinkDeleted(context.Background(), db.SetLinkDeletedParams{Code: deleted.Code, UserID: u.ID}); err != nil {
		t.Fatalf("set link deleted: %v", err)
	}

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, authedReq(http.MethodGet, "/?limit=10", token, nil))
	if rr.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	var resp dto.LinksResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if len(resp.Links) != 2 {
		t.Fatalf("expected 2 links, got %d", len(resp.Links))
	}
	if resp.HasMore {
		t.Fatalf("expected has_more=false got true")
	}
	if resp.Cursor != nil {
		t.Fatalf("expected nil cursor, got %v", *resp.Cursor)
	}

	gotCodes := map[string]bool{}
	for _, l := range resp.Links {
		gotCodes[l.Code] = true
		if l.Code == deleted.Code {
			t.Fatalf("expected deleted code %q to be excluded", deleted.Code)
		}
	}
	if !gotCodes[l1.Code] || !gotCodes[l2.Code] {
		t.Fatalf("expected active codes %q and %q in response", l1.Code, l2.Code)
	}
}

func TestShortURLRouter_ListPaginationNormalizationBoundaries(t *testing.T) {
	testCases := []struct {
		name          string
		query         string
		wantCount     int
		wantHasMore   bool
		wantCursorNil bool
	}{
		{
			name:          "limit zero clamped to min",
			query:         "/?limit=0",
			wantCount:     1,
			wantHasMore:   true,
			wantCursorNil: false,
		},
		{
			name:          "negative limit clamped to min",
			query:         "/?limit=-1",
			wantCount:     1,
			wantHasMore:   true,
			wantCursorNil: false,
		},
		{
			name:          "huge limit clamped to max and returns all",
			query:         "/?limit=999999",
			wantCount:     4,
			wantHasMore:   false,
			wantCursorNil: true,
		},
		{
			name:          "negative cursor clamped to min yields empty",
			query:         "/?limit=10&cursor=-1",
			wantCount:     0,
			wantHasMore:   false,
			wantCursorNil: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d, q, _ := newShortURLIntegrationSetup(t, 42)
			h := ShortURLRouter(d, newShortURLTestRepo(q))
			u := seedUser(t, q, "list-normalize-user")
			token, err := auth.GenerateToken(u.ID, d.Cfg.JWTSecret, d.Cfg.JWTExpiry)
			if err != nil {
				t.Fatalf("generate token: %v", err)
			}

			seedLink(t, q, u.ID, "n1")
			seedLink(t, q, u.ID, "n2")
			seedLink(t, q, u.ID, "n3")
			seedLink(t, q, u.ID, "n4")

			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, authedReq(http.MethodGet, tc.query, token, nil))
			if rr.Code != http.StatusOK {
				t.Fatalf("expected %d, got %d", http.StatusOK, rr.Code)
			}

			var resp dto.LinksResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("decode response: %v", err)
			}

			if len(resp.Links) != tc.wantCount {
				t.Fatalf("expected %d links, got %d", tc.wantCount, len(resp.Links))
			}
			if resp.HasMore != tc.wantHasMore {
				t.Fatalf("expected has_more=%v got %v", tc.wantHasMore, resp.HasMore)
			}
			if tc.wantCursorNil && resp.Cursor != nil {
				t.Fatalf("expected nil cursor, got %v", *resp.Cursor)
			}
			if !tc.wantCursorNil && resp.Cursor == nil {
				t.Fatalf("expected non-nil cursor")
			}
		})
	}
}

func TestShortURLRouter_CreateAndDelete(t *testing.T) {
	d, q, _ := newShortURLIntegrationSetup(t, 42)
	h := ShortURLRouter(d, newShortURLTestRepo(q))
	u := seedUser(t, q, "create-delete-user")
	tokenForUser, err := auth.GenerateToken(u.ID, d.Cfg.JWTSecret, d.Cfg.JWTExpiry)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	createBody := []byte(`{"original_url":"https://example.com/new"}`)
	createRR := httptest.NewRecorder()
	h.ServeHTTP(createRR, authedReq(http.MethodPost, "/", tokenForUser, createBody))

	if createRR.Code != http.StatusCreated {
		t.Fatalf("expected %d, got %d", http.StatusCreated, createRR.Code)
	}

	var created db.Link
	if err := json.NewDecoder(createRR.Body).Decode(&created); err != nil {
		t.Fatalf("decode created response: %v", err)
	}
	if created.Code == "" {
		t.Fatalf("expected generated code")
	}

	if _, err := q.GetLinkByCode(context.Background(), created.Code); err != nil {
		t.Fatalf("expected created link to exist, got err=%v", err)
	}

	testCases := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{name: "delete existing", path: "/" + created.Code, wantStatus: http.StatusNoContent},
		{name: "delete soft-deleted", path: "/" + created.Code, wantStatus: http.StatusNotFound},
		{name: "delete not-exist", path: "/missing-code", wantStatus: http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, authedReq(http.MethodDelete, tc.path, tokenForUser, nil))
			if rr.Code != tc.wantStatus {
				t.Fatalf("expected %d, got %d", tc.wantStatus, rr.Code)
			}
		})
	}

	if _, err := q.GetLinkByCode(context.Background(), created.Code); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected deleted link to be hidden, got err=%v", err)
	}
}

func TestShortURLRouter_DeleteForeignOwnedLinkReturnsNotFound(t *testing.T) {
	d, q, _ := newShortURLIntegrationSetup(t, 42)
	h := ShortURLRouter(d, newShortURLTestRepo(q))

	owner := seedUser(t, q, "owner-user")
	attacker := seedUser(t, q, "attacker-user")
	link := seedLink(t, q, owner.ID, "owned1234")

	attackerToken, err := auth.GenerateToken(attacker.ID, d.Cfg.JWTSecret, d.Cfg.JWTExpiry)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, authedReq(http.MethodDelete, "/"+link.Code, attackerToken, nil))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, rr.Code)
	}

	got, err := q.GetLinkByCode(context.Background(), link.Code)
	if err != nil {
		t.Fatalf("expected link still exists, got err=%v", err)
	}
	if got.UserID != owner.ID {
		t.Fatalf("expected link owner id %d, got %d", owner.ID, got.UserID)
	}
}

func TestShortURLRouter_CreateValidationError(t *testing.T) {
	d, q, token := newShortURLIntegrationSetup(t, 42)
	h := ShortURLRouter(d, newShortURLTestRepo(q))
	_ = seedUser(t, q, "validation-user")

	badBody := []byte(`{"original_url":"not-a-url"}`)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, authedReq(http.MethodPost, "/", token, badBody))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestShortURLRouter_CreateMalformedJSONError(t *testing.T) {
	d, q, token := newShortURLIntegrationSetup(t, 42)
	h := ShortURLRouter(d, newShortURLTestRepo(q))
	_ = seedUser(t, q, "malformed-user")

	badBody := []byte(`{"original_url":`)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, authedReq(http.MethodPost, "/", token, badBody))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var apiErr dto.APIErrorRes
	if err := json.NewDecoder(rr.Body).Decode(&apiErr); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if apiErr.Error == "" {
		t.Fatalf("expected non-empty error message")
	}
}
