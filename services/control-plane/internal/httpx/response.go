package httpx

import (
	"encoding/json"
	"net/http"
)

type Envelope struct {
	Success   bool       `json:"success"`
	Data      any        `json:"data,omitempty"`
	Error     *ErrorBody `json:"error,omitempty"`
	RequestID string     `json:"request_id"`
}

type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success:   true,
		Data:      data,
		RequestID: RequestID(r),
	})
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code string, message string, details map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{
		Success: false,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: RequestID(r),
	})
}

func RequestID(r *http.Request) string {
	if requestID := r.Header.Get("X-Request-Id"); requestID != "" {
		return requestID
	}
	return "req_local"
}
