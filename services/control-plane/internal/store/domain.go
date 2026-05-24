package store

import "time"

type Organization struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	ParentID  string    `json:"parent_id,omitempty"`
	Name      string    `json:"name"`
	OrgType   string    `json:"org_type"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrganizationNode struct {
	Organization
	Children []OrganizationNode `json:"children"`
}

type OrganizationSummary struct {
	OrganizationID string `json:"organization_id"`
	OrgCount       int    `json:"org_count"`
	TeamCount      int    `json:"team_count"`
	ChannelCount   int    `json:"channel_count"`
}

type Team struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	LeaderUserID   string    `json:"leader_user_id,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type Channel struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	OrganizationID string    `json:"organization_id"`
	Platform       string    `json:"platform"`
	DisplayName    string    `json:"display_name"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
