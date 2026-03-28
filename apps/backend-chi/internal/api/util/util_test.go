package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/apierror"
	"github.com/go-playground/validator/v10"
)

func TestParseAndValidateJSON(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	validateOK := func(p *payload) error {
		if p.Name == "" {
			return errors.New("name required")
		}
		return nil
	}

	testCases := []struct {
		name       string
		body       string
		validateFn func(*payload) error
		wantName   string
		wantErr    bool
	}{
		{
			name:       "valid json with validation",
			body:       `{"name":"alice","age":30}`,
			validateFn: validateOK,
			wantName:   "alice",
			wantErr:    false,
		},
		{
			name:       "invalid json",
			body:       `{"name":`,
			validateFn: validateOK,
			wantErr:    true,
		},
		{
			name:       "nil validator skips validation",
			body:       `{"name":"bob","age":20}`,
			validateFn: nil,
			wantName:   "bob",
			wantErr:    false,
		},
		{
			name:       "validator returns error",
			body:       `{"name":"","age":20}`,
			validateFn: validateOK,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tc.body))
			got, err := ParseAndValidateJSON[payload](req, tc.validateFn)
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tc.wantErr && got.Name != tc.wantName {
				t.Fatalf("name mismatch: got %q want %q", got.Name, tc.wantName)
			}
		})
	}
}

func TestSendError(t *testing.T) {
	type resp struct {
		Error string `json:"error"`
	}

	type user struct {
		Name string `validate:"required"`
	}

	validate := validator.New()
	validationErr := func() error {
		err := validate.Struct(user{})
		if err == nil {
			return errors.New("expected validation error")
		}
		return err
	}()

	testCases := []struct {
		name         string
		err          error
		wantStatus   int
		wantErrorMsg string
		expectPanic  bool
	}{
		{
			name:         "apierror is handled",
			err:          apierror.NewApiError(401, "unauthorized", errors.New("auth")),
			wantStatus:   401,
			wantErrorMsg: "unauthorized",
		},
		{
			name:        "validation errors are handled",
			err:         validationErr,
			wantStatus:  400,
			expectPanic: false,
		},
		{
			name:        "unknown error panics",
			err:         errors.New("boom"),
			expectPanic: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rr := httptest.NewRecorder()

			didPanic := false
			func() {
				defer func() {
					if r := recover(); r != nil {
						didPanic = true
					}
				}()
				SendError(rr, tc.err)
			}()

			if tc.expectPanic {
				if !didPanic {
					t.Fatalf("expected panic")
				}
				return
			}
			if didPanic {
				t.Fatalf("unexpected panic")
			}

			if rr.Code != tc.wantStatus {
				t.Fatalf("status mismatch: got %d want %d", rr.Code, tc.wantStatus)
			}

			if tc.wantErrorMsg == "" {
				return
			}

			var got resp
			if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if got.Error != tc.wantErrorMsg {
				t.Fatalf("error message mismatch: got %q want %q", got.Error, tc.wantErrorMsg)
			}
		})
	}
}
