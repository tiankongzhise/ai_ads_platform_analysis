package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, databaseURL string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) CreateTenantAndUser(tenant Tenant, user User, passwordHash string) error {
	ctx := context.Background()
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `
		INSERT INTO auth.tenants (id, name, tenant_type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
	`, tenant.ID, tenant.Name, tenant.Type, tenant.Status, tenant.CreatedAt); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO auth.users (id, tenant_id, email, phone, display_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`, user.ID, user.TenantID, user.Email, emptyToNil(user.Phone), user.DisplayName, user.Status, user.CreatedAt); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO auth.password_credentials (user_id, password_hash, password_algo, updated_at)
		VALUES ($1, $2, $3, $4)
	`, user.ID, passwordHash, "pbkdf2_sha256", user.CreatedAt); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *PostgresStore) UserByEmail(email string) (User, string, bool) {
	var user User
	var passwordHash string
	err := s.pool.QueryRow(context.Background(), `
		SELECT u.id, u.tenant_id, u.email, COALESCE(u.phone, ''), u.display_name, u.status, u.created_at, c.password_hash
		FROM auth.users u
		JOIN auth.password_credentials c ON c.user_id = u.id
		WHERE lower(u.email) = lower($1)
	`, email).Scan(&user.ID, &user.TenantID, &user.Email, &user.Phone, &user.DisplayName, &user.Status, &user.CreatedAt, &passwordHash)
	return user, passwordHash, err == nil
}

func (s *PostgresStore) UserByID(userID string) (User, bool) {
	var user User
	err := s.pool.QueryRow(context.Background(), `
		SELECT id, tenant_id, email, COALESCE(phone, ''), display_name, status, created_at
		FROM auth.users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.TenantID, &user.Email, &user.Phone, &user.DisplayName, &user.Status, &user.CreatedAt)
	return user, err == nil
}

func (s *PostgresStore) TenantByID(tenantID string) (Tenant, bool) {
	var tenant Tenant
	err := s.pool.QueryRow(context.Background(), `
		SELECT id, name, tenant_type, status, created_at
		FROM auth.tenants
		WHERE id = $1
	`, tenantID).Scan(&tenant.ID, &tenant.Name, &tenant.Type, &tenant.Status, &tenant.CreatedAt)
	return tenant, err == nil
}

func (s *PostgresStore) SaveRefreshToken(token RefreshToken) {
	_, _ = s.pool.Exec(context.Background(), `
		INSERT INTO auth.refresh_tokens (id, tenant_id, user_id, token_hash, expires_at, rotation_parent_id, created_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), now())
	`, token.ID, token.TenantID, token.UserID, token.TokenHash, token.ExpiresAt, token.RotationParentID)
}

func (s *PostgresStore) RefreshToken(tokenHash string) (RefreshToken, bool) {
	var token RefreshToken
	err := s.pool.QueryRow(context.Background(), `
		SELECT id, tenant_id, user_id, token_hash, expires_at, revoked_at, COALESCE(rotation_parent_id::text, '')
		FROM auth.refresh_tokens
		WHERE token_hash = $1
	`, tokenHash).Scan(&token.ID, &token.TenantID, &token.UserID, &token.TokenHash, &token.ExpiresAt, &token.RevokedAt, &token.RotationParentID)
	return token, err == nil
}

func (s *PostgresStore) RevokeRefreshToken(tokenHash string) {
	_, _ = s.pool.Exec(context.Background(), `
		UPDATE auth.refresh_tokens
		SET revoked_at = COALESCE(revoked_at, now())
		WHERE token_hash = $1
	`, tokenHash)
}

func (s *PostgresStore) BlacklistAccessToken(tokenHash string, expiresAt time.Time) {
	// The initial migration does not create a persisted access-token blacklist table yet.
	// Access token blacklist remains in-memory until Redis is wired in the platform-ops phase.
}

func (s *PostgresStore) IsBlacklisted(tokenHash string) bool {
	return false
}

func (s *PostgresStore) SaveOrganization(org Organization) {
	_, _ = s.pool.Exec(context.Background(), `
		INSERT INTO auth.organizations (id, tenant_id, parent_id, name, org_type, status, created_at, updated_at)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8)
	`, org.ID, org.TenantID, org.ParentID, org.Name, org.OrgType, org.Status, org.CreatedAt, org.UpdatedAt)
}

func (s *PostgresStore) Organization(tenantID string, orgID string) (Organization, bool) {
	var org Organization
	err := s.pool.QueryRow(context.Background(), `
		SELECT id, tenant_id, COALESCE(parent_id::text, ''), name, org_type, status, created_at, updated_at
		FROM auth.organizations
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, orgID).Scan(&org.ID, &org.TenantID, &org.ParentID, &org.Name, &org.OrgType, &org.Status, &org.CreatedAt, &org.UpdatedAt)
	return org, err == nil
}

