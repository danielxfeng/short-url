package dto

import (
	"reflect"
	"strings"
	"time"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/models"
	"github.com/go-playground/validator/v10"
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
	Provider    models.ProviderEnum `json:"provider" validate:"required,trim,oneof=GOOGLE GITHUB"`
	ProviderID  string              `json:"provider_id" validate:"required,trim,min=1,max=255"`
	DisplayName *string             `json:"display_name" validate:"omitempty,trim,min=1,max=255"`
	ProfilePic  *string             `json:"profile_pic" validate:"omitempty,trim,min=1,max=255"`
}

type UserResponse struct {
	ID          int32               `json:"id"`
	Provider    models.ProviderEnum `json:"provider"`
	ProviderID  string              `json:"provider_id"`
	DisplayName *string             `json:"display_name,omitempty"`
	ProfilePic  *string             `json:"profile_pic,omitempty"`
}

type UserWithTokenResponse struct {
	UserResponse
	Token string `json:"token"`
}

type CreateLinkReq struct {
	OriginalUrl string  `json:"original_url" validate:"required,trim,url"`
	Code        *string `json:"code,omitempty" validate:"omitempty,trim,min=1,max=255"`
	Note        *string `json:"note,omitempty" validate:"omitempty,trim,min=1,max=255"`
}

type LinkResponse struct {
	ID          int32     `json:"id"`
	Code        string    `json:"code"`
	OriginalUrl string    `json:"original_url"`
	Note        *string   `json:"note,omitempty"`
	Clicks      int32     `json:"clicks"`
	CreatedAt   time.Time `json:"created_at"`
	IsDeleted   bool      `json:"is_deleted"`
}

type LinksResponse struct {
	Links   []LinkResponse `json:"links"`
	HasMore bool           `json:"has_more"`
	Cursor  *int32         `json:"cursor,omitempty"`
}
