package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"eduadcrm/services/control-plane/internal/config"
	"eduadcrm/services/control-plane/internal/store"
)

type Service struct {
	cfg   *config.Manager
	store *store.MemoryStore
}

type RegisterRequest struct {
	TenantName  string `json:"tenant_name"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type TokenResponse struct {
	AccessToken           string       `json:"access_token"`
	RefreshToken          string       `json:"refresh_token"`
	TokenType             string       `json:"token_type"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  store.User   `json:"user"`
	Tenant                store.Tenant `json:"tenant"`
}

func NewService(cfg *config.Manager, memory *store.MemoryStore) *Service {
	return &Service{cfg: cfg, store: memory}
}

func (s *Service) Register(req RegisterRequest) (TokenResponse, error) {
	if strings.TrimSpace(req.TenantName) == "" || strings.TrimSpace(req.Email) == "" || len(req.Password) < 8 {
		return TokenResponse{}, errors.New("tenant_name, email and password length >= 8 are required")
	}
	authCfg, err := s.cfg.Auth()
	if err != nil {
		return TokenResponse{}, err
	}
	tenant := store.Tenant{
		ID:        newID(),
		Name:      req.TenantName,
		Type:      "formal",
		Status:    "active",
		CreatedAt: time.Now().UTC(),
	}
	user := store.User{
		ID:          newID(),
		TenantID:    tenant.ID,
		Email:       strings.ToLower(strings.TrimSpace(req.Email)),
		Phone:       req.Phone,
		DisplayName: req.DisplayName,
		Status:      "active",
		CreatedAt:   time.Now().UTC(),
	}
	if user.DisplayName == "" {
		user.DisplayName = user.Email
	}
	passwordHash, err := HashPassword(req.Password, authCfg.PasswordIterations)
	if err != nil {
		return TokenResponse{}, err
	}
	if err := s.store.CreateTenantAndUser(tenant, user, passwordHash); err != nil {
		return TokenResponse{}, err
	}
	return s.issueTokens(user, tenant)
}

func (s *Service) Login(req LoginRequest) (TokenResponse, error) {
	user, passwordHash, ok := s.store.UserByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if !ok || !VerifyPassword(passwordHash, req.Password) {
		return TokenResponse{}, errors.New("invalid email or password")
	}
	tenant := store.Tenant{ID: user.TenantID, Name: "当前租户", Type: "formal", Status: "active"}
	return s.issueTokens(user, tenant)
}

func (s *Service) Refresh(refreshToken string) (TokenResponse, error) {
	tokenHash := HashToken(refreshToken)
	saved, ok := s.store.RefreshToken(tokenHash)
	if !ok || saved.RevokedAt != nil || time.Now().UTC().After(saved.ExpiresAt) {
		return TokenResponse{}, errors.New("invalid refresh token")
	}
	user, ok := s.store.UserByID(saved.UserID)
	if !ok {
		return TokenResponse{}, errors.New("user not found")
	}
	authCfg, err := s.cfg.Auth()
	if err != nil {
		return TokenResponse{}, err
	}
	if authCfg.RefreshTokenRotation {
		s.store.RevokeRefreshToken(tokenHash)
	}
	tenant := store.Tenant{ID: user.TenantID, Name: "当前租户", Type: "formal", Status: "active"}
	return s.issueTokens(user, tenant)
}

func (s *Service) Logout(accessToken string, refreshToken string) error {
	authCfg, err := s.cfg.Auth()
	if err != nil {
		return err
	}
	if refreshToken != "" {
		s.store.RevokeRefreshToken(HashToken(refreshToken))
	}
	if authCfg.LogoutBlacklist && accessToken != "" {
		claims, err := Verify(accessToken, authCfg.JWTSecret)
		if err == nil {
			s.store.BlacklistAccessToken(HashToken(accessToken), time.Unix(claims.Expires, 0))
		}
	}
	return nil
}

func (s *Service) VerifyAccessToken(accessToken string) (Claims, error) {
	authCfg, err := s.cfg.Auth()
	if err != nil {
		return Claims{}, err
	}
	if s.store.IsBlacklisted(HashToken(accessToken)) {
		return Claims{}, errors.New("token blacklisted")
	}
	return Verify(accessToken, authCfg.JWTSecret)
}

func (s *Service) issueTokens(user store.User, tenant store.Tenant) (TokenResponse, error) {
	authCfg, err := s.cfg.Auth()
	if err != nil {
		return TokenResponse{}, err
	}
	now := time.Now().UTC()
	accessExpires := now.Add(authCfg.AccessTokenTTL)
	refreshExpires := now.Add(authCfg.RefreshTokenTTL)
	claims := Claims{
		Subject:  user.ID,
		TenantID: user.TenantID,
		Email:    user.Email,
		Roles:    []string{"tenant_admin"},
		Issuer:   authCfg.Issuer,
		IssuedAt: now.Unix(),
		Expires:  accessExpires.Unix(),
		TokenID:  newID(),
	}
	accessToken, err := Sign(claims, authCfg.JWTSecret)
	if err != nil {
		return TokenResponse{}, err
	}
	refreshToken, err := RandomToken()
	if err != nil {
		return TokenResponse{}, err
	}
	s.store.SaveRefreshToken(store.RefreshToken{
		ID:        newID(),
		TenantID:  user.TenantID,
		UserID:    user.ID,
		TokenHash: HashToken(refreshToken),
		ExpiresAt: refreshExpires,
	})
	return TokenResponse{
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		TokenType:             "Bearer",
		AccessTokenExpiresAt:  accessExpires,
		RefreshTokenExpiresAt: refreshExpires,
		User:                  user,
		Tenant:                tenant,
	}, nil
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(bytes[:])
}