func (s *PostgresStore) Organizations(tenantID string) []Organization {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, tenant_id, COALESCE(parent_id::text, ''), name, org_type, status, created_at, updated_at
		FROM auth.organizations
		WHERE tenant_id = $1
		ORDER BY parent_id NULLS FIRST, created_at ASC
	`, tenantID)
	if err != nil {
		return []Organization{}
	}
	defer rows.Close()
	orgs := []Organization{}
	for rows.Next() {
		var org Organization
		if err := rows.Scan(&org.ID, &org.TenantID, &org.ParentID, &org.Name, &org.OrgType, &org.Status, &org.CreatedAt, &org.UpdatedAt); err == nil {
			orgs = append(orgs, org)
		}
	}
	return orgs
}

func (s *PostgresStore) OrganizationTree(tenantID string) []OrganizationNode {
	return buildOrganizationTree(s.Organizations(tenantID))
}

func (s *PostgresStore) UpdateOrganization(tenantID string, orgID string, name string, orgType string, status string, updatedAt time.Time) (Organization, error) {
	org, ok := s.Organization(tenantID, orgID)
	if !ok {
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
	_, err := s.pool.Exec(context.Background(), `
		UPDATE auth.organizations
		SET name = $3, org_type = $4, status = $5, updated_at = $6
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, orgID, org.Name, org.OrgType, org.Status, org.UpdatedAt)
	return org, err
}

func (s *PostgresStore) MoveOrganization(tenantID string, orgID string, parentID string, updatedAt time.Time) (Organization, error) {
	orgs := s.Organizations(tenantID)
	orgByID := map[string]Organization{}
	for _, org := range orgs {
		orgByID[org.ID] = org
	}
	org, ok := orgByID[orgID]
	if !ok {
		return Organization{}, ErrNotFound
	}
	if parentID == orgID {
		return Organization{}, ErrCycleMove
	}
	if parentID != "" {
		parent, ok := orgByID[parentID]
		if !ok {
			return Organization{}, ErrInvalidParent
		}
		for current := parent; current.ParentID != ""; {
			if current.ParentID == orgID {
				return Organization{}, ErrCycleMove
			}
			next, ok := orgByID[current.ParentID]
			if !ok {
				break
			}
			current = next
		}
	}
	org.ParentID = parentID
	org.UpdatedAt = updatedAt
	_, err := s.pool.Exec(context.Background(), `
		UPDATE auth.organizations
		SET parent_id = NULLIF($3, '')::uuid, updated_at = $4
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, orgID, parentID, updatedAt)
	return org, err
}

func (s *PostgresStore) OrganizationSummary(tenantID string, orgID string) (OrganizationSummary, error) {
	orgs := s.Organizations(tenantID)
	if orgID != "" {
		found := false
		for _, org := range orgs {
			if org.ID == orgID {
				found = true
				break
			}
		}
		if !found {
			return OrganizationSummary{}, ErrNotFound
		}
	}
	ids := subtreeIDs(orgs, orgID)
	inScope := func(id string) bool {
		if orgID == "" {
			return true
		}
		return ids[id]
	}
	summary := OrganizationSummary{OrganizationID: orgID}
	for _, org := range orgs {
		if inScope(org.ID) {
			summary.OrgCount++
		}
	}
	for _, team := range s.Teams(tenantID) {
		if inScope(team.OrganizationID) {
			summary.TeamCount++
		}
	}
	for _, channel := range s.Channels(tenantID) {
		if inScope(channel.OrganizationID) {
			summary.ChannelCount++
		}
	}
	return summary, nil
}

func (s *PostgresStore) SaveTeam(team Team) {
	_, _ = s.pool.Exec(context.Background(), `
		INSERT INTO auth.admission_teams (id, tenant_id, organization_id, name, leader_user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6, $7, $8)
	`, team.ID, team.TenantID, team.OrganizationID, team.Name, team.LeaderUserID, team.Status, team.CreatedAt, team.UpdatedAt)
}

func (s *PostgresStore) UpdateTeam(tenantID string, teamID string, name string, leaderUserID string, status string, updatedAt time.Time) (Team, error) {
	team, ok := findTeamByID(s.Teams(tenantID), teamID)
	if !ok {
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
	_, err := s.pool.Exec(context.Background(), `
		UPDATE auth.admission_teams
		SET name = $3, leader_user_id = NULLIF($4, '')::uuid, status = $5, updated_at = $6
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, teamID, team.Name, team.LeaderUserID, team.Status, team.UpdatedAt)
	return team, err
}

