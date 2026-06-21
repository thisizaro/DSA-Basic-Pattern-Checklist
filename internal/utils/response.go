package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// APIError is the standard JSON error shape returned by every handler.
type APIError struct {
	Error string `json:"error"`
}

// WriteJSON encodes v as JSON and writes it with the given status code.
// Centralizing this avoids repeating header-setting + encoding boilerplate
// in every handler.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON encode error: %v", err)
	}
}

// WriteError writes a JSON-formatted error response.
func WriteError(w http.ResponseWriter, status int, message string) {
	log.Printf("ERROR: status=%d msg=%s", status, message)
	WriteJSON(w, status, APIError{Error: message})
}

// DecodeJSON decodes a request body into v, returning a descriptive error
// on malformed JSON so handlers can surface a clean 400 response.
func DecodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}
