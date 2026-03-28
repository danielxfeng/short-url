package apierror_test

import (
	"errors"
	"testing"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/apierror"
)

func TestAPIError_UnwrapAsIs(t *testing.T) {
	originalErr := errors.New("original")

	testCases := []struct {
		name        string
		status      int
		message     string
		err         error
		wantIs      error
		wantMessage string
		wantErrMsg  string
	}{
		{
			name:        "wraps original error",
			status:      500,
			message:     "boom",
			err:         originalErr,
			wantIs:      originalErr,
			wantMessage: "boom: original",
		},
		{
			name:        "nil original error creates message error",
			status:      400,
			message:     "bad request",
			err:         nil,
			wantMessage: "bad request: bad request",
			wantErrMsg:  "bad request",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := apierror.NewApiError(tc.status, tc.message, tc.err)

			var apiErr *apierror.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected errors.As to find *APIError")
			}

			if apiErr.Status != tc.status {
				t.Fatalf("status mismatch: got %d want %d", apiErr.Status, tc.status)
			}
			if apiErr.Message != tc.message {
				t.Fatalf("message mismatch: got %q want %q", apiErr.Message, tc.message)
			}

			wantIs := tc.wantIs
			if wantIs == nil {
				wantIs = apiErr.Err
			}
			if !errors.Is(err, wantIs) {
				t.Fatalf("expected errors.Is to match wrapped error")
			}
			if tc.wantErrMsg != "" && apiErr.Err.Error() != tc.wantErrMsg {
				t.Fatalf("wrapped error message mismatch: got %q want %q", apiErr.Err.Error(), tc.wantErrMsg)
			}
			if err.Error() != tc.wantMessage {
				t.Fatalf("error string mismatch: got %q want %q", err.Error(), tc.wantMessage)
			}
		})
	}
}
