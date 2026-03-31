package router

import (
	"context"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/apierror"
	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/dto"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/mymiddleware"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/util"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

const MAX_LIMIT = 200
const MIN_LIMIT = 1
const DEFAULT_LIMIT = 20
const MAX_CURSOR = math.MaxInt32 - 1
const MIN_CURSOR = 0
const DEFAULT_CURSOR = MAX_CURSOR
const MAX_CONFLICT_RETRIES = 5
const LENGTH_CODE = 8

type ShortURLRepository interface {
	CreateLink(ctx context.Context, arg db.CreateLinkParams) (db.Link, error)
	GetLinkByCode(ctx context.Context, code string) (db.Link, error)
	GetLinkByCodeWithDeleted(ctx context.Context, code string) (db.Link, error)
	GetLinksByUserID(ctx context.Context, arg db.GetLinksByUserIDParams) ([]db.Link, error)
	SetLinkDeleted(ctx context.Context, arg db.SetLinkDeletedParams) (int32, error)
	SetLinkClicked(ctx context.Context, code string) (int32, error)
}

func ShortURLRouter(dep *dep.Dep, repo ShortURLRepository) http.Handler {
	r := chi.NewRouter()

	r.Get("/{code}", func(w http.ResponseWriter, r *http.Request) {
		code := chi.URLParam(r, "code")
		link, err := repo.GetLinkByCode(r.Context(), code)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Redirect(w, r, dep.Cfg.NotFoundPage, http.StatusFound)
				return
			}
			panic(err)
		}

		http.Redirect(w, r, link.OriginalUrl, http.StatusFound)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		go func() {
			defer cancel()

			_, err := repo.SetLinkClicked(ctx, link.Code)
			if err != nil {
				dep.Logger.Error("Failed to update click count", "error", err)
			}
		}()
	})

	// All routes below require authentication.
	r.Group(func(r chi.Router) {
		r.Use(mymiddleware.Auth(dep.Cfg.JWTSecret))

		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(mymiddleware.UserIDContextKey).(int32)
			limit := util.ParseInt32ClampedOrDefault(r.URL.Query().Get("limit"), DEFAULT_LIMIT, MIN_LIMIT, MAX_LIMIT)
			cursor := util.ParseInt32ClampedOrDefault(r.URL.Query().Get("cursor"), DEFAULT_CURSOR, MIN_CURSOR, MAX_CURSOR)
			intLimit := int(limit)

			links, err := repo.GetLinksByUserID(r.Context(), db.GetLinksByUserIDParams{
				UserID: userID,
				Limit:  limit + 1, // fetch one extra to check if there's a next page
				ID:     cursor,
			})

			if err != nil {
				panic(err)
			}

			hasNext := len(links) > intLimit // if we got more than the requested limit, there is a next page

			var nextCursor *int32
			if hasNext {
				links = links[:intLimit]             // trim the extra record
				nextCursor = &links[len(links)-1].ID // next cursor is the ID of the last record in the current page
			}

			resp := dto.LinksResponse{
				Links:   LinksToDTO(links),
				HasMore: hasNext,
				Cursor:  nextCursor,
			}
			util.SendJSON(w, http.StatusOK, resp)
		})

		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(mymiddleware.UserIDContextKey).(int32)

			req, err := util.ParseAndValidateJSON(r, func(data *dto.CreateLinkReq) error {
				return dto.Validate.Struct(data)
			})

			if err != nil {
				util.SendError(w, apierror.NewApiError(400, "invalid request body", err))
				return
			}

			retryCount := MAX_CONFLICT_RETRIES
			code := ""

			for retryCount > 0 {
				code = util.GenerateRandomString(LENGTH_CODE)
				_, err := repo.GetLinkByCodeWithDeleted(r.Context(), code)
				if errors.Is(err, pgx.ErrNoRows) {
					break // code is unique
				} else if err != nil {
					panic(err) // unexpected error
				}

				retryCount--
			}

			if retryCount == 0 { // exhausted retries
				util.SendError(w, apierror.NewApiError(500, "failed to generate unique code", nil))
				return
			}

			link, err := repo.CreateLink(r.Context(), db.CreateLinkParams{
				UserID:      userID,
				Code:        code,
				OriginalUrl: req.OriginalUrl,
			})

			if err != nil {
				panic(err)
			}

			util.SendJSON(w, http.StatusCreated, link)
		})

		r.Delete("/{code}", func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(mymiddleware.UserIDContextKey).(int32)

			code := chi.URLParam(r, "code")
			if code == "" {
				util.SendError(w, apierror.NewApiError(400, "code is required", nil))
				return
			}

			_, err := repo.SetLinkDeleted(r.Context(), db.SetLinkDeletedParams{
				Code:   code,
				UserID: userID,
			})
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					util.SendError(w, apierror.NewApiError(404, "link not found", nil))
					return
				}
				panic(err)
			}

			w.WriteHeader(http.StatusNoContent)
		})
	})

	return r
}
