package app

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"eduadcrm/services/control-plane/internal/auth"
	"eduadcrm/services/control-plane/internal/config"
	"eduadcrm/services/control-plane/internal/httpx"
	"eduadcrm/services/control-plane/internal/store"
)

type Server struct {
	config *config.Manager
	auth   *auth.Service
	store  store.Repository
}

func NewServer(manager *config.Manager) *Server {
	return NewServerWithStore(manager, store.NewMemoryStore())
}

func NewServerWithStore(manager *config.Manager, repository store.Repository) *Server {
	return &Server{
		config: manager,
		auth:   auth.NewService(manager, repository),
		store:  repository,
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/config", s.getConfig)
	mux.HandleFunc("PATCH /api/config/{key}", s.updateConfig)
	mux.HandleFunc("POST /api/config/{key}/rollback", s.rollbackConfig)
	mux.HandleFunc("GET /api/config/change-logs", s.configLogs)
	mux.HandleFunc("POST /api/auth/register", s.register)
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/auth/refresh", s.refresh)
	mux.HandleFunc("POST /api/auth/logout", s.logout)
	mux.HandleFunc("GET /api/auth/profile", s.profile)
	mux.HandleFunc("GET /api/orgs/tree", s.listOrganizations)
	mux.HandleFunc("POST /api/orgs", s.createOrganization)
	mux.HandleFunc("PATCH /api/orgs/{org_id}", s.updateOrganization)
	mux.HandleFunc("POST /api/orgs/{org_id}/move", s.moveOrganization)
	mux.HandleFunc("GET /api/orgs/{org_id}/summary", s.organizationSummary)
	mux.HandleFunc("GET /api/teams", s.listTeams)
	mux.HandleFunc("POST /api/teams", s.createTeam)
	mux.HandleFunc("PATCH /api/teams/{team_id}", s.updateTeam)
	mux.HandleFunc("GET /api/channels", s.listChannels)
	mux.HandleFunc("POST /api/channels", s.createChannel)
	mux.HandleFunc("PATCH /api/channels/{channel_id}", s.updateChannel)
	return withCORS(mux)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, map[string]string{"status": "ok", "service": "control-plane"})
}

func (s *Server) getConfig(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, s.config.Public())
}

func (s *Server) updateConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_JSON", "请求体不是有效 JSON", nil)
		return
	}
	changedBy := "system"
	if claims, ok := s.claimsFromRequest(r); ok {
		changedBy = claims.Subject
	}
	value, err := s.config.Update(key, body.Value, changedBy)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "CONFIG_INVALID", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, value)
}

func (s *Server) configLogs(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, s.config.Logs())
}

func (s *Server) rollbackConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	changedBy := "system"
	if claims, ok := s.claimsFromRequest(r); ok {
		changedBy = claims.Subject
	}
	value, err := s.config.Rollback(key, changedBy)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "CONFIG_ROLLBACK_FAILED", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, value)
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_JSON", "请求体不是有效 JSON", nil)
		return
	}
	resp, err := s.auth.Register(req)
	if err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "REGISTER_FAILED", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusCreated, resp)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_JSON", "请求体不是有效 JSON", nil)
		return
	}
	resp, err := s.auth.Login(req)
	if err != nil {
		httpx.WriteError(w, r, http.StatusUnauthorized, "LOGIN_FAILED", "邮箱或密码错误", nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (s *Server) refresh(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_REFRESH_TOKEN", "refresh_token 不能为空", nil)
		return
	}
	resp, err := s.auth.Refresh(body.RefreshToken)
	if err != nil {
		httpx.WriteError(w, r, http.StatusUnauthorized, "REFRESH_FAILED", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, resp)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	accessToken := bearerToken(r)
	if err := s.auth.Logout(accessToken, body.RefreshToken); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "LOGOUT_FAILED", err.Error(), nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]bool{"logged_out": true})
}