func (s *PostgresStore) Teams(tenantID string) []Team {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, tenant_id, organization_id, name, COALESCE(leader_user_id::text, ''), status, created_at, updated_at
		FROM auth.admission_teams
		WHERE tenant_id = $1
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return []Team{}
	}
	defer rows.Close()
	teams := []Team{}
	for rows.Next() {
		var team Team
		if err := rows.Scan(&team.ID, &team.TenantID, &team.OrganizationID, &team.Name, &team.LeaderUserID, &team.Status, &team.CreatedAt, &team.UpdatedAt); err == nil {
			teams = append(teams, team)
		}
	}
	return teams
}

func (s *PostgresStore) SaveChannel(channel Channel) {
	_, _ = s.pool.Exec(context.Background(), `
		INSERT INTO auth.ad_channels (id, tenant_id, organization_id, platform, display_name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, channel.ID, channel.TenantID, channel.OrganizationID, channel.Platform, channel.DisplayName, channel.Status, channel.CreatedAt, channel.UpdatedAt)
}

func (s *PostgresStore) UpdateChannel(tenantID string, channelID string, displayName string, status string, updatedAt time.Time) (Channel, error) {
	channel, ok := findChannelByID(s.Channels(tenantID), channelID)
	if !ok {
		return Channel{}, ErrNotFound
	}
	if displayName != "" {
		channel.DisplayName = displayName
	}
	if status != "" {
		channel.Status = status
	}
	channel.UpdatedAt = updatedAt
	_, err := s.pool.Exec(context.Background(), `
		UPDATE auth.ad_channels
		SET display_name = $3, status = $4, updated_at = $5
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, channelID, channel.DisplayName, channel.Status, channel.UpdatedAt)
	return channel, err
}

func (s *PostgresStore) Channels(tenantID string) []Channel {
	rows, err := s.pool.Query(context.Background(), `
		SELECT id, tenant_id, organization_id, platform, display_name, status, created_at, updated_at
		FROM auth.ad_channels
		WHERE tenant_id = $1
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return []Channel{}
	}
	defer rows.Close()
	channels := []Channel{}
	for rows.Next() {
		var channel Channel
		if err := rows.Scan(&channel.ID, &channel.TenantID, &channel.OrganizationID, &channel.Platform, &channel.DisplayName, &channel.Status, &channel.CreatedAt, &channel.UpdatedAt); err == nil {
			channels = append(channels, channel)
		}
	}
	return channels
}

func buildOrganizationTree(orgs []Organization) []OrganizationNode {
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

func subtreeIDs(orgs []Organization, rootID string) map[string]bool {
	ids := map[string]bool{}
	if rootID == "" {
		for _, org := range orgs {
			ids[org.ID] = true
		}
		return ids
	}
	ids[rootID] = true
	changed := true
	for changed {
		changed = false
		for _, org := range orgs {
			if ids[org.ParentID] && !ids[org.ID] {
				ids[org.ID] = true
				changed = true
			}
		}
	}
	return ids
}

func findTeamByID(items []Team, id string) (Team, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return Team{}, false
}

func findChannelByID(items []Channel, id string) (Channel, bool) {
	for _, item := range items {
		if item.ID == id {
			return item, true
		}
	}
	return Channel{}, false
}

func emptyToNil(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func postgresNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return fmt.Errorf("postgres store: %w", err)
}

var _ Repository = (*PostgresStore)(nil)
