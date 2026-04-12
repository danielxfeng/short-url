package dto

import (
	"net"
	"net/netip"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/models"
	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate
var shortCodePattern = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

func InitValidator() {
	Validate = validator.New()
	_ = Validate.RegisterValidation("trim", trimValue) // SIDE EFFECT: trims the value
	_ = Validate.RegisterValidation("shortcode", shortCodeValue)
	_ = Validate.RegisterValidation("safe_target_url", safeTargetURLValue)
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

func shortCodeValue(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.String:
		return shortCodePattern.MatchString(field.String())
	case reflect.Pointer:
		if field.IsNil() {
			return true
		}

		elem := field.Elem()
		if elem.Kind() != reflect.String {
			return false
		}

		return shortCodePattern.MatchString(elem.String())
	default:
		return false
	}
}

func safeTargetURLValue(fl validator.FieldLevel) bool {
	field := fl.Field()
	if field.Kind() != reflect.String {
		return false
	}

	raw := field.String()
	parsed, err := url.Parse(raw)
	if err != nil {
		return false
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return false
	}

	if parsed.User != nil {
		return false
	}

	host := parsed.Hostname()
	if host == "" {
		return false
	}

	normalizedHost := strings.TrimSuffix(strings.ToLower(host), ".")
	if isBlockedLocalHostname(normalizedHost) {
		return false
	}

	if addr, err := netip.ParseAddr(host); err == nil {
		return isAllowedPublicIP(addr)
	}

	return true
}

func isBlockedLocalHostname(host string) bool {
	if host == "localhost" || host == "local" {
		return true
	}

	blockedSuffixes := []string{
		".localhost",
		".local",
		".internal",
		".home",
		".lan",
	}

	for _, suffix := range blockedSuffixes {
		if strings.HasSuffix(host, suffix) {
			return true
		}
	}

	return false
}

func isAllowedPublicIP(addr netip.Addr) bool {
	if addr.IsLoopback() || addr.IsPrivate() || addr.IsMulticast() || addr.IsLinkLocalUnicast() ||
		addr.IsLinkLocalMulticast() || addr.IsUnspecified() {
		return false
	}

	ip := net.ParseIP(addr.String())
	if ip == nil {
		return false
	}

	blockedCIDRs := []string{
		"100.64.0.0/10", // carrier-grade NAT
		"198.18.0.0/15", // benchmarking
	}

	for _, cidr := range blockedCIDRs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			return false
		}
		if network.Contains(ip) {
			return false
		}
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
	OriginalUrl string  `json:"original_url" validate:"required,trim,safe_target_url"`
	Code        *string `json:"code,omitempty" validate:"omitempty,trim,min=1,max=255,shortcode"`
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
