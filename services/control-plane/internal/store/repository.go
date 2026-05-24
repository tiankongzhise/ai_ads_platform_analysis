package store

import "time"

type Repository interface {
	CreateTenantAndUser(tenant Tenant, user User, passwordHash string) error
	UserByEmail(email string) (User, string, bool)
	UserByID(userID string) (User, bool)
	TenantByID(tenantID string) (Tenant, bool)

	SaveRefreshToken(token RefreshToken)
	RefreshToken(tokenHash string) (RefreshToken, bool)
	RevokeRefreshToken(tokenHash string)
	BlacklistAccessToken(tokenHash string, expiresAt time.Time)
	IsBlacklisted(tokenHash string) bool

	SaveOrganization(org Organization)
	Organization(tenantID string, orgID string) (Organization, bool)
	Organizations(tenantID string) []Organization
	OrganizationTree(tenantID string) []OrganizationNode
	UpdateOrganization(tenantID string, orgID string, name string, orgType string, status string, updatedAt time.Time) (Organization, error)
	MoveOrganization(tenantID string, orgID string, parentID string, updatedAt time.Time) (Organization, error)
	OrganizationSummary(tenantID string, orgID string) (OrganizationSummary, error)

	SaveTeam(team Team)
	UpdateTeam(tenantID string, teamID string, name string, leaderUserID string, status string, updatedAt time.Time) (Team, error)
	Teams(tenantID string) []Team

	SaveChannel(channel Channel)
	UpdateChannel(tenantID string, channelID string, displayName string, status string, updatedAt time.Time) (Channel, error)
	Channels(tenantID string) []Channel
}

var _ Repository = (*MemoryStore)(nil)
