package dto

import (
	"testing"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/repository/models"
	"github.com/go-playground/validator/v10"
)

func newValidatorWithTrim(t *testing.T) *validator.Validate {
	t.Helper()

	v := validator.New()
	if err := v.RegisterValidation("trim", trimValue); err != nil {
		t.Fatalf("register trim validation: %v", err)
	}

	return v
}

func strPtr(s string) *string {
	return &s
}

func TestTrimValue(t *testing.T) {
	v := newValidatorWithTrim(t)

	testCases := []struct {
		name       string
		in         any
		wantPass   bool
		assertFunc func(t *testing.T, got any)
	}{
		{
			name: "trims leading space",
			in: &struct {
				Value string `validate:"trim"`
			}{
				Value: "  hello",
			},
			wantPass: true,
			assertFunc: func(t *testing.T, got any) {
				t.Helper()
				payload := got.(*struct {
					Value string `validate:"trim"`
				})
				if payload.Value != "hello" {
					t.Fatalf("got %q want %q", payload.Value, "hello")
				}
			},
		},
		{
			name: "trims trailing space",
			in: &struct {
				Value string `validate:"trim"`
			}{
				Value: "hello  ",
			},
			wantPass: true,
			assertFunc: func(t *testing.T, got any) {
				t.Helper()
				payload := got.(*struct {
					Value string `validate:"trim"`
				})
				if payload.Value != "hello" {
					t.Fatalf("got %q want %q", payload.Value, "hello")
				}
			},
		},
		{
			name: "preserves inner space",
			in: &struct {
				Value string `validate:"trim"`
			}{
				Value: "hello  world",
			},
			wantPass: true,
			assertFunc: func(t *testing.T, got any) {
				t.Helper()
				payload := got.(*struct {
					Value string `validate:"trim"`
				})
				if payload.Value != "hello  world" {
					t.Fatalf("got %q want %q", payload.Value, "hello  world")
				}
			},
		},
		{
			name: "trims leading and trailing, preserves inner",
			in: &struct {
				Value string `validate:"trim"`
			}{
				Value: "  hello  world  ",
			},
			wantPass: true,
			assertFunc: func(t *testing.T, got any) {
				t.Helper()
				payload := got.(*struct {
					Value string `validate:"trim"`
				})
				if payload.Value != "hello  world" {
					t.Fatalf("got %q want %q", payload.Value, "hello  world")
				}
			},
		},
		{
			name: "trims pointer to string",
			in: &struct {
				Value *string `validate:"trim"`
			}{
				Value: strPtr("  hello  "),
			},
			wantPass: true,
			assertFunc: func(t *testing.T, got any) {
				t.Helper()
				payload := got.(*struct {
					Value *string `validate:"trim"`
				})
				if payload.Value == nil || *payload.Value != "hello" {
					t.Fatalf("got %v want %q", payload.Value, "hello")
				}
			},
		},
		{
			name: "nil pointer with omitempty",
			in: &struct {
				Value *string `validate:"omitempty,trim"`
			}{
				Value: nil,
			},
			wantPass: true,
		},
		{
			name: "pointer to int type fails",
			in: &struct {
				Value *int `validate:"trim"`
			}{
				Value: func() *int {
					v := 42
					return &v
				}(),
			},
			wantPass: false,
		},
		{
			name: "int type fails",
			in: &struct {
				Value int `validate:"trim"`
			}{
				Value: 42,
			},
			wantPass: false,
		},
		{
			name: "whitespace only fails min constraint after trim",
			in: &struct {
				Value string `validate:"trim,min=1"`
			}{
				Value: "    ",
			},
			wantPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(tc.in)
			if tc.wantPass && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.wantPass && err == nil {
				t.Fatalf("expected error, got nil")
			}

			if tc.assertFunc != nil {
				tc.assertFunc(t, tc.in)
			}
		})
	}
}

func TestUpsertUserReqValidation(t *testing.T) {
	v := newValidatorWithTrim(t)

	testCases := []struct {
		name     string
		in       UpsertUserReq
		wantPass bool
	}{
		{
			name: "required provider and provider_id applied",
			in: UpsertUserReq{
				Provider:   models.ProviderEnumGOOGLE,
				ProviderID: "provider-id",
			},
			wantPass: true,
		},
		{
			name: "missing provider fails required",
			in: UpsertUserReq{
				Provider:   "",
				ProviderID: "provider-id",
			},
			wantPass: false,
		},
		{
			name: "oneof rule applied",
			in: UpsertUserReq{
				Provider:   "INVALID",
				ProviderID: "provider-id",
			},
			wantPass: false,
		},
		{
			name: "trim applied to provider",
			in: UpsertUserReq{
				Provider:   "  GOOGLE  ",
				ProviderID: "provider-id",
			},
			wantPass: true,
		},
		{
			name: "trim applied to provider_id",
			in: UpsertUserReq{
				Provider:   models.ProviderEnumGOOGLE,
				ProviderID: "  id  ",
			},
			wantPass: true,
		},
		{
			name: "min rule applied to provider_id",
			in: UpsertUserReq{
				Provider:   models.ProviderEnumGOOGLE,
				ProviderID: "",
			},
			wantPass: false,
		},
		{
			name: "omitempty allows nil display_name",
			in: UpsertUserReq{
				Provider:    models.ProviderEnumGOOGLE,
				ProviderID:  "provider-id",
				DisplayName: nil,
			},
			wantPass: true,
		},
		{
			name: "trim applied to display_name",
			in: UpsertUserReq{
				Provider:    models.ProviderEnumGOOGLE,
				ProviderID:  "provider-id",
				DisplayName: strPtr("  Alice  "),
			},
			wantPass: true,
		},
		{
			name: "min rule applied to display_name",
			in: UpsertUserReq{
				Provider:    models.ProviderEnumGOOGLE,
				ProviderID:  "provider-id",
				DisplayName: strPtr(""),
			},
			wantPass: false,
		},
		{
			name: "omitempty allows nil profile_pic",
			in: UpsertUserReq{
				Provider:   models.ProviderEnumGOOGLE,
				ProviderID: "provider-id",
				ProfilePic: nil,
			},
			wantPass: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(&tc.in)
			if tc.wantPass && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.wantPass && err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestCreateLinkReqValidation(t *testing.T) {
	v := newValidatorWithTrim(t)

	testCases := []struct {
		name     string
		in       CreateLinkReq
		wantPass bool
	}{
		{
			name: "required rule applied",
			in: CreateLinkReq{
				OriginalUrl: "https://example.com",
			},
			wantPass: true,
		},
		{
			name: "empty url fails required",
			in: CreateLinkReq{
				OriginalUrl: "",
			},
			wantPass: false,
		},
		{
			name: "trim applied",
			in: CreateLinkReq{
				OriginalUrl: "  https://example.com  ",
			},
			wantPass: true,
		},
		{
			name: "url validation rule applied",
			in: CreateLinkReq{
				OriginalUrl: "not-a-url",
			},
			wantPass: false,
		},
		{
			name: "spaces only fails required after trim",
			in: CreateLinkReq{
				OriginalUrl: "    ",
			},
			wantPass: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := v.Struct(&tc.in)
			if tc.wantPass && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.wantPass && err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