func (s *Server) profile(w http.ResponseWriter, r *http.Request) {
	claims, ok := s.claimsFromRequest(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	user, exists := s.store.UserByID(claims.Subject)
	if !exists {
		httpx.WriteError(w, r, http.StatusUnauthorized, "USER_NOT_FOUND", "用户不存在", nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, map[string]any{"user": user, "claims": claims})
}

func (s *Server) listOrganizations(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, s.store.OrganizationTree(tenantID))
}

func (s *Server) createOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	var body struct {
		ParentID string `json:"parent_id"`
		Name     string `json:"name"`
		OrgType  string `json:"org_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_ORG", "组织名称不能为空", nil)
		return
	}
	if body.OrgType == "" {
		body.OrgType = "school"
	}
	org := store.Organization{
		ID:        newLocalID(),
		TenantID:  tenantID,
		ParentID:  body.ParentID,
		Name:      body.Name,
		OrgType:   body.OrgType,
		Status:    "active",
		CreatedAt: nowUTC(),
		UpdatedAt: nowUTC(),
	}
	s.store.SaveOrganization(org)
	httpx.WriteJSON(w, r, http.StatusCreated, org)
}

func (s *Server) updateOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	var body struct {
		Name    string `json:"name"`
		OrgType string `json:"org_type"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_JSON", "请求体不是有效 JSON", nil)
		return
	}
	org, err := s.store.UpdateOrganization(tenantID, r.PathValue("org_id"), body.Name, body.OrgType, body.Status, nowUTC())
	if err != nil {
		httpx.WriteError(w, r, http.StatusNotFound, "ORG_NOT_FOUND", "组织不存在", nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, org)
}

func (s *Server) moveOrganization(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	var body struct {
		ParentID string `json:"parent_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_JSON", "请求体不是有效 JSON", nil)
		return
	}
	org, err := s.store.MoveOrganization(tenantID, r.PathValue("org_id"), body.ParentID, nowUTC())
	if err != nil {
		status := http.StatusBadRequest
		code := "ORG_MOVE_INVALID"
		message := err.Error()
		if err == store.ErrNotFound {
			status = http.StatusNotFound
			code = "ORG_NOT_FOUND"
			message = "组织不存在"
		}
		httpx.WriteError(w, r, status, code, message, nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, org)
}

func (s *Server) organizationSummary(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	summary, err := s.store.OrganizationSummary(tenantID, r.PathValue("org_id"))
	if err != nil {
		httpx.WriteError(w, r, http.StatusNotFound, "ORG_NOT_FOUND", "组织不存在", nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, summary)
}

func (s *Server) listTeams(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, s.store.Teams(tenantID))
}

func (s *Server) createTeam(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	var body struct {
		OrganizationID string `json:"organization_id"`
		Name           string `json:"name"`
		LeaderUserID   string `json:"leader_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.OrganizationID == "" || body.Name == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_TEAM", "团队名称和组织 ID 不能为空", nil)
		return
	}
	team := store.Team{
		ID:             newLocalID(),
		TenantID:       tenantID,
		OrganizationID: body.OrganizationID,
		Name:           body.Name,
		LeaderUserID:   body.LeaderUserID,
		Status:         "active",
		CreatedAt:      nowUTC(),
		UpdatedAt:      nowUTC(),
	}
	s.store.SaveTeam(team)
	httpx.WriteJSON(w, r, http.StatusCreated, team)
}

func (s *Server) updateTeam(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	var body struct {
		Name         string `json:"name"`
		LeaderUserID string `json:"leader_user_id"`
		Status       string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_JSON", "请求体不是有效 JSON", nil)
		return
	}
	team, err := s.store.UpdateTeam(tenantID, r.PathValue("team_id"), body.Name, body.LeaderUserID, body.Status, nowUTC())
	if err != nil {
		httpx.WriteError(w, r, http.StatusNotFound, "TEAM_NOT_FOUND", "团队不存在", nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, team)
}

func (s *Server) listChannels(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, s.store.Channels(tenantID))
}

func (s *Server) createChannel(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	var body struct {
		OrganizationID string `json:"organization_id"`
		Platform       string `json:"platform"`
		DisplayName    string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.OrganizationID == "" || body.Platform == "" {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_CHANNEL", "组织 ID 和平台不能为空", nil)
		return
	}
	if body.DisplayName == "" {
		body.DisplayName = body.Platform
	}
	channel := store.Channel{
		ID:             newLocalID(),
		TenantID:       tenantID,
		OrganizationID: body.OrganizationID,
		Platform:       body.Platform,
		DisplayName:    body.DisplayName,
		Status:         "active",
		CreatedAt:      nowUTC(),
		UpdatedAt:      nowUTC(),
	}
	s.store.SaveChannel(channel)
	httpx.WriteJSON(w, r, http.StatusCreated, channel)
}

func (s *Server) updateChannel(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantID(r)
	if !ok {
		httpx.WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "请先登录", nil)
		return
	}
	var body struct {
		DisplayName string `json:"display_name"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httpx.WriteError(w, r, http.StatusBadRequest, "INVALID_JSON", "请求体不是有效 JSON", nil)
		return
	}
	channel, err := s.store.UpdateChannel(tenantID, r.PathValue("channel_id"), body.DisplayName, body.Status, nowUTC())
	if err != nil {
		httpx.WriteError(w, r, http.StatusNotFound, "CHANNEL_NOT_FOUND", "渠道不存在", nil)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, channel)
}

func (s *Server) claimsFromRequest(r *http.Request) (auth.Claims, bool) {
	token := bearerToken(r)
	if token == "" {
		return auth.Claims{}, false
	}
	claims, err := s.auth.VerifyAccessToken(token)
	return claims, err == nil
}

func (s *Server) tenantID(r *http.Request) (string, bool) {
	claims, ok := s.claimsFromRequest(r)
	if !ok {
		return "", false
	}
	return claims.TenantID, true
}

func newLocalID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return strings.ReplaceAll(nowUTC().Format("20060102150405.000000000"), ".", "")
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(bytes[:])
	return encoded[0:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:32]
}

func nowUTC() time.Time {
	return time.Now().UTC()
}

func bearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
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
