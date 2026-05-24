package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"eduadcrm/services/ad-integration/internal/oauth"
	"eduadcrm/services/ad-integration/internal/store"
)

type Server struct {
	store *store.MemoryStore
	oauth *oauth.Service
}

type envelope struct {
	Success   bool       `json:"success"`
	Data      any        `json:"data,omitempty"`
	Error     *errorBody `json:"error,omitempty"`
	RequestID string     `json:"request_id"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewServer() *Server {
	memory := store.NewMemoryStore()
	return &Server{store: memory, oauth: oauth.NewService(memory)}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("POST /api/oauth/{platform}/authorize", s.authorize)
	mux.HandleFunc("GET /api/oauth/{platform}/callback", s.callback)
	mux.HandleFunc("GET /api/ad-accounts", s.accounts)
	mux.HandleFunc("GET /api/ad-sync/jobs", s.syncJobs)
	return withCORS(mux)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, map[string]string{"status": "ok", "service": "ad-integration"})
}

func (s *Server) authorize(w http.ResponseWriter, r *http.Request) {
	platform := strings.ToLower(r.PathValue("platform"))
	var req oauth.AuthorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "INVALID_JSON", "请求体不是有效 JSON")
		return
	}
	resp, err := s.oauth.Authorize(platform, req)
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "OAUTH_AUTHORIZE_FAILED", err.Error())
		return
	}
	writeJSON(w, r, http.StatusOK, resp)
}

func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	platform := strings.ToLower(r.PathValue("platform"))
	result, err := s.oauth.Callback(platform, r.URL.Query())
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "OAUTH_CALLBACK_FAILED", err.Error())
		return
	}
	writeJSON(w, r, http.StatusOK, result)
}

func (s *Server) accounts(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, s.store.Accounts())
}

func (s *Server) syncJobs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, r, http.StatusOK, s.store.SyncJobs())
}

func writeJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Success: true, Data: data, RequestID: requestID(r)})
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(envelope{Success: false, Error: &errorBody{Code: code, Message: message}, RequestID: requestID(r)})
}

func requestID(r *http.Request) string {
	if requestID := r.Header.Get("X-Request-Id"); requestID != "" {
		return requestID
	}
	return "req_local"
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
