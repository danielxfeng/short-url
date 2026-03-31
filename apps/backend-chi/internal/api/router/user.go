package router

import (
	"context"
	"errors"
	"net/http"
	"sync"

	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/mymiddleware"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/util"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type StateStore struct {
	mu     sync.Mutex
	Status map[string]string // state -> verifier
}

func newStateStore() *StateStore {
	return &StateStore{
		Status: make(map[string]string),
	}
}

func (s *StateStore) Add(state string, verifier string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Status[state] = verifier
}

func (s *StateStore) GetAndDelete(state string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	verifier, exists := s.Status[state]
	if exists {
		delete(s.Status, state) // State is single-use, delete it after validating
	}
	return verifier
}

var store = newStateStore() // In-memory store for OAuth states, may use redis or db in production

type UserRepository interface {
	GetUserByID(ctx context.Context, id int32) (db.User, error)
	DeleteUser(ctx context.Context, id int32) (db.User, error)
	UpsertUser(ctx context.Context, arg db.UpsertUserParams) (db.User, error)
}

func UserRouter(dep *dep.Dep, repo UserRepository, helper OauthHelper) http.Handler {
	r := chi.NewRouter()

	r.Route("/auth", func(r chi.Router) {
		r.Get("/{provider}", func(w http.ResponseWriter, r *http.Request) {
			provider := chi.URLParam(r, "provider")
			oauthCfg, ok := helper.GetConfigForProvider(provider)
			if !ok {
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "unsupported provider"), http.StatusFound)
				return
			}

			url := helper.GetOauthAuthURL(&oauthCfg.Config, store)
			http.Redirect(w, r, url, http.StatusFound)
		})

		r.Get("/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
			provider := chi.URLParam(r, "provider")
			code := r.URL.Query().Get("code")
			state := r.URL.Query().Get("state")

			if code == "" {
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "code not found in query"), http.StatusFound)
				return
			}

			if state == "" {
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "invalid state"), http.StatusFound)
				return
			}

			verifier := store.GetAndDelete(state)
			if verifier == "" {
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "invalid state"), http.StatusFound)
				return
			}

			oauthCfg, ok := helper.GetConfigForProvider(provider)
			if !ok {
				dep.Logger.Warn("unsupported provider", "provider", provider)
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "unsupported provider"), http.StatusFound)
				return
			}

			client, err := helper.ExchangeCodeAndGetClient(r.Context(), &oauthCfg.Config, code, verifier)
			if err != nil {
				dep.Logger.Warn("failed to exchange code for token", "error", err)
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "failed to exchange code for token"), http.StatusFound)
				return
			}

			userInfo, err := oauthCfg.GetUserInfo(client)
			if err != nil {
				dep.Logger.Warn("failed to get user info", "error", err)
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "failed to get user info"), http.StatusFound)
				return
			}

			if userInfo == nil {
				dep.Logger.Error("failed to get user info", "error", "user info is nil")
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "failed to get user info"), http.StatusFound)
				return
			}

			user, err := repo.UpsertUser(r.Context(), *userInfo)
			if err != nil {
				dep.Logger.Error("failed to upsert user", "error", err)
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "failed to upsert user"), http.StatusFound)
				return
			}

			token, err := util.GenerateToken(user.ID, dep.Cfg.JWTSecret, dep.Cfg.JWTExpiry)
			if err != nil {
				dep.Logger.Error("failed to generate token", "error", err)
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "failed to generate token"), http.StatusFound)
				return
			}

			http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "auth", token), http.StatusFound)
		})
	})

	r.Route("/me", func(r chi.Router) {
		r.Use(mymiddleware.Auth(dep.Cfg.JWTSecret))

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(mymiddleware.UserIDContextKey).(int32)

			user, err := repo.GetUserByID(r.Context(), userID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					http.Error(w, "user not found", http.StatusNotFound)
					return
				}
				panic(err)
			}

			util.SendJSON(w, http.StatusOK, UserToDTO(user))
		})

		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(mymiddleware.UserIDContextKey).(int32)

			_, err := repo.DeleteUser(r.Context(), userID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					http.Error(w, "user not found", http.StatusNotFound)
					return
				}
				panic(err)
			}

			w.WriteHeader(http.StatusNoContent)
		})
	})

	return r
}
