package platform

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"eduadcrm/services/ad-integration/internal/store"
)

type SyncRequest struct {
	Account     store.Account
	ReportTypes []string
	DateFrom    string
	DateTo      string
	Now         time.Time
}

type SyncResult struct {
	Entities []store.AdEntity
	Reports  []store.RawReportRow
	Meta     map[string]any
}

type Client interface {
	Platform() string
	Sync(ctx context.Context, req SyncRequest) (SyncResult, error)
}

type Registry struct {
	clients map[string]Client
}

func NewRegistry() *Registry {
	registry := &Registry{clients: map[string]Client{}}
	registry.Register(NewLocalClient("douyin"))
	registry.Register(NewLocalClient("tencent"))
	return registry
}

func (r *Registry) Register(client Client) {
	r.clients[client.Platform()] = client
}

func (r *Registry) Client(platform string) (Client, bool) {
	client, ok := r.clients[strings.ToLower(platform)]
	return client, ok
}

func DefaultReportTypes(platform string) []string {
	switch strings.ToLower(platform) {
	case "douyin", "tencent":
		return []string{"account_daily", "campaign_daily", "adgroup_daily", "account_hourly"}
	default:
		return []string{"account_daily"}
	}
}

type LocalClient struct {
	platform string
}

func NewLocalClient(platform string) *LocalClient {
	return &LocalClient{platform: strings.ToLower(platform)}
}

func (c *LocalClient) Platform() string {
	return c.platform
}

func (c *LocalClient) Sync(ctx context.Context, req SyncRequest) (SyncResult, error) {
	if req.Account.ID == "" {
		return SyncResult{}, errors.New("account is required")
	}
	if req.Account.Platform != c.platform {
		return SyncResult{}, fmt.Errorf("account platform %s does not match client %s", req.Account.Platform, c.platform)
	}
	if req.Now.IsZero() {
		req.Now = time.Now().UTC()
	}
	dates, err := syncDates(req.DateFrom, req.DateTo, req.Now)
	if err != nil {
		return SyncResult{}, err
	}
	reportTypes := req.ReportTypes
	if len(reportTypes) == 0 {
		reportTypes = DefaultReportTypes(c.platform)
	}
	entities := c.entities(req.Account, req.Now)
	entityByType := groupEntities(entities)
	rows := []store.RawReportRow{}
	for _, reportType := range reportTypes {
		select {
		case <-ctx.Done():
			return SyncResult{}, ctx.Err()
		default:
		}
		rows = append(rows, c.reportRows(req.Account, reportType, dates, entityByType, req.Now)...)
	}
	return SyncResult{
		Entities: entities,
		Reports:  rows,
		Meta: map[string]any{
			"client_mode":       "local_sdk_shape",
			"platform":          c.platform,
			"date_count":        len(dates),
			"report_type_count": len(reportTypes),
			"sdk_reference":     c.sdkReferences(),
		},
	}, nil
}

func (c *LocalClient) entities(account store.Account, now time.Time) []store.AdEntity {
	accountExternalID := account.ExternalAccountID
	if accountExternalID == "" {
		accountExternalID = "demo_" + c.platform + "_account"
	}
	accountName := account.AccountName
	if accountName == "" {
		accountName = c.displayName() + "账户"
	}
	campaigns := []struct {
		id     string
		name   string
		status string
	}{
		{id: accountExternalID + "_campaign_1", name: "2026春招线索计划", status: "enable"},
		{id: accountExternalID + "_campaign_2", name: "开放日到校计划", status: "enable"},
	}
	entities := []store.AdEntity{
		{
			ID:         entityID(account.ID, c.platform, "account", accountExternalID),
			TenantID:   account.TenantID,
			AccountID:  account.ID,
			Platform:   c.platform,
			EntityType: "account",
			ExternalID: accountExternalID,
			Name:       accountName,
			Status:     "active",
			Raw:        c.accountRaw(accountExternalID, accountName),
			SyncedAt:   now,
		},
	}
	for index, campaign := range campaigns {
		entities = append(entities, store.AdEntity{
			ID:               entityID(account.ID, c.platform, "campaign", campaign.id),
			TenantID:         account.TenantID,
			AccountID:        account.ID,
			Platform:         c.platform,
			EntityType:       "campaign",
			ExternalID:       campaign.id,
			ParentExternalID: accountExternalID,
			Name:             campaign.name,
			Status:           campaign.status,
			Raw:              c.campaignRaw(accountExternalID, campaign.id, campaign.name, campaign.status),
			SyncedAt:         now,
		})
		for unit := 1; unit <= 2; unit++ {
			unitID := fmt.Sprintf("%s_adgroup_%d", campaign.id, unit)
			entities = append(entities, store.AdEntity{
				ID:               entityID(account.ID, c.platform, "adgroup", unitID),
				TenantID:         account.TenantID,
				AccountID:        account.ID,
				Platform:         c.platform,
				EntityType:       "adgroup",
				ExternalID:       unitID,
				ParentExternalID: campaign.id,
				Name:             fmt.Sprintf("%s-单元%d", campaign.name, unit),
				Status:           "enable",
				Raw:              c.adgroupRaw(accountExternalID, campaign.id, unitID, index, unit),
				SyncedAt:         now,
			})
		}
	}
	return entities
}

