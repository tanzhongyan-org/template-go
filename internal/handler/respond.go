package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Message string `json:"message"`
}

// writeJSON encodes v before writing so an encoding failure becomes a 500
// error envelope instead of a truncated 2xx response.
func writeJSON(w http.ResponseWriter, status int, v any) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(buf.Bytes()); err != nil {
		log.Printf("write response: %v", err)
	}
}

// writeError sends the one error envelope every endpoint uses:
// {"error":{"message":"..."}} with Content-Type application/json.
func writeError(w http.ResponseWriter, status int, message string) {
	body, err := json.Marshal(errorBody{Error: errorDetail{Message: message}})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(append(body, '\n')); err != nil {
		log.Printf("write error response: %v", err)
	}
}

// decodeJSON reads the request body into dst. It answers 413 when the body
// exceeds the LimitBody cap and 400 for any other malformed payload, and
// reports whether the caller may continue.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "Request body too large")
			return false
		}
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return false
	}
	return true
}
