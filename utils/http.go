package utils

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func ParseUintParam(r *http.Request, key string) (uint, error) {
	raw := chi.URLParam(r, key)
	if raw == "" {
		return 0, errors.New("missing param")
	}
	u64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || u64 == 0 {
		return 0, errors.New("invalid param")
	}
	return uint(u64), nil
}
