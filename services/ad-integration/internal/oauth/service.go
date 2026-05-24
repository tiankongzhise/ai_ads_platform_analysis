package oauth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"eduadcrm/services/ad-integration/internal/store"
)

type Service struct {
	store store.Repository
	queue store.MessageQueue
}

type AuthorizeRequest struct {
	TenantID             string `json:"tenant_id"`
	UserID               string `json:"user_id"`
	OrganizationID       string `json:"organization_id"`
	ChannelID            string `json:"channel_id"`
	RedirectAfterSuccess string `json:"redirect_after_success"`
}

type AuthorizeResponse struct {
	Platform  string    `json:"platform"`
	State     string    `json:"state"`
	AuthURL   string    `json:"auth_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CallbackResult struct {
	Account store.Account `json:"account"`
	Job     store.SyncJob `json:"sync_job"`
}

func NewService(repository store.Repository) *Service {
	return &Service{store: repository}
}

func NewServiceWithQueue(repository store.Repository, queue store.MessageQueue) *Service {
	return &Service{store: repository, queue: queue}
}

func (s *Service) Authorize(platform string, req AuthorizeRequest) (AuthorizeResponse, error) {
	platform = strings.ToLower(platform)
	if !SupportedPlatform(platform) {
		return AuthorizeResponse{}, fmt.Errorf("unsupported platform %s", platform)
	}
	if req.TenantID == "" || req.UserID == "" || req.OrganizationID == "" || req.ChannelID == "" {
		return AuthorizeResponse{}, errors.New("tenant_id, user_id, organization_id and channel_id are required")
	}
	stateValue := randomURLToken()
	expiresAt := time.Now().UTC().Add(10 * time.Minute)
	s.store.SaveState(store.OAuthState{
		State:                stateValue,
		TenantID:             req.TenantID,
		UserID:               req.UserID,
		OrganizationID:       req.OrganizationID,
		ChannelID:            req.ChannelID,
		Platform:             platform,
		RedirectAfterSuccess: req.RedirectAfterSuccess,
		ExpiresAt:            expiresAt,
	})
	return AuthorizeResponse{
		Platform:  platform,
		State:     stateValue,
		AuthURL:   buildMockAuthURL(platform, stateValue),
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Service) Callback(platform string, query url.Values) (CallbackResult, error) {
	platform = strings.ToLower(platform)
	stateValue := query.Get("state")
	if stateValue == "" {
		return CallbackResult{}, errors.New("state is required")
	}
	state, ok := s.store.ConsumeState(stateValue)
	if !ok || state.Platform != platform {
		return CallbackResult{}, errors.New("invalid, expired or replayed state")
	}
	code := firstNonEmpty(query.Get("code"), query.Get("auth_code"), query.Get("authorization_code"))
	if code == "" {
		return CallbackResult{}, errors.New("authorization code is required")
	}
	externalAccountID := firstNonEmpty(query.Get("account_id"), query.Get("advertiser_id"), "mock_"+platform+"_account")
	accountName := firstNonEmpty(query.Get("account_name"), platform+" 授权账户")
	tokenRef := fmt.Sprintf("kms://ad-platform-token/%s/%s/%s", platform, state.TenantID, externalAccountID)
	token := store.Token{
		ID:                     newID(),
		TenantID:               state.TenantID,
		Platform:               platform,
		ExternalAccountID:      externalAccountID,
		TokenRef:               tokenRef,
		AccessTokenCiphertext:  "ciphertext:" + code,
		RefreshTokenCiphertext: "ciphertext:refresh:" + code,
		AccessTokenExpiresAt:   time.Now().UTC().Add(24 * time.Hour),
		RefreshTokenExpiresAt:  time.Now().UTC().Add(30 * 24 * time.Hour),
		CreatedAt:              time.Now().UTC(),
	}
	s.store.SaveToken(token)
	account := store.Account{
		ID:                newID(),
		TenantID:          state.TenantID,
		OrganizationID:    state.OrganizationID,
		ChannelID:         state.ChannelID,
		Platform:          platform,
		ExternalAccountID: externalAccountID,
		AccountName:       accountName,
		Status:            "active",
		TokenRef:          tokenRef,
		LastSyncStatus:    "pending",
		CreatedAt:         time.Now().UTC(),
	}
	s.store.SaveAccount(account)
	job := store.SyncJob{
		ID:         newID(),
		TenantID:   state.TenantID,
		AccountID:  account.ID,
		Platform:   platform,
		ReportType: "account_daily",
		DateFrom:   time.Now().UTC().AddDate(0, 0, -7).Format("2006-01-02"),
		DateTo:     time.Now().UTC().Format("2006-01-02"),
		Status:     "pending",
		Reason:     "oauth_callback",
		CreatedAt:  time.Now().UTC(),
	}
	s.store.SaveSyncJob(job)
	if s.queue != nil {
		_ = s.queue.Publish("ad_sync_queue", store.QueueMessage{
			MessageID:     job.ID,
			SchemaVersion: "1.0",
			EventType:     "ad.sync.requested",
			TenantID:      job.TenantID,
			TraceID:       "oauth_callback",
			OccurredAt:    job.CreatedAt,
			Producer:      "ad-integration-service",
			Payload: map[string]any{
				"account_id":   job.AccountID,
				"platform":     job.Platform,
				"report_types": []string{job.ReportType},
				"date_from":    job.DateFrom,
				"date_to":      job.DateTo,
				"reason":       job.Reason,
			},
		})
	}
	return CallbackResult{Account: account, Job: job}, nil
}

func SupportedPlatform(platform string) bool {
	switch strings.ToLower(platform) {
	case "douyin", "tencent", "baidu", "xiaohongshu":
		return true
	default:
		return false
	}
}

func buildMockAuthURL(platform string, state string) string {
	base := map[string]string{
		"douyin":      "https://ad.oceanengine.com/openapi/audit/oauth.html",
		"tencent":     "https://developers.e.qq.com/oauth/authorize",
		"baidu":       "https://dev2.baidu.com/oauth/authorize",
		"xiaohongshu": "https://ad-market.xiaohongshu.com/oauth/authorize",
	}[platform]
	values := url.Values{}
	values.Set("state", state)
	values.Set("response_type", "code")
	values.Set("redirect_uri", "http://localhost:8081/api/oauth/"+platform+"/callback")
	return base + "?" + values.Encode()
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func randomURLToken() string {
	var bytes [32]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return base64.RawURLEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return base64.RawURLEncoding.EncodeToString(bytes[:])
}

func newID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405.000000000")))
	}
	return hex.EncodeToString(bytes[:])
}
