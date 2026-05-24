package store

import (
	"errors"
	"testing"
	"time"
)

func TestMoveOrganizationRejectsCycle(t *testing.T) {
	memory := NewMemoryStore()
	now := time.Now().UTC()
	root := Organization{ID: "root", TenantID: "tenant-1", Name: "集团", OrgType: "group", Status: "active", CreatedAt: now}
	school := Organization{ID: "school", TenantID: "tenant-1", ParentID: "root", Name: "学校", OrgType: "school", Status: "active", CreatedAt: now}
	campus := Organization{ID: "campus", TenantID: "tenant-1", ParentID: "school", Name: "校区", OrgType: "campus", Status: "active", CreatedAt: now}
	memory.SaveOrganization(root)
	memory.SaveOrganization(school)
	memory.SaveOrganization(campus)

	if _, err := memory.MoveOrganization("tenant-1", "root", "campus", now); !errors.Is(err, ErrCycleMove) {
		t.Fatalf("expected ErrCycleMove, got %v", err)
	}
}

func TestOrganizationSummaryCountsSubtreeResources(t *testing.T) {
	memory := NewMemoryStore()
	now := time.Now().UTC()
	memory.SaveOrganization(Organization{ID: "root", TenantID: "tenant-1", Name: "集团", OrgType: "group", Status: "active", CreatedAt: now})
	memory.SaveOrganization(Organization{ID: "school", TenantID: "tenant-1", ParentID: "root", Name: "学校", OrgType: "school", Status: "active", CreatedAt: now})
	memory.SaveOrganization(Organization{ID: "other", TenantID: "tenant-1", Name: "其他学校", OrgType: "school", Status: "active", CreatedAt: now})
	memory.SaveTeam(Team{ID: "team-1", TenantID: "tenant-1", OrganizationID: "school", Name: "一队", Status: "active", CreatedAt: now})
	memory.SaveTeam(Team{ID: "team-2", TenantID: "tenant-1", OrganizationID: "other", Name: "二队", Status: "active", CreatedAt: now})
	memory.SaveChannel(Channel{ID: "channel-1", TenantID: "tenant-1", OrganizationID: "school", Platform: "douyin", DisplayName: "抖音", Status: "active", CreatedAt: now})

	summary, err := memory.OrganizationSummary("tenant-1", "root")
	if err != nil {
		t.Fatal(err)
	}
	if summary.OrgCount != 2 {
		t.Fatalf("expected 2 orgs in root subtree, got %d", summary.OrgCount)
	}
	if summary.TeamCount != 1 {
		t.Fatalf("expected 1 team in root subtree, got %d", summary.TeamCount)
	}
	if summary.ChannelCount != 1 {
		t.Fatalf("expected 1 channel in root subtree, got %d", summary.ChannelCount)
	}
}
