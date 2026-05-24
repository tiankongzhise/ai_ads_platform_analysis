package store

import (
	"sort"
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
	ID         string         `json:"id"`
	TenantID   string         `json:"tenant_id"`
	AccountID  string         `json:"account_id"`
	Platform   string         `json:"platform"`
	ReportType string         `json:"report_type"`
	DateFrom   string         `json:"date_from"`
	DateTo     string         `json:"date_to"`
	Status     string         `json:"status"`
	Reason     string         `json:"reason"`
	CreatedAt  time.Time      `json:"created_at"`
	StartedAt  time.Time      `json:"started_at,omitempty"`
	FinishedAt time.Time      `json:"finished_at,omitempty"`
	ResultMeta map[string]any `json:"result_meta,omitempty"`
	ErrorMeta  map[string]any `json:"error_meta,omitempty"`
}

type AdEntity struct {
	ID               string         `json:"id"`
	TenantID         string         `json:"tenant_id"`
	AccountID        string         `json:"account_id"`
	Platform         string         `json:"platform"`
	EntityType       string         `json:"entity_type"`
	ExternalID       string         `json:"external_id"`
	ParentExternalID string         `json:"parent_external_id,omitempty"`
	Name             string         `json:"name"`
	Status           string         `json:"status"`
	Raw              map[string]any `json:"raw"`
	SyncedAt         time.Time      `json:"synced_at"`
}

type AdEntityFilter struct {
	AccountID  string
	Platform   string
	EntityType string
	Limit      int
}

type ReportMetrics struct {
	Cost        float64 `json:"cost"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	Conversions int64   `json:"conversions"`
}

type RawReportRow struct {
	ID               string         `json:"id"`
	TenantID         string         `json:"tenant_id"`
	AccountID        string         `json:"account_id"`
	Platform         string         `json:"platform"`
	ReportType       string         `json:"report_type"`
	Granularity      string         `json:"granularity"`
	StatDate         string         `json:"stat_date"`
	StatHour         string         `json:"stat_hour,omitempty"`
	EntityType       string         `json:"entity_type"`
	ExternalEntityID string         `json:"external_entity_id"`
	Metrics          ReportMetrics  `json:"metrics"`
	Raw              map[string]any `json:"raw"`
	SyncedAt         time.Time      `json:"synced_at"`
}

type RawReportFilter struct {
	AccountID  string
	Platform   string
	ReportType string
	Limit      int
}

type MemoryStore struct {
	mu       sync.RWMutex
	states   map[string]OAuthState
	tokens   map[string]Token
	accounts map[string]Account
	jobs     []SyncJob
	entities map[string]AdEntity
	reports  map[string]RawReportRow
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		states:   map[string]OAuthState{},
		tokens:   map[string]Token{},
		accounts: map[string]Account{},
		jobs:     []SyncJob{},
		entities: map[string]AdEntity{},
		reports:  map[string]RawReportRow{},
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

func (s *MemoryStore) AccountByID(accountID string) (Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, account := range s.accounts {
		if account.ID == accountID {
			return account, true
		}
	}
	return Account{}, false
}

func (s *MemoryStore) SaveSyncJob(job SyncJob) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for index, existing := range s.jobs {
		if existing.ID == job.ID {
			s.jobs[index] = job
			return
		}
	}
	s.jobs = append(s.jobs, job)
}

func (s *MemoryStore) SyncJobs() []SyncJob {
	s.mu.RLock()
	defer s.mu.RUnlock()
	jobs := make([]SyncJob, len(s.jobs))
	copy(jobs, s.jobs)
	return jobs
}

func (s *MemoryStore) SaveAdEntities(entities []AdEntity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, entity := range entities {
		s.entities[entity.ID] = entity
	}
}

func (s *MemoryStore) AdEntities(filter AdEntityFilter) []AdEntity {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entities := make([]AdEntity, 0, len(s.entities))
	for _, entity := range s.entities {
		if filter.AccountID != "" && entity.AccountID != filter.AccountID {
			continue
		}
		if filter.Platform != "" && entity.Platform != filter.Platform {
			continue
		}
		if filter.EntityType != "" && entity.EntityType != filter.EntityType {
			continue
		}
		entities = append(entities, entity)
	}
	sort.Slice(entities, func(i, j int) bool {
		if entities[i].EntityType == entities[j].EntityType {
			return entities[i].ExternalID < entities[j].ExternalID
		}
		return entities[i].EntityType < entities[j].EntityType
	})
	return limitEntities(entities, filter.Limit)
}

func (s *MemoryStore) SaveRawReportRows(rows []RawReportRow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, row := range rows {
		s.reports[row.ID] = row
	}
}

func (s *MemoryStore) RawReportRows(filter RawReportFilter) []RawReportRow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rows := make([]RawReportRow, 0, len(s.reports))
	for _, row := range s.reports {
		if filter.AccountID != "" && row.AccountID != filter.AccountID {
			continue
		}
		if filter.Platform != "" && row.Platform != filter.Platform {
			continue
		}
		if filter.ReportType != "" && row.ReportType != filter.ReportType {
			continue
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].StatDate == rows[j].StatDate {
			if rows[i].ReportType == rows[j].ReportType {
				return rows[i].ExternalEntityID < rows[j].ExternalEntityID
			}
			return rows[i].ReportType < rows[j].ReportType
		}
		return rows[i].StatDate > rows[j].StatDate
	})
	return limitReports(rows, filter.Limit)
}

func limitEntities(entities []AdEntity, limit int) []AdEntity {
	if limit <= 0 || len(entities) <= limit {
		return entities
	}
	return entities[:limit]
}

func limitReports(rows []RawReportRow, limit int) []RawReportRow {
	if limit <= 0 || len(rows) <= limit {
		return rows
	}
	return rows[:limit]
}
