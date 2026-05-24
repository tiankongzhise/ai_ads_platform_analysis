package platform

import (
	"context"
	"testing"
	"time"

	"eduadcrm/services/ad-integration/internal/store"
)

func TestLocalClientSyncsDouyinReportsWithRawShape(t *testing.T) {
	client := NewLocalClient("douyin")
	result, err := client.Sync(context.Background(), SyncRequest{
		Account:     testAccount("douyin"),
		ReportTypes: DefaultReportTypes("douyin"),
		DateFrom:    "2026-05-20",
		DateTo:      "2026-05-20",
		Now:         time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Entities) != 7 {
		t.Fatalf("expected 7 entities, got %d", len(result.Entities))
	}
	if len(result.Reports) != 10 {
		t.Fatalf("expected 10 report rows, got %d", len(result.Reports))
	}
	if result.Reports[0].Raw["sdk_api"] == "" {
		t.Fatalf("expected sdk_api in raw payload")
	}
}

func TestLocalClientSyncsTencentHourlyReports(t *testing.T) {
	client := NewLocalClient("tencent")
	result, err := client.Sync(context.Background(), SyncRequest{
		Account:     testAccount("tencent"),
		ReportTypes: []string{"account_hourly"},
		DateFrom:    "2026-05-20",
		DateTo:      "2026-05-20",
		Now:         time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Reports) != 3 {
		t.Fatalf("expected 3 hourly rows, got %d", len(result.Reports))
	}
	if result.Reports[0].Raw["sdk_service"] != "HourlyReports().Get" {
		t.Fatalf("expected Tencent hourly SDK reference, got %v", result.Reports[0].Raw["sdk_service"])
	}
}

func testAccount(platform string) store.Account {
	return store.Account{
		ID:                "account-1",
		TenantID:          "tenant-1",
		Platform:          platform,
		ExternalAccountID: platform + "-external-1",
		AccountName:       platform + " account",
		Status:            "active",
	}
}
