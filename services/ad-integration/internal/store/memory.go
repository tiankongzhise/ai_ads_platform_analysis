package store

import (
	"sync"
	"time"
)

type OAuthState struct {
	State                string     `json:"state"`
	TenantID             string     `json:"tenant_id"`
	UserID               string     `json:"user_id"`
	OrganizationID       string     `json:"organization_id"`
	ChannelID            string     `json:"channel_id"`
	Platform             string     `json:"platform"`
	RedirectAfterSuccess string     `json:"redirect_after_success"`
	ExpiresAt            time.Time  `json:"expires_at"`
	UsedAt               *time.Time `json:"used_at,omitempty"`
}

type Token struct {
	ID                     string    `json:"id"`
	TenantID               string    `json:"tenant_id"`
	Platform               string    `json:"platform"`
	ExternalAccountID      string    `json:"external_account_id"`
	TokenRef               string    `json:"token_ref"`
	AccessTokenCiphertext  string    `json:"access_token_ciphertext"`
	RefreshTokenCiphertext string    `json:"refresh_token_ciphertext,omitempty"`
	AccessTokenExpiresAt   time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt  time.Time `json:"refresh_token_expires_at"`
	CreatedAt              time.Time `json:"created_at"`
}

type Account struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	OrganizationID    string    `json:"organization_id"`
	ChannelID         string    `json:"channel_id"`
	Platform          string    `json:"platform"`
	ExternalAccountID string    `json:"external_account_id"`
	AccountName       string    `json:"account_name"`
	Status            string    `json:"status"`
	TokenRef          string    `json:"token_ref"`
	LastSyncStatus    string    `json:"last_sync_status"`
	CreatedAt         time.Time `json:"created_at"`
}

type SyncJob struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	AccountID  string    `json:"account_id"`
	Platform   string    `json:"platform"`
	ReportType string    `json:"report_type"`
	DateFrom   string    `json:"date_from"`
	DateTo     string    `json:"date_to"`
	Status     string    `json:"status"`
	Reason     string    `json:"reason"`
	CreatedAt  time.Time `json:"created_at"`
}

type MemoryStore struct {
	mu       sync.RWMutex
	states   map[string]OAuthState
	tokens   map[string]Token
	accounts map[string]Account
	jobs     []SyncJob
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		states:   map[string]OAuthState{},
		tokens:   map[string]Token{},
		accounts: map[string]Account{},
		jobs:     []SyncJob{},
	}
}

func (s *MemoryStore) SaveState(state OAuthState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state.State] = state
}

func (s *MemoryStore) ConsumeState(stateValue string) (OAuthState, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	state, ok := s.states[stateValue]
	if !ok || state.UsedAt != nil || time.Now().UTC().After(state.ExpiresAt) {
		return OAuthState{}, false
	}
	now := time.Now().UTC()
	state.UsedAt = &now
	s.states[stateValue] = state
	return state, true
}

func (s *MemoryStore) SaveToken(token Token) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token.TokenRef] = token
}

func (s *MemoryStore) SaveAccount(account Account) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := account.TenantID + ":" + account.Platform + ":" + account.ExternalAccountID
	s.accounts[key] = account
}

func (s *MemoryStore) Accounts() []Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	accounts := make([]Account, 0, len(s.accounts))
	for _, account := range s.accounts {
		accounts = append(accounts, account)
	}
	return accounts
}

func (s *MemoryStore) SaveSyncJob(job SyncJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.jobs = append(s.jobs, job)
}

func (s *MemoryStore) SyncJobs() []SyncJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	jobs := make([]SyncJob, len(s.jobs))
	copy(jobs, s.jobs)
	return jobs
}
