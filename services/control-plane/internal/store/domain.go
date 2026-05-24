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
}

type Team struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	OrganizationID string    `json:"organization_id"`
	Name           string    `json:"name"`
	LeaderUserID   string    `json:"leader_user_id,omitempty"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type Channel struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	OrganizationID string    `json:"organization_id"`
	Platform       string    `json:"platform"`
	DisplayName    string    `json:"display_name"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}
