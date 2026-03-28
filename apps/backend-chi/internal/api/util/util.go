package util

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/apierror"
	"github.com/danielxfeng/short-url/apps/backend-chi/internal/api/dto"
	"github.com/go-playground/validator/v10"
)

func ParseAndValidateJSON[T any](r *http.Request, validateFn func(*T) error) (T, error) {
	var data T
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		return data, apierror.NewApiError(400, "failed to parse the json body", err)
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
