package db_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
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

func mkUser(t *testing.T, q *db.Queries, providerID string) db.User {
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

func mkLink(t *testing.T, q *db.Queries, userID int32, code string) db.Link {
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

func runBeforeEachReset(t *testing.T, q *db.Queries) {
	t.Helper()
	if err := q.ResetDb(context.Background()); err != nil {
		t.Fatalf("setup test db: %v", err)
	}
}

func TestCreateLink(t *testing.T) {
	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)

	u1 := mkUser(t, q, "create-u1")
	u2 := mkUser(t, q, "create-u2")
	deleted := mkLink(t, q, u2.ID, "deleted-code")
	if _, err := q.SetLinkDeleted(context.Background(), deleted.ID); err != nil {
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
	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)
	u := mkUser(t, q, "del-user")

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
	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)
	u := mkUser(t, q, "get-link-user")
	seed := mkLink(t, q, u.ID, "find-me")

	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{name: "happy", code: "find-me"},
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

func TestGetLinksByUserID(t *testing.T) {
	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)
	u := mkUser(t, q, "list-user")
	created := make([]db.Link, 0, 5)
	for i := 0; i < 5; i++ {
		created = append(created, mkLink(t, q, u.ID, fmt.Sprintf("c-%d", i)))
	}

	if len(created) != 5 {
		t.Fatalf("expected 5 created links, got %d", len(created))
	}
	expectedDesc := []int32{created[4].ID, created[3].ID, created[2].ID, created[1].ID, created[0].ID}

	page1, err := q.GetLinksByUserID(context.Background(), db.GetLinksByUserIDParams{UserID: u.ID, ID: 2147483647, Limit: 2})
	if err != nil {
		t.Fatalf("get page1: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected page1 len 2, got %d", len(page1))
	}
	if page1[0].ID != expectedDesc[0] || page1[1].ID != expectedDesc[1] {
		t.Fatalf("unexpected page1 ids: got [%d %d], want [%d %d]", page1[0].ID, page1[1].ID, expectedDesc[0], expectedDesc[1])
	}
	if !(page1[0].ID > page1[1].ID) {
		t.Fatalf("page1 not descending: %+v", page1)
	}

	cursor := page1[len(page1)-1].ID
	page2, err := q.GetLinksByUserID(context.Background(), db.GetLinksByUserIDParams{UserID: u.ID, ID: cursor, Limit: 2})
	if err != nil {
		t.Fatalf("get page2: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("expected page2 len 2, got %d", len(page2))
	}
	if page2[0].ID != expectedDesc[2] || page2[1].ID != expectedDesc[3] {
		t.Fatalf("unexpected page2 ids: got [%d %d], want [%d %d]", page2[0].ID, page2[1].ID, expectedDesc[2], expectedDesc[3])
	}
	if !(page2[0].ID > page2[1].ID) {
		t.Fatalf("page2 not descending: %+v", page2)
	}
	for _, l := range page2 {
		if l.ID >= cursor {
			t.Fatalf("cursor boundary broken: link id %d should be < cursor %d", l.ID, cursor)
		}
	}
	if page1[0].ID == page2[0].ID || page1[0].ID == page2[1].ID || page1[1].ID == page2[0].ID || page1[1].ID == page2[1].ID {
		t.Fatalf("page overlap detected: page1=%v page2=%v", []int32{page1[0].ID, page1[1].ID}, []int32{page2[0].ID, page2[1].ID})
	}

	page3, err := q.GetLinksByUserID(context.Background(), db.GetLinksByUserIDParams{UserID: u.ID, ID: page2[len(page2)-1].ID, Limit: 2})
	if err != nil {
		t.Fatalf("get page3: %v", err)
	}
	if len(page3) != 1 {
		t.Fatalf("expected page3 len 1, got %d", len(page3))
	}
	if page3[0].ID != expectedDesc[4] {
		t.Fatalf("unexpected page3 id: got %d, want %d", page3[0].ID, expectedDesc[4])
	}

	allLimited3, err := q.GetLinksByUserID(context.Background(), db.GetLinksByUserIDParams{UserID: u.ID, ID: 2147483647, Limit: 3})
	if err != nil {
		t.Fatalf("get limit 3: %v", err)
	}
	if len(allLimited3) != 3 {
		t.Fatalf("expected len 3 with limit 3, got %d", len(allLimited3))
	}

	all, err := q.GetLinksByUserID(context.Background(), db.GetLinksByUserIDParams{UserID: u.ID, ID: 2147483647, Limit: 20})
	if err != nil {
		t.Fatalf("get full page: %v", err)
	}
	if len(all) != 5 {
		t.Fatalf("expected full len 5, got %d", len(all))
	}
	for i := range expectedDesc {
		if all[i].ID != expectedDesc[i] {
			t.Fatalf("unexpected full page order at %d: got %d, want %d", i, all[i].ID, expectedDesc[i])
		}
	}

	empty, err := q.GetLinksByUserID(context.Background(), db.GetLinksByUserIDParams{UserID: u.ID, ID: 1, Limit: 20})
	if err != nil {
		t.Fatalf("get empty page: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected empty result, got %d", len(empty))
	}

	notFoundUser, err := q.GetLinksByUserID(context.Background(), db.GetLinksByUserIDParams{UserID: 999999, ID: 2147483647, Limit: 20})
	if err != nil {
		t.Fatalf("get user not found page: %v", err)
	}
	if len(notFoundUser) != 0 {
		t.Fatalf("expected empty for missing user, got %d", len(notFoundUser))
	}
}

func TestGetUserByID(t *testing.T) {
	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)
	u := mkUser(t, q, "get-user")

	tests := []struct {
		name    string
		id      int32
		wantErr bool
	}{
		{name: "happy", id: u.ID},
		{name: "not exist", id: 999999, wantErr: true},
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
			if got.ID != u.ID || got.ProviderID != u.ProviderID {
				t.Fatalf("unexpected user: %+v", got)
			}
		})
	}
}

