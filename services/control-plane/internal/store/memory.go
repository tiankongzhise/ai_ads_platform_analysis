package store

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidParent = errors.New("invalid parent organization")
	ErrCycleMove     = errors.New("organization move would create a cycle")
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

func (s *MemoryStore) Organization(tenantID string, orgID string) (Organization, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	org, ok := s.orgs[orgID]
	return org, ok && org.TenantID == tenantID
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
	sort.Slice(out, func(i, j int) bool {
		if out[i].ParentID == out[j].ParentID {
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		}
		return out[i].ParentID < out[j].ParentID
	})
	return out
}

func (s *MemoryStore) OrganizationTree(tenantID string) []OrganizationNode {
	orgs := s.Organizations(tenantID)
	childrenByParent := map[string][]OrganizationNode{}
	for _, org := range orgs {
		node := OrganizationNode{Organization: org, Children: []OrganizationNode{}}
		childrenByParent[org.ParentID] = append(childrenByParent[org.ParentID], node)
	}
	var attach func(parentID string) []OrganizationNode
	attach = func(parentID string) []OrganizationNode {
		nodes := childrenByParent[parentID]
		for index := range nodes {
			nodes[index].Children = attach(nodes[index].ID)
		}
		return nodes
	}
	return attach("")
}

func (s *MemoryStore) UpdateOrganization(tenantID string, orgID string, name string, orgType string, status string, updatedAt time.Time) (Organization, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	org, ok := s.orgs[orgID]
	if !ok || org.TenantID != tenantID {
		return Organization{}, ErrNotFound
	}
	if name != "" {
		org.Name = name
	}
	if orgType != "" {
		org.OrgType = orgType
	}
	if status != "" {
		org.Status = status
	}
	org.UpdatedAt = updatedAt
	s.orgs[orgID] = org
	return org, nil
}

func (s *MemoryStore) MoveOrganization(tenantID string, orgID string, parentID string, updatedAt time.Time) (Organization, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	org, ok := s.orgs[orgID]
	if !ok || org.TenantID != tenantID {
		return Organization{}, ErrNotFound
	}
	if parentID == orgID {
		return Organization{}, ErrCycleMove
	}
	if parentID != "" {
		parent, ok := s.orgs[parentID]
		if !ok || parent.TenantID != tenantID {
			return Organization{}, ErrInvalidParent
		}
		for current := parent; current.ParentID != ""; {
			if current.ParentID == orgID {
				return Organization{}, ErrCycleMove
			}
			next, ok := s.orgs[current.ParentID]
			if !ok || next.TenantID != tenantID {
				break
			}
			current = next
		}
	}
	org.ParentID = parentID
	org.UpdatedAt = updatedAt
	s.orgs[orgID] = org
	return org, nil
}

func (s *MemoryStore) OrganizationSummary(tenantID string, orgID string) (OrganizationSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if orgID != "" {
		org, ok := s.orgs[orgID]
		if !ok || org.TenantID != tenantID {
			return OrganizationSummary{}, ErrNotFound
		}
	}
	descendants := map[string]bool{}
	if orgID != "" {
		descendants[orgID] = true
		changed := true
		for changed {
			changed = false
			for _, org := range s.orgs {
				if org.TenantID == tenantID && descendants[org.ParentID] && !descendants[org.ID] {
					descendants[org.ID] = true
					changed = true
				}
			}
		}
	}
	inScope := func(candidateOrgID string) bool {
		if orgID == "" {
			return true
		}
		return descendants[candidateOrgID]
	}
	summary := OrganizationSummary{OrganizationID: orgID}
	for _, org := range s.orgs {
		if org.TenantID == tenantID && inScope(org.ID) {
			summary.OrgCount++
		}
	}
	for _, team := range s.teams {
		if team.TenantID == tenantID && inScope(team.OrganizationID) {
			summary.TeamCount++
		}
	}
	for _, channel := range s.channels {
		if channel.TenantID == tenantID && inScope(channel.OrganizationID) {
			summary.ChannelCount++
		}
	}
	return summary, nil
}

func (s *MemoryStore) SaveTeam(team Team) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.teams[team.ID] = team
}

func (s *MemoryStore) UpdateTeam(tenantID string, teamID string, name string, leaderUserID string, status string, updatedAt time.Time) (Team, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	team, ok := s.teams[teamID]
	if !ok || team.TenantID != tenantID {
		return Team{}, ErrNotFound
	}
	if name != "" {
		team.Name = name
	}
	if leaderUserID != "" {
		team.LeaderUserID = leaderUserID
	}
	if status != "" {
		team.Status = status
	}
	team.UpdatedAt = updatedAt
	s.teams[teamID] = team
	return team, nil
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

func (s *MemoryStore) UpdateChannel(tenantID string, channelID string, displayName string, status string, updatedAt time.Time) (Channel, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	channel, ok := s.channels[channelID]
	if !ok || channel.TenantID != tenantID {
		return Channel{}, ErrNotFound
	}
	if displayName != "" {
		channel.DisplayName = displayName
	}
	if status != "" {
		channel.Status = status
	}
	channel.UpdatedAt = updatedAt
	s.channels[channelID] = channel
	return channel, nil
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
