package models

import (
	"context"

	db "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/db"
)

type ProviderEnum = db.ProviderEnum

const (
	ProviderEnumGOOGLE = db.ProviderEnumGOOGLE
	ProviderEnumGITHUB = db.ProviderEnumGITHUB
)

type User = db.User
type UpsertUserParams = db.UpsertUserParams

type UserRepository interface {
	GetUserByID(ctx context.Context, id int32) (User, error)
	DeleteUser(ctx context.Context, id int32) (User, error)
	UpsertUser(ctx context.Context, arg UpsertUserParams) (User, error)
}