func TestResetDb(t *testing.T) {
	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)
	tests := []struct {
		name    string
		prepare func(t *testing.T)
		exec    func(t *testing.T) error
		wantErr bool
	}{
		{
			name: "happy",
			prepare: func(t *testing.T) {
				u := mkUser(t, q, "reset-u")
				_ = mkLink(t, q, u.ID, "reset-l")
			},
			exec: func(t *testing.T) error {
				return q.ResetDb(context.Background())
			},
		},
		{
			name: "db is down",
			exec: func(t *testing.T) error {
				testDbURL, err := dep.GetEnvStrOrError("TEST_DB_URL")
				if err != nil {
					t.Fatalf("read TEST_DB_URL: %v", err)
				}
				pool, err := db.NewPool(testDbURL)
				if err != nil {
					t.Fatalf("new pool: %v", err)
				}
				q2 := db.New(pool)
				db.ClosePool(pool)
				return q2.ResetDb(context.Background())
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}
			err := tc.exec(t)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			_, err = q.GetUserByID(context.Background(), 1)
			expectNoRows(t, err)
		})
	}
}

func TestSetLinkClicked(t *testing.T) {
	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)
	u := mkUser(t, q, "click-user")
	l0 := mkLink(t, q, u.ID, "c0")
	l8 := mkLink(t, q, u.ID, "c8")
	for i := 0; i < 8; i++ {
		if _, err := q.SetLinkClicked(context.Background(), l8.ID); err != nil {
			t.Fatalf("seed click: %v", err)
		}
	}

	tests := []struct {
		name      string
		id        int32
		wantErr   bool
		wantClick int32
	}{
		{name: "happy count=0", id: l0.ID, wantClick: 1},
		{name: "happy count=8", id: l8.ID, wantClick: 9},
		{name: "link doesnot exist", id: 999999, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			click, err := q.SetLinkClicked(context.Background(), tc.id)
			if tc.wantErr {
				expectNoRows(t, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if click != tc.wantClick {
				t.Fatalf("expected click %d, got %d", tc.wantClick, click)
			}
		})
	}
}

func TestSetLinkDeleted(t *testing.T) {
	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)
	u := mkUser(t, q, "delete-link-user")
	active := mkLink(t, q, u.ID, "active")
	gone := mkLink(t, q, u.ID, "gone")
	if _, err := q.SetLinkDeleted(context.Background(), gone.ID); err != nil {
		t.Fatalf("seed deleted link: %v", err)
	}

	tests := []struct {
		name    string
		id      int32
		wantErr bool
		wantID  int32
	}{
		{name: "happy", id: active.ID, wantID: active.ID},
		{name: "link doesnot exist", id: 999999, wantErr: true},
		{name: "link has been deleted", id: gone.ID, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			id, err := q.SetLinkDeleted(context.Background(), tc.id)
			if tc.wantErr {
				expectNoRows(t, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if id != tc.wantID {
				t.Fatalf("expected id %d, got %d", tc.wantID, id)
			}
		})
	}
}

func TestUpsertUser(t *testing.T) {
	q := sharedQueries
	if q == nil {
		t.Fatalf("setup test db: shared queries not initialized")
	}
	runBeforeEachReset(t, q)
	firstName := "first"
	firstPic := "https://example.com/first"
	newName := "updated"
	newPic := "https://example.com/updated"

	tests := []struct {
		name    string
		prepare func(t *testing.T)
		arg     db.UpsertUserParams
		check   func(t *testing.T, got db.User)
	}{
		{
			name: "happy user not exist",
			arg: db.UpsertUserParams{
				Provider:    db.ProviderEnumGITHUB,
				ProviderID:  "u-new",
				DisplayName: &firstName,
				ProfilePic:  &firstPic,
			},
			check: func(t *testing.T, got db.User) {
				if got.ID == 0 || got.ProviderID != "u-new" {
					t.Fatalf("unexpected user: %+v", got)
				}
			},
		},
		{
			name: "happy existing",
			prepare: func(t *testing.T) {
				_, err := q.UpsertUser(context.Background(), db.UpsertUserParams{
					Provider:    db.ProviderEnumGITHUB,
					ProviderID:  "u-old",
					DisplayName: &firstName,
					ProfilePic:  &firstPic,
				})
				if err != nil {
					t.Fatalf("seed user: %v", err)
				}
			},
			arg: db.UpsertUserParams{
				Provider:    db.ProviderEnumGITHUB,
				ProviderID:  "u-old",
				DisplayName: &newName,
				ProfilePic:  &newPic,
			},
			check: func(t *testing.T, got db.User) {
				if got.ProviderID != "u-old" {
					t.Fatalf("unexpected user: %+v", got)
				}
				if got.DisplayName == nil || *got.DisplayName != newName {
					t.Fatalf("expected updated name, got %+v", got)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.prepare != nil {
				tc.prepare(t)
			}
			got, err := q.UpsertUser(context.Background(), tc.arg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}
