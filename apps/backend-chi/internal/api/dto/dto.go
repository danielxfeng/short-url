package dto

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	sqlcdb "github.com/danielxfeng/short-url/apps/backend-chi/internal/api/db/sqlc"
)

var Validate *validator.Validate

func InitValidator() {
	Validate = validator.New()
	_ = Validate.RegisterValidation("trim", trimValue) // SIDE EFFECT: trims the value
}

// Space Trimming, SIDE EFFECT!
func trimValue(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {

	case reflect.String:
		trimmed := strings.TrimSpace(field.String())
		field.SetString(trimmed)

	case reflect.Pointer:
		if field.IsNil() {
			return true
		}

		elem := field.Elem()
		if elem.Kind() != reflect.String {
			return false
		}

		trimmed := strings.TrimSpace(elem.String())
		elem.SetString(trimmed)

	default:
		return false
	}

	return true
}

type APIErrorRes struct {
	Error string `json:"error"`
}

type UpsertUserReq struct {
	Provider    sqlcdb.ProviderEnum `json:"provider" validate:"required,trim,oneof=GOOGLE GITHUB"`
	ProviderID  string              `json:"provider_id" validate:"required,trim,min=1,max=255"`
	DisplayName *string             `json:"display_name" validate:"omitempty,trim,min=1,max=255"`
	ProfilePic  *string             `json:"profile_pic" validate:"omitempty,trim,min=1,max=255"`
}

type UserResponse struct {
	ID          int32               `json:"id"`
	Provider    sqlcdb.ProviderEnum `json:"provider"`
	ProviderID  string              `json:"provider_id"`
	DisplayName *string             `json:"display_name,omitempty"`
	ProfilePic  *string             `json:"profile_pic,omitempty"`
}

type UserWithTokenResponse struct {
	UserResponse
	Token string `json:"token"`
}

type CreateLinkReq struct {
	OriginalUrl string `json:"original_url" validate:"required,trim,url"`
}

type LinkResponse struct {
	ID          int32  `json:"id"`
	Code        string `json:"code"`
	OriginalUrl string `json:"original_url"`
	Clicks      int32  `json:"clicks"`
	CreatedAt   string `json:"created_at"`
	IsDeleted   bool   `json:"is_deleted"`
}

type GetLinksReq struct {
	Limit  *int32 `json:"limit,omitempty"  validate:"omitempty,min=1,max=20"`
	Cursor *int32 `json:"cursor,omitempty" validate:"omitempty"`
}

type LinksResponse struct {
	Links   []LinkResponse `json:"links"`
	HasMore bool           `json:"has_more"`
	Cursor  *int32         `json:"cursor,omitempty"`
}
