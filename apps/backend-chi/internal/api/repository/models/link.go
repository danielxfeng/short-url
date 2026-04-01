package models

import (
	"context"

	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/db"
)

type Link = db.Link
type CreateLinkParams = db.CreateLinkParams
type GetLinksByUserIDParams = db.GetLinksByUserIDParams
type SetLinkDeletedParams = db.SetLinkDeletedParams

type LinkRepository interface {
	CreateLink(ctx context.Context, arg CreateLinkParams) (Link, error)
	GetLinkByCode(ctx context.Context, code string) (Link, error)
	GetLinkByCodeWithDeleted(ctx context.Context, code string) (Link, error)
	GetLinksByUserID(ctx context.Context, arg GetLinksByUserIDParams) ([]Link, error)
	SetLinkDeleted(ctx context.Context, arg SetLinkDeletedParams) (int32, error)
	SetLinkClicked(ctx context.Context, code string) (int32, error)
}
