package syncer

import (
	"context"
	"errors"
	"strings"
	"time"

	"eduadcrm/services/ad-integration/internal/platform"
	"eduadcrm/services/ad-integration/internal/store"
)

type Service struct {
	store    store.Repository
	registry *platform.Registry
}

type RunRequest struct {
	AccountID   string   `json:"account_id"`
	Platform    string   `json:"platform"`
	ReportTypes []string `json:"report_types"`
	DateFrom    string   `json:"date_from"`
	DateTo      string   `json:"date_to"`
	Reason      string   `json:"reason"`
}

type RunResult struct {
	Job          store.SyncJob `json:"job"`
	EntityCount  int           `json:"entity_count"`
	ReportRows   int           `json:"report_rows"`
	ReportTypes  []string      `json:"report_types"`
	ClientMode   string        `json:"client_mode"`
	SdkReference []string      `json:"sdk_reference"`
}

func NewService(repository store.Repository, registry *platform.Registry) *Service {
	if registry == nil {
		registry = platform.NewRegistry()
	}
	return &Service{store: repository, registry: registry}
}

func (s *Service) Run(ctx context.Context, req RunRequest) (RunResult, error) {
	account, err := s.resolveAccount(req)
	if err != nil {
		return RunResult{}, err
	}
	client, ok := s.registry.Client(account.Platform)
	if !ok {
		return RunResult{}, errors.New("unsupported sync platform")
	}
	now := time.Now().UTC()
	dateFrom := firstNonEmpty(req.DateFrom, now.AddDate(0, 0, -7).Format("2006-01-02"))
	dateTo := firstNonEmpty(req.DateTo, now.Format("2006-01-02"))
	reportTypes := req.ReportTypes
	if len(reportTypes) == 0 {
		reportTypes = platform.DefaultReportTypes(account.Platform)
	}
	job := store.SyncJob{
		ID:         newID(),
		TenantID:   account.TenantID,
		AccountID:  account.ID,
		Platform:   account.Platform,
		ReportType: "multi",
		DateFrom:   dateFrom,
		DateTo:     dateTo,
		Status:     "running",
		Reason:     firstNonEmpty(req.Reason, "manual_sync"),
		CreatedAt:  now,
		StartedAt:  now,
		ResultMeta: map[string]any{"report_types": reportTypes},
	}
	s.store.SaveSyncJob(job)
	result, err := client.Sync(ctx, platform.SyncRequest{
		Account:     account,
		ReportTypes: reportTypes,
		DateFrom:    dateFrom,
		DateTo:      dateTo,
		Now:         now,
	})
	finishedAt := time.Now().UTC()
	job.FinishedAt = finishedAt
	if err != nil {
		job.Status = "failed"
		job.ErrorMeta = map[string]any{"error": err.Error()}
		s.store.SaveSyncJob(job)
		return RunResult{}, err
	}
	s.store.SaveAdEntities(result.Entities)
	s.store.SaveRawReportRows(result.Reports)
	job.Status = "success"
	job.ResultMeta = map[string]any{
		"entity_count": len(result.Entities),
		"report_rows":  len(result.Reports),
		"report_types": reportTypes,
		"client_meta":  result.Meta,
	}
	s.store.SaveSyncJob(job)
	return RunResult{
		Job:          job,
		EntityCount:  len(result.Entities),
		ReportRows:   len(result.Reports),
		ReportTypes:  reportTypes,
		ClientMode:   stringMeta(result.Meta, "client_mode"),
		SdkReference: stringSliceMeta(result.Meta, "sdk_reference"),
	}, nil
}

func (s *Service) resolveAccount(req RunRequest) (store.Account, error) {
	if req.AccountID != "" {
		account, ok := s.store.AccountByID(req.AccountID)
		if !ok {
			return store.Account{}, errors.New("account not found")
		}
		return account, nil
	}
	platformName := strings.ToLower(firstNonEmpty(req.Platform, "douyin"))
	for _, account := range s.store.Accounts() {
		if account.Platform == platformName {
			return account, nil
		}
	}
	account := demoAccount(platformName)
	s.store.SaveAccount(account)
	return account, nil
}

func demoAccount(platformName string) store.Account {
	now := time.Now().UTC()
	return store.Account{
		ID:                stableUUID("demo-account:" + platformName),
		TenantID:          "00000000-0000-4000-8000-000000000001",
		OrganizationID:    "demo-org",
		ChannelID:         "demo-channel-" + platformName,
		Platform:          platformName,
		ExternalAccountID: "demo_" + platformName + "_account",
		AccountName:       platformName + " 演示账户",
		Status:            "active",
		TokenRef:          "kms://demo/" + platformName,
		LastSyncStatus:    "pending",
		CreatedAt:         now,
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func stringMeta(meta map[string]any, key string) string {
	value, _ := meta[key].(string)
	return value
}

func stringSliceMeta(meta map[string]any, key string) []string {
	values, ok := meta[key].([]string)
	if ok {
		return values
	}
	out := []string{}
	if rawValues, ok := meta[key].([]any); ok {
		for _, value := range rawValues {
			if text, ok := value.(string); ok {
				out = append(out, text)
			}
		}
	}
	return out
}

func newID() string {
	return stableUUID(time.Now().UTC().Format(time.RFC3339Nano))
}

func stableUUID(value string) string {
	sum := uint64(1469598103934665603)
	for _, b := range []byte(value) {
		sum ^= uint64(b)
		sum *= 1099511628211
	}
	return "00000000-0000-4000-8" + hex3(sum&0xfff) + "-" + hex12((sum>>12)&0xffffffffffff)
}

func hex3(value uint64) string {
	const digits = "0123456789abcdef"
	return string([]byte{digits[(value>>8)&0xf], digits[(value>>4)&0xf], digits[value&0xf]})
}

func hex12(value uint64) string {
	const digits = "0123456789abcdef"
	buf := make([]byte, 12)
	for i := 11; i >= 0; i-- {
		buf[i] = digits[value&0xf]
		value >>= 4
	}
	return string(buf)
}
