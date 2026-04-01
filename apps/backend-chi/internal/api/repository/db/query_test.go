// This is not for testing the auto-generated code, but to test the SQL logic.

package db_test

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/db"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	sharedPool    *pgxpool.Pool
	sharedQueries *db.Queries
)

type testStore struct {
	q *db.Queries
}

func loadTestEnv() {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return
	}

	envPath := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../../../.env"))
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

func newTestStore(t *testing.T) *testStore {
	t.Helper()
	if sharedQueries == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}

	s := &testStore{q: sharedQueries}
	s.reset(t)
	return s
}

func (s *testStore) reset(t *testing.T) {
	t.Helper()
	if err := s.q.ResetDb(context.Background()); err != nil {
		t.Fatalf("setup test db: %v", err)
	}
}

func (s *testStore) mkUser(t *testing.T, providerID string) db.User {
	t.Helper()
	name := "name-" + providerID
	pic := "https://example.com/" + providerID

	u, err := s.q.UpsertUser(context.Background(), db.UpsertUserParams{
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

func (s *testStore) mkLink(t *testing.T, userID int32, code string) db.Link {
	t.Helper()
	l, err := s.q.CreateLink(context.Background(), db.CreateLinkParams{
		UserID:      userID,
		Code:        code,
		OriginalUrl: "https://example.com/" + code,
	})
	if err != nil {
		t.Fatalf("seed link: %v", err)
	}
	return l
}

func expectPgCode(t *testing.T, err error, want string) {
	t.Helper()
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("expected pg error, got %T (%v)", err, err)
	}
	if pgErr.Code != want {
		t.Fatalf("expected pg code %s, got %s", want, pgErr.Code)
	}
}

func expectNoRows(t *testing.T, err error) {
	t.Helper()
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected pgx.ErrNoRows, got %v", err)
	}
}

func TestCreateLink(t *testing.T) {
	store := newTestStore(t)
	q := store.q

	u1 := store.mkUser(t, "create-u1")
	u2 := store.mkUser(t, "create-u2")
	deleted := store.mkLink(t, u2.ID, "deleted-code")
	if _, err := q.SetLinkDeleted(context.Background(), db.SetLinkDeletedParams{Code: deleted.Code, UserID: u2.ID}); err != nil {
		t.Fatalf("seed delete link: %v", err)
	}
	if _, err := q.CreateLink(context.Background(), db.CreateLinkParams{UserID: u1.ID, Code: "dup", OriginalUrl: "https://example.com/dup"}); err != nil {
		t.Fatalf("seed dup link: %v", err)
	}

	tests := []struct {
		name       string
		arg        db.CreateLinkParams
		wantErr    bool
		wantPgCode string
		check      func(t *testing.T, got db.Link)
	}{
		{
			name: "happy",
			arg:  db.CreateLinkParams{UserID: u1.ID, Code: "ok", OriginalUrl: "https://example.com/ok"},
			check: func(t *testing.T, got db.Link) {
				if got.ID == 0 || got.UserID != u1.ID || got.Code != "ok" || got.Clicks != 0 {
					t.Fatalf("unexpected link: %+v", got)
				}
			},
		},
		{name: "userId doesnot exist", arg: db.CreateLinkParams{UserID: 999999, Code: "no-user", OriginalUrl: "https://x"}, wantErr: true, wantPgCode: "23503"},
		{name: "code is duplicated", arg: db.CreateLinkParams{UserID: u1.ID, Code: "dup", OriginalUrl: "https://x"}, wantErr: true, wantPgCode: "23505"},
		{name: "code duplicated soft deleted", arg: db.CreateLinkParams{UserID: u2.ID, Code: "deleted-code", OriginalUrl: "https://x"}, wantErr: true, wantPgCode: "23505"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := q.CreateLink(context.Background(), tc.arg)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if tc.wantPgCode != "" {
					expectPgCode(t, err, tc.wantPgCode)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	store := newTestStore(t)
	q := store.q
	u := store.mkUser(t, "del-user")

	tests := []struct {
		name    string
		id      int32
		wantErr bool
		check   func(t *testing.T, got db.User)
	}{
		{name: "happy", id: u.ID, check: func(t *testing.T, got db.User) {
			if got.ID != u.ID {
				t.Fatalf("unexpected deleted user: %+v", got)
			}
		}},
		{name: "user doesnot exist", id: 999999, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := q.DeleteUser(context.Background(), tc.id)
			if tc.wantErr {
				expectNoRows(t, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}

func TestGetLinkByCode(t *testing.T) {
	store := newTestStore(t)
	q := store.q
	u := store.mkUser(t, "get-link-user")
	seed := store.mkLink(t, u.ID, "find-me")
	deleted := store.mkLink(t, u.ID, "find-me-deleted")
	if _, err := q.SetLinkDeleted(context.Background(), db.SetLinkDeletedParams{Code: deleted.Code, UserID: u.ID}); err != nil {
		t.Fatalf("seed deleted link: %v", err)
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{name: "happy", code: "find-me"},
		{name: "soft deleted link is hidden", code: deleted.Code, wantErr: true},
		{name: "link doesnot exist", code: "missing", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := q.GetLinkByCode(context.Background(), tc.code)
			if tc.wantErr {
				expectNoRows(t, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != seed.ID || got.Code != seed.Code {
				t.Fatalf("unexpected link: %+v", got)
			}
		})
	}
}

func TestGetLinkByCodeWithDeleted(t *testing.T) {
	store := newTestStore(t)
	q := store.q
	u := store.mkUser(t, "get-link-with-deleted-user")
	active := store.mkLink(t, u.ID, "active-code")
	deleted := store.mkLink(t, u.ID, "deleted-code")
	if _, err := q.SetLinkDeleted(context.Background(), db.SetLinkDeletedParams{Code: deleted.Code, UserID: u.ID}); err != nil {
		t.Fatalf("seed deleted link: %v", err)
	}

	tests := []struct {
		name    string
		code    string
		wantErr bool
		wantID  int32
	}{
		{name: "active link", code: active.Code, wantID: active.ID},
		{name: "deleted link still returned", code: deleted.Code, wantID: deleted.ID},
		{name: "missing link", code: "missing", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := q.GetLinkByCodeWithDeleted(context.Background(), tc.code)
			if tc.wantErr {
				expectNoRows(t, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != tc.wantID {
				t.Fatalf("unexpected link: %+v", got)
			}
		})
	}
}

func TestGetLinksByUserID(t *testing.T) {
	store := newTestStore(t)
	q := store.q
	u1 := store.mkUser(t, "links-u1")
	u2 := store.mkUser(t, "links-u2")
	l1 := store.mkLink(t, u1.ID, "u1-1")
	l2 := store.mkLink(t, u1.ID, "u1-2")
	l3 := store.mkLink(t, u1.ID, "u1-3")
	_ = l1
	_ = l2
	if _, err := q.SetLinkDeleted(context.Background(), db.SetLinkDeletedParams{Code: l2.Code, UserID: u1.ID}); err != nil {
		t.Fatalf("seed deleted link: %v", err)
	}
	_ = store.mkLink(t, u2.ID, "u2-1")

	tests := []struct {
		name      string
		arg       db.GetLinksByUserIDParams
		wantCodes []string
	}{
		{
			name:      "returns active links desc",
			arg:       db.GetLinksByUserIDParams{UserID: u1.ID, ID: 1<<31 - 1, Limit: 10},
			wantCodes: []string{l3.Code, l1.Code},
		},
		{
			name:      "respects cursor and limit",
			arg:       db.GetLinksByUserIDParams{UserID: u1.ID, ID: l3.ID, Limit: 1},
			wantCodes: []string{l1.Code},
		},
		{
			name:      "other user isolated",
			arg:       db.GetLinksByUserIDParams{UserID: u2.ID, ID: 1<<31 - 1, Limit: 10},
			wantCodes: []string{"u2-1"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := q.GetLinksByUserID(context.Background(), tc.arg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.wantCodes) {
				t.Fatalf("expected %d links, got %d", len(tc.wantCodes), len(got))
			}
			for i, code := range tc.wantCodes {
				if got[i].Code != code {
					t.Fatalf("unexpected code at %d: got %q want %q", i, got[i].Code, code)
				}
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	store := newTestStore(t)
	q := store.q
	u := store.mkUser(t, "get-user")

	tests := []struct {
		name    string
		id      int32
		wantErr bool
	}{
		{name: "happy", id: u.ID},
		{name: "missing user", id: 999999, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := q.GetUserByID(context.Background(), tc.id)
			if tc.wantErr {
				expectNoRows(t, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != u.ID {
				t.Fatalf("unexpected user: %+v", got)
			}
		})
	}
}

func TestResetDb(t *testing.T) {
	store := newTestStore(t)
	q := store.q
	u := store.mkUser(t, "reset-user")
	_ = store.mkLink(t, u.ID, "reset-link")

	if err := q.ResetDb(context.Background()); err != nil {
		t.Fatalf("reset db: %v", err)
	}

	if _, err := q.GetUserByID(context.Background(), u.ID); !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf("expected no rows after reset, got %v", err)
	}
}

func TestSetLinkClicked(t *testing.T) {
	store := newTestStore(t)
	q := store.q
	u := store.mkUser(t, "clicked-user")
	l := store.mkLink(t, u.ID, "clicked-code")
	deleted := store.mkLink(t, u.ID, "clicked-deleted")
	if _, err := q.SetLinkDeleted(context.Background(), db.SetLinkDeletedParams{Code: deleted.Code, UserID: u.ID}); err != nil {
		t.Fatalf("seed deleted link: %v", err)
	}

	tests := []struct {
		name       string
		code       string
		wantErr    bool
		wantClicks int32
	}{
		{name: "increments active link", code: l.Code, wantClicks: 1},
		{name: "missing link", code: "missing", wantErr: true},
		{name: "deleted link", code: deleted.Code, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := q.SetLinkClicked(context.Background(), tc.code)
			if tc.wantErr {
				expectNoRows(t, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantClicks {
				t.Fatalf("unexpected clicks: got %d want %d", got, tc.wantClicks)
			}
		})
	}
}

func TestSetLinkDeleted(t *testing.T) {
	store := newTestStore(t)
	q := store.q
	u1 := store.mkUser(t, "delete-link-u1")
	u2 := store.mkUser(t, "delete-link-u2")
	l := store.mkLink(t, u1.ID, "delete-link")
	deleted := store.mkLink(t, u1.ID, "already-deleted")
	if _, err := q.SetLinkDeleted(context.Background(), db.SetLinkDeletedParams{Code: deleted.Code, UserID: u1.ID}); err != nil {
		t.Fatalf("seed deleted link: %v", err)
	}

	tests := []struct {
		name    string
		arg     db.SetLinkDeletedParams
		wantErr bool
		wantID  int32
	}{
		{name: "happy", arg: db.SetLinkDeletedParams{Code: l.Code, UserID: u1.ID}, wantID: l.ID},
		{name: "wrong user", arg: db.SetLinkDeletedParams{Code: l.Code, UserID: u2.ID}, wantErr: true},
		{name: "already deleted", arg: db.SetLinkDeletedParams{Code: deleted.Code, UserID: u1.ID}, wantErr: true},
		{name: "missing link", arg: db.SetLinkDeletedParams{Code: "missing", UserID: u1.ID}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := q.SetLinkDeleted(context.Background(), tc.arg)
			if tc.wantErr {
				expectNoRows(t, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantID {
				t.Fatalf("unexpected id: got %d want %d", got, tc.wantID)
			}
		})
	}
}

func TestUpsertUser(t *testing.T) {
	store := newTestStore(t)
	q := store.q

	name1 := "first"
	pic1 := "https://example.com/first"
	name2 := "second"
	pic2 := "https://example.com/second"

	tests := []struct {
		name string
		arg  db.UpsertUserParams
	}{
		{
			name: "insert new user",
			arg: db.UpsertUserParams{
				Provider:    db.ProviderEnumGOOGLE,
				ProviderID:  "provider-1",
				DisplayName: &name1,
				ProfilePic:  &pic1,
			},
		},
		{
			name: "update existing user",
			arg: db.UpsertUserParams{
				Provider:    db.ProviderEnumGOOGLE,
				ProviderID:  "provider-1",
				DisplayName: &name2,
				ProfilePic:  &pic2,
			},
		},
	}

	var firstID int32
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := q.UpsertUser(context.Background(), tc.arg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.Provider != tc.arg.Provider || got.ProviderID != tc.arg.ProviderID {
				t.Fatalf("unexpected user identity: %+v", got)
			}

			if tc.name == "insert new user" {
				firstID = got.ID
			} else if got.ID != firstID {
				t.Fatalf("expected same user id on upsert, got %d want %d", got.ID, firstID)
			}
		})
	}
}

func TestConcurrentUpsertUser(t *testing.T) {
	store := newTestStore(t)
	q := store.q

	name := "concurrent"
	pic := "https://example.com/concurrent"
	errCh := make(chan error, 2)

	run := func(providerID string) {
		pool, err := db.NewPool(os.Getenv("TEST_DB_URL"))
		if err != nil {
			errCh <- err
			return
		}
		defer db.ClosePool(pool)

		q2 := db.New(pool)
		_, err = q2.UpsertUser(context.Background(), db.UpsertUserParams{
			Provider:    db.ProviderEnumGOOGLE,
			ProviderID:  providerID,
			DisplayName: &name,
			ProfilePic:  &pic,
		})
		errCh <- err
	}

	go run("same-provider-id")
	go run("same-provider-id")

	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			t.Fatalf("concurrent upsert failed: %v", err)
		}
	}

	got, err := q.GetLinksByUserID(context.Background(), db.GetLinksByUserIDParams{UserID: 1, ID: 1<<31 - 1, Limit: 10})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		_ = got
	}
}

func TestProviderEnumScanAndValue(t *testing.T) {
	var e db.ProviderEnum
	if err := e.Scan("GOOGLE"); err != nil {
		t.Fatalf("scan string: %v", err)
	}
	if e != db.ProviderEnumGOOGLE {
		t.Fatalf("unexpected enum after string scan: %q", e)
	}

	if err := e.Scan([]byte("GITHUB")); err != nil {
		t.Fatalf("scan bytes: %v", err)
	}
	if e != db.ProviderEnumGITHUB {
		t.Fatalf("unexpected enum after bytes scan: %q", e)
	}

	if err := e.Scan(123); err == nil {
		t.Fatalf("expected scan error for invalid type")
	}

	valid := db.NullProviderEnum{ProviderEnum: db.ProviderEnumGOOGLE, Valid: true}
	v, err := valid.Value()
	if err != nil {
		t.Fatalf("value valid: %v", err)
	}
	if v != string(db.ProviderEnumGOOGLE) {
		t.Fatalf("unexpected driver value: %v", v)
	}

	invalid := db.NullProviderEnum{}
	v, err = invalid.Value()
	if err != nil {
		t.Fatalf("value invalid: %v", err)
	}
	if v != nil {
		t.Fatalf("expected nil value for invalid enum, got %v", v)
	}
}

func TestProviderEnumStringFormatting(t *testing.T) {
	if got := string(db.ProviderEnumGOOGLE); got != "GOOGLE" {
		t.Fatalf("unexpected string formatting: %q", got)
	}
}
