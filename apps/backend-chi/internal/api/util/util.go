package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/apierror"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/dto"
	"github.com/go-playground/validator/v10"
)

func ParseAndValidateJSON[T any](r *http.Request, validateFn func(*T) error) (T, error) {
	var data T
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		return data, apierror.NewApiError(400, "failed to parse the json body", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return data, apierror.NewApiError(400, "failed to parse the json body", errors.New("request body must contain a single JSON object"))
	}

	if validateFn == nil {
		return data, nil
	}

	if err := validateFn(&data); err != nil {
		return data, err
	}

	return data, nil
}

func SendJSON[T any](w http.ResponseWriter, status int, data T) {
	payload, err := json.Marshal(data)
	if err != nil {
		SendError(w, apierror.NewApiError(500, "failed to serialize the response", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(payload)
}

func SendError(w http.ResponseWriter, err error) {
	var apiErr *apierror.APIError

	// Handle APIError
	if errors.As(err, &apiErr) {
		SendJSON(w, apiErr.Status, dto.APIErrorRes{Error: apiErr.Message})
		return
	}

	var validationErr validator.ValidationErrors

	// Handle validation errors
	if errors.As(err, &validationErr) {
		messages := make([]string, 0, len(validationErr))
		for _, fe := range validationErr {
			messages = append(messages, fe.Error())
		}
		SendJSON(w, 400, dto.APIErrorRes{Error: strings.Join(messages, "; ")})
		return
	}

	// It would be a 500 so recovery handles the rest.
	panic(err)
}

func ParseInt32ClampedOrDefault(value string, defaultValue int32, min int32, max int32) int32 {

	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	v32 := int32(v)

	if v32 > max {
		return max
	}

	if v32 < min {
		return min
	}

	return v32
}

// https://github.com/go-chi/chi/blob/master/middleware/request_id.go
func GenerateRandomString(length int) string {
	var buf [12]byte
	var b64 string
	for len(b64) < length {
		_, err := rand.Read(buf[:])
		if err != nil {
			panic("failed to generate random string: " + err.Error())
		}
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}
	return b64[:length]
}

func AssembleURL(baseURL, key string, value string) string {
	u, _ := url.Parse(baseURL)

	q := u.Query()
	q.Set(key, value)
	u.RawQuery = q.Encode()
	return u.String()
}
