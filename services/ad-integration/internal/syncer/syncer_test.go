package syncer

import (
	"context"
	"testing"

	"eduadcrm/services/ad-integration/internal/store"
)

func TestRunCreatesDemoAccountAndStoresSyncArtifacts(t *testing.T) {
	repository := store.NewMemoryStore()
	service := NewService(repository, nil)
	result, err := service.Run(context.Background(), RunRequest{
		Platform: "douyin",
		DateFrom: "2026-05-20",
		DateTo:   "2026-05-20",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Job.Status != "success" {
		t.Fatalf("expected success job, got %s", result.Job.Status)
	}
	if result.EntityCount != 7 {
		t.Fatalf("expected 7 entities, got %d", result.EntityCount)
	}
	if result.ReportRows == 0 {
		t.Fatal("expected report rows")
	}
	if len(repository.Accounts()) != 1 {
		t.Fatalf("expected demo account to be saved")
	}
	if len(repository.RawReportRows(store.RawReportFilter{Platform: "douyin"})) != result.ReportRows {
		t.Fatalf("expected stored report rows to match result")
	}
}
