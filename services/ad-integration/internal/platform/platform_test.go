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

func TestLocalClientSyncsBaiduReportsWithSdkShape(t *testing.T) {
	client := NewLocalClient("baidu")
	result, err := client.Sync(context.Background(), SyncRequest{
		Account:     testAccount("baidu"),
		ReportTypes: DefaultReportTypes("baidu"),
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
	if len(result.Reports) != 7 {
		t.Fatalf("expected 7 report rows, got %d", len(result.Reports))
	}
	if result.Reports[0].Raw["header_model"] != "ApiRequestHeader" {
		t.Fatalf("expected Baidu request header shape, got %v", result.Reports[0].Raw["header_model"])
	}
	if result.Reports[0].Raw["sdk_service"] != "ReportService" {
		t.Fatalf("expected Baidu ReportService reference, got %v", result.Reports[0].Raw["sdk_service"])
	}
}

func TestLocalClientSyncsXiaohongshuReportsWithUnitAndRealtimeShape(t *testing.T) {
	client := NewLocalClient("xiaohongshu")
	result, err := client.Sync(context.Background(), SyncRequest{
		Account:     testAccount("xiaohongshu"),
		ReportTypes: DefaultReportTypes("xiaohongshu"),
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
	if len(result.Reports) != 8 {
		t.Fatalf("expected 8 report rows, got %d", len(result.Reports))
	}
	var sawUnit bool
	var sawRealtime bool
	for _, row := range result.Reports {
		if row.EntityType == "unit" && row.Raw["sdk_model"] == "offline.Request" {
			sawUnit = true
		}
		if row.ReportType == "account_realtime" && row.Raw["sdk_model"] == "realtime.AdvertiserRequest" {
			sawRealtime = true
		}
	}
	if !sawUnit {
		t.Fatal("expected xiaohongshu unit report rows")
	}
	if !sawRealtime {
		t.Fatal("expected xiaohongshu realtime advertiser report row")
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
