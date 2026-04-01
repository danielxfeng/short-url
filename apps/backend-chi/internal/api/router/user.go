package router

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/auth"
	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
	stateStore "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/statestore"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/dto"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/mymiddleware"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/util"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int32) (db.User, error)
	DeleteUser(ctx context.Context, id int32) (db.User, error)
	UpsertUser(ctx context.Context, arg db.UpsertUserParams) (db.User, error)
}

func UserToDTO(user db.User) dto.UserResponse {
	return dto.UserResponse{
		ID:          user.ID,
		Provider:    user.Provider,
		ProviderID:  user.ProviderID,
		DisplayName: user.DisplayName,
		ProfilePic:  user.ProfilePic,
	}
}

func UserRouter(dep *dep.Dep, repo UserRepository, oauth auth.OauthHandler, store stateStore.StateStore) http.Handler {
	r := chi.NewRouter()

	r.Route("/auth", func(r chi.Router) {
		r.Get("/{provider}", func(w http.ResponseWriter, r *http.Request) {
			provider := chi.URLParam(r, "provider")
			oauthCfg, ok := oauth.GetConfigForProvider(provider)
			if !ok {
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "unsupported provider"), http.StatusFound)
				return
			}

			url := oauth.GetOauthAuthURL(&oauthCfg.Config, store)
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

			oauthCfg, ok := oauth.GetConfigForProvider(provider)
			if !ok {
				dep.Logger.Warn("unsupported provider", "provider", provider)
				http.Redirect(w, r, util.AssembleURL(dep.Cfg.FrontendRedirectURL, "error", "unsupported provider"), http.StatusFound)
				return
			}

			client, err := oauth.ExchangeCodeAndGetClient(r.Context(), &oauthCfg.Config, code, verifier)
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

			token, err := auth.GenerateToken(user.ID, dep.Cfg.JWTSecret, dep.Cfg.JWTExpiry)
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