func (c *LocalClient) reportRows(account store.Account, reportType string, dates []string, entities map[string][]store.AdEntity, now time.Time) []store.RawReportRow {
	entityType, granularity := reportShape(reportType)
	targets := entities[entityType]
	if len(targets) == 0 {
		return nil
	}
	rows := []store.RawReportRow{}
	for dateIndex, date := range dates {
		if granularity == "hour" {
			for _, hour := range []string{"00", "08", "16"} {
				rows = append(rows, c.rowsForTargets(account, reportType, granularity, date, hour, dateIndex, targets, now)...)
			}
			continue
		}
		rows = append(rows, c.rowsForTargets(account, reportType, granularity, date, "", dateIndex, targets, now)...)
	}
	return rows
}

func (c *LocalClient) rowsForTargets(account store.Account, reportType string, granularity string, date string, hour string, dateIndex int, targets []store.AdEntity, now time.Time) []store.RawReportRow {
	rows := make([]store.RawReportRow, 0, len(targets))
	for index, target := range targets {
		metrics := deterministicMetrics(c.platform, reportType, target.ExternalID, dateIndex, index, hour)
		row := store.RawReportRow{
			ID:               reportID(account.ID, c.platform, reportType, target.ExternalID, date, hour),
			TenantID:         account.TenantID,
			AccountID:        account.ID,
			Platform:         c.platform,
			ReportType:       reportType,
			Granularity:      granularity,
			StatDate:         date,
			StatHour:         hour,
			EntityType:       target.EntityType,
			ExternalEntityID: target.ExternalID,
			Metrics:          metrics,
			Raw:              c.reportRaw(reportType, date, hour, target, metrics),
			SyncedAt:         now,
		}
		rows = append(rows, row)
	}
	return rows
}

func (c *LocalClient) accountRaw(externalID string, name string) map[string]any {
	if c.platform == "douyin" {
		return map[string]any{"advertiser_id": externalID, "advertiser_name": name, "status": "STATUS_ENABLE", "sdk_model": "AdvertiserInfoV2"}
	}
	return map[string]any{"account_id": externalID, "account_name": name, "account_type": "ACCOUNT_TYPE_ADVERTISER", "sdk_model": "AuthorizerStruct"}
}

func (c *LocalClient) campaignRaw(accountID string, campaignID string, name string, status string) map[string]any {
	if c.platform == "douyin" {
		return map[string]any{"advertiser_id": accountID, "campaign_id": campaignID, "campaign_name": name, "opt_status": status, "sdk_api": "CampaignGetV2Api"}
	}
	return map[string]any{"account_id": accountID, "campaign_id": campaignID, "campaign_name": name, "configured_status": status, "sdk_api": "Campaigns().Get"}
}

func (c *LocalClient) adgroupRaw(accountID string, campaignID string, adgroupID string, campaignIndex int, unit int) map[string]any {
	if c.platform == "douyin" {
		return map[string]any{"advertiser_id": accountID, "campaign_id": campaignID, "ad_id": adgroupID, "name": fmt.Sprintf("douyin_ad_%d_%d", campaignIndex+1, unit), "sdk_api": "AdGetV2Api"}
	}
	return map[string]any{"account_id": accountID, "campaign_id": campaignID, "adgroup_id": adgroupID, "adgroup_name": fmt.Sprintf("tencent_adgroup_%d_%d", campaignIndex+1, unit), "sdk_api": "Adgroups().Get"}
}

func (c *LocalClient) reportRaw(reportType string, date string, hour string, target store.AdEntity, metrics store.ReportMetrics) map[string]any {
	if c.platform == "douyin" {
		raw := map[string]any{
			"stat_cost":         metrics.Cost,
			"show_cnt":          metrics.Impressions,
			"click_cnt":         metrics.Clicks,
			"convert_cnt":       metrics.Conversions,
			"time_granularity":  reportGranularity(reportType),
			"stat_datetime":     date,
			"sdk_response_body": "data.list[]",
		}
		if target.EntityType == "account" {
			raw["advertiser_id"] = target.ExternalID
			raw["sdk_api"] = "ReportAdvertiserGetV2Api"
		}
		if target.EntityType == "campaign" {
			raw["campaign_id"] = target.ExternalID
			raw["sdk_api"] = "ReportCampaignGetV2Api"
		}
		if target.EntityType == "adgroup" {
			raw["ad_id"] = target.ExternalID
			raw["sdk_api"] = "ReportAdGetV2Api"
		}
		if hour != "" {
			raw["stat_datetime"] = date + " " + hour + ":00:00"
		}
		return raw
	}
	raw := map[string]any{
		"date":              date,
		"view_count":        metrics.Impressions,
		"valid_click_count": metrics.Clicks,
		"cost":              metrics.Cost,
		"conversions":       metrics.Conversions,
		"level":             tencentLevel(target.EntityType),
		"sdk_service":       "DailyReports().Get",
	}
	if hour != "" {
		raw["hour"] = hour
		raw["sdk_service"] = "HourlyReports().Get"
	}
	raw["account_id"] = target.ExternalID
	if target.EntityType == "campaign" {
		raw["campaign_id"] = target.ExternalID
	}
	if target.EntityType == "adgroup" {
		raw["adgroup_id"] = target.ExternalID
	}
	return raw
}

