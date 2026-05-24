package store

import (
	"errors"
	"sync"
	"time"
)

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"tenant_type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone,omitempty"`
	DisplayName string    `json:"display_name"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

type RefreshToken struct {
	ID               string
	TenantID         string
	UserID           string
	TokenHash        string
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	RotationParentID string
}

type MemoryStore struct {
	mu              sync.RWMutex
	tenants         map[string]Tenant
	users           map[string]User
	userIDsByEmail  map[string]string
	passwordsByUser map[string]string
	refreshTokens   map[string]RefreshToken
	blacklist       map[string]time.Time
	orgs            map[string]Organization
	teams           map[string]Team
	channels        map[string]Channel
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tenants:         map[string]Tenant{},
		users:           map[string]User{},
		userIDsByEmail:  map[string]string{},
		passwordsByUser: map[string]string{},
		refreshTokens:   map[string]RefreshToken{},
		blacklist:       map[string]time.Time{},
		orgs:            map[string]Organization{},
		teams:           map[string]Team{},
		channels:        map[string]Channel{},
	}
}

func (s *MemoryStore) CreateTenantAndUser(tenant Tenant, user User, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.userIDsByEmail[user.Email]; exists {
		return errors.New("email already exists")
	}
	s.tenants[tenant.ID] = tenant
	s.users[user.ID] = user
	s.userIDsByEmail[user.Email] = user.ID
	s.passwordsByUser[user.ID] = passwordHash
	return nil
}

func (s *MemoryStore) UserByEmail(email string) (User, string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	userID, ok := s.userIDsByEmail[email]
	if !ok {
		return User{}, "", false
	}
	user := s.users[userID]
	return user, s.passwordsByUser[userID], true
}

func (s *MemoryStore) UserByID(userID string) (User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok := s.users[userID]
	return user, ok
}

func (s *MemoryStore) SaveRefreshToken(token RefreshToken) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshTokens[token.TokenHash] = token
}

func (s *MemoryStore) RefreshToken(tokenHash string) (RefreshToken, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	token, ok := s.refreshTokens[tokenHash]
	return token, ok
}

func (s *MemoryStore) RevokeRefreshToken(tokenHash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	token := s.refreshTokens[tokenHash]
	now := time.Now().UTC()
	token.RevokedAt = &now
	s.refreshTokens[tokenHash] = token
}

func (s *MemoryStore) BlacklistAccessToken(tokenHash string, expiresAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blacklist[tokenHash] = expiresAt
}

func (s *MemoryStore) IsBlacklisted(tokenHash string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expiresAt, ok := s.blacklist[tokenHash]
	return ok && time.Now().UTC().Before(expiresAt)
}

func (s *MemoryStore) SaveOrganization(org Organization) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orgs[org.ID] = org
}

func (s *MemoryStore) Organizations(tenantID string) []Organization {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Organization{}
	for _, org := range s.orgs {
		if org.TenantID == tenantID {
			out = append(out, org)
		}
	}
	return out
}

func (s *MemoryStore) SaveTeam(team Team) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teams[team.ID] = team
}

func (s *MemoryStore) Teams(tenantID string) []Team {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Team{}
	for _, team := range s.teams {
		if team.TenantID == tenantID {
			out = append(out, team)
		}
	}
	return out
}

func (s *MemoryStore) SaveChannel(channel Channel) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.channels[channel.ID] = channel
}

func (s *MemoryStore) Channels(tenantID string) []Channel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Channel{}
	for _, channel := range s.channels {
		if channel.TenantID == tenantID {
			out = append(out, channel)
		}
	}
	return out
}
