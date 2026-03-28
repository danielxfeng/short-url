package dep_test

import (
	"testing"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/dep"
)

const testKey = "TEST_KEY"
const notSet = "notSet"

func setEnv(t *testing.T, v string) {
	t.Helper()

	if v == notSet {
		return
	}

	t.Setenv(testKey, v)
}

func TestGetEnvStrOrDefault(t *testing.T) {
	const validValue = "v1"
	const defaultValue = "v"
	const emptyValue = ""

	testCases := []struct {
		name     string
		envValue string
		expected string
	}{
		{
			name:     "valid env string",
			envValue: validValue,
			expected: validValue,
		},
		{
			name:     "empty env string",
			envValue: emptyValue,
			expected: defaultValue,
		},
		{
			name:     "env not set",
			envValue: notSet,
			expected: defaultValue,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, tc.envValue)

			if got := dep.GetEnvStrOrDefault(testKey, defaultValue); got != tc.expected {
				t.Fatalf("expect: %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestGetEnvIntOrDefault(t *testing.T) {
	const validValue = "10"
	const validExpected = 10
	const defaultValue = 22
	const invalidValue = "a"

	testCases := []struct {
		name     string
		envValue string
		expected int
	}{
		{
			name:     "valid env (int)",
			envValue: validValue,
			expected: validExpected,
		},
		{
			name:     "env not set (int)",
			envValue: notSet,
			expected: defaultValue,
		},
		{
			name:     "invalid env (int)",
			envValue: invalidValue,
			expected: defaultValue,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, tc.envValue)

			if got := dep.GetEnvIntOrDefault(testKey, defaultValue); got != tc.expected {
				t.Fatalf("expected: %d, got: %d", tc.expected, got)
			}
		})
	}
}

func TestGetEnvStrOrError(t *testing.T) {
	const validValue = "v1"
	const emptyValue = ""
	const errorValue = "error"

	testCases := []struct {
		name      string
		envValue  string
		expected  string
		expectErr bool
	}{
		{
			name:      "valid env string",
			envValue:  validValue,
			expected:  validValue,
			expectErr: false,
		},
		{
			name:      "empty env string",
			envValue:  emptyValue,
			expected:  errorValue,
			expectErr: true,
		},
		{
			name:      "env not set",
			envValue:  notSet,
			expected:  errorValue,
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setEnv(t, tc.envValue)

			got, err := dep.GetEnvStrOrError(testKey)

			if tc.expectErr && err == nil {
				t.Fatalf("expected error, got %q", got)
			}

			if !tc.expectErr && err != nil {
				t.Fatalf("expected %q, got error %v", tc.expected, err)
			}

			if !tc.expectErr && got != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
