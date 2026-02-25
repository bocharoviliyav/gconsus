package rest

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

const (
	serverErrorMsg   = "server error"
	notFoundErrorMsg = "not found"
)

type Error struct {
	Message string
}

type ErrorResponse struct {
	Errors []Error
}

func ReturnResponse(w http.ResponseWriter, v any) {
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("ReturnResponse failed to json.Encode", "err", err)
	}
}

func ReturnCreateResponse(w http.ResponseWriter, v any) {
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("ReturnResponse failed to json.Encode", "err", err)
	}
}

func ReturnServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)

	resp := ErrorResponse{
		Errors: []Error{{Message: serverErrorMsg}},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("ReturnServerError failed to json.Encode", "err", err)
	}
}

func ReturnRequestError(w http.ResponseWriter, errorMessage string) {
	w.WriteHeader(http.StatusBadRequest)

	resp := ErrorResponse{
		Errors: []Error{{Message: errorMessage}},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("ReturnRequestError failed to json.Encode", "err", err)
	}
}

func ReturnNotFoundError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)

	resp := ErrorResponse{
		Errors: []Error{{Message: notFoundErrorMsg}},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("ReturnNotFoundError failed to json.Encode", "err", err)
	}
}

func ReturnNotFound(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusNotFound)

	resp := ErrorResponse{
		Errors: []Error{{Message: message}},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("ReturnNotFound failed to json.Encode", "err", err)
	}
}

func ReturnConflict(w http.ResponseWriter, message string) {
	w.WriteHeader(http.StatusConflict)

	resp := ErrorResponse{
		Errors: []Error{{Message: message}},
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("ReturnConflict failed to json.Encode", "err", err)
	}
}

func ReturnNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func ReturnCreated(w http.ResponseWriter, v any) {
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("ReturnCreated failed to json.Encode", "err", err)
	}
}
