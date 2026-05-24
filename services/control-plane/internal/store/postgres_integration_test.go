package store

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestPostgresStoreIntegration(t *testing.T) {
	databaseURL := os.Getenv("CONTROL_PLANE_INTEGRATION_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set CONTROL_PLANE_INTEGRATION_DATABASE_URL to run PostgreSQL integration test")
	}
	repository, err := NewPostgresStore(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer repository.Close()

	now := time.Now().UTC()
	tenant := Tenant{ID: "11111111-1111-4111-8111-111111111111", Name: "集成测试租户", Type: "formal", Status: "active", CreatedAt: now}
	user := User{ID: "22222222-2222-4222-8222-222222222222", TenantID: tenant.ID, Email: "integration@example.com", DisplayName: "Integration", Status: "active", CreatedAt: now}
	_ = repository.CreateTenantAndUser(tenant, user, "hash")

	root := Organization{ID: "33333333-3333-4333-8333-333333333333", TenantID: tenant.ID, Name: "集团", OrgType: "group", Status: "active", CreatedAt: now, UpdatedAt: now}
	repository.SaveOrganization(root)
	team := Team{ID: "44444444-4444-4444-8444-444444444444", TenantID: tenant.ID, OrganizationID: root.ID, Name: "招生一队", Status: "active", CreatedAt: now, UpdatedAt: now}
	repository.SaveTeam(team)

	summary, err := repository.OrganizationSummary(tenant.ID, root.ID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.OrgCount == 0 || summary.TeamCount == 0 {
		t.Fatalf("expected persisted org and team to be counted, got %+v", summary)
	}
}