func (c *LocalClient) sdkReferences() []string {
	if c.platform == "douyin" {
		return []string{"ReportAdvertiserGetV2Api", "ReportCampaignGetV2Api", "ReportAdGetV2Api"}
	}
	return []string{"DailyReports().Get", "HourlyReports().Get"}
}

func (c *LocalClient) displayName() string {
	if c.platform == "douyin" {
		return "抖音/巨量引擎"
	}
	return "腾讯广告"
}

func syncDates(dateFrom string, dateTo string, now time.Time) ([]string, error) {
	if dateTo == "" {
		dateTo = now.Format("2006-01-02")
	}
	if dateFrom == "" {
		dateFrom = now.AddDate(0, 0, -7).Format("2006-01-02")
	}
	start, err := time.Parse("2006-01-02", dateFrom)
	if err != nil {
		return nil, fmt.Errorf("invalid date_from: %w", err)
	}
	end, err := time.Parse("2006-01-02", dateTo)
	if err != nil {
		return nil, fmt.Errorf("invalid date_to: %w", err)
	}
	if end.Before(start) {
		return nil, errors.New("date_to must be greater than or equal to date_from")
	}
	if end.Sub(start) > 31*24*time.Hour {
		return nil, errors.New("date range must be no more than 31 days")
	}
	dates := []string{}
	for day := start; !day.After(end); day = day.AddDate(0, 0, 1) {
		dates = append(dates, day.Format("2006-01-02"))
	}
	return dates, nil
}

func groupEntities(entities []store.AdEntity) map[string][]store.AdEntity {
	grouped := map[string][]store.AdEntity{}
	for _, entity := range entities {
		grouped[entity.EntityType] = append(grouped[entity.EntityType], entity)
	}
	return grouped
}

func reportShape(reportType string) (string, string) {
	switch reportType {
	case "campaign_daily":
		return "campaign", "day"
	case "adgroup_daily":
		return "adgroup", "day"
	case "account_hourly":
		return "account", "hour"
	default:
		return "account", "day"
	}
}

func reportGranularity(reportType string) string {
	if strings.HasSuffix(reportType, "_hourly") {
		return "STAT_TIME_GRANULARITY_HOURLY"
	}
	return "STAT_TIME_GRANULARITY_DAILY"
}

func tencentLevel(entityType string) string {
	switch entityType {
	case "campaign":
		return "REPORT_LEVEL_CAMPAIGN"
	case "adgroup":
		return "REPORT_LEVEL_ADGROUP"
	default:
		return "REPORT_LEVEL_ACCOUNT"
	}
}

func deterministicMetrics(platform string, reportType string, externalID string, dateIndex int, entityIndex int, hour string) store.ReportMetrics {
	seed := stableNumber(platform + ":" + reportType + ":" + externalID)
	hourFactor := 1
	if hour != "" {
		hourFactor = 1 + int(stableNumber(hour)%3)
	}
	impressions := int64(800+seed%700+uint64(dateIndex*97+entityIndex*53)) * int64(hourFactor)
	clicks := impressions / int64(18+seed%7)
	conversions := clicks / int64(8+seed%5)
	cost := float64(impressions) * (0.018 + float64(seed%9)/1000)
	if strings.Contains(reportType, "campaign") {
		cost *= 1.25
	}
	if strings.Contains(reportType, "adgroup") {
		cost *= 0.7
	}
	return store.ReportMetrics{
		Cost:        round2(cost),
		Impressions: impressions,
		Clicks:      clicks,
		Conversions: conversions,
	}
}

func stableNumber(value string) uint64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(value))
	return hash.Sum64()
}

func round2(value float64) float64 {
	return float64(int(value*100+0.5)) / 100
}

func entityID(accountID string, platform string, entityType string, externalID string) string {
	return stableUUID("entity:" + accountID + ":" + platform + ":" + entityType + ":" + externalID)
}

func reportID(accountID string, platform string, reportType string, externalID string, date string, hour string) string {
	return stableUUID("report:" + accountID + ":" + platform + ":" + reportType + ":" + externalID + ":" + date + ":" + hour)
}

func stableUUID(value string) string {
	sum := stableNumber(value)
	return fmt.Sprintf("00000000-0000-4000-8%03x-%012x", sum&0xfff, (sum>>12)&0xffffffffffff)
}
