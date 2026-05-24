package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db    *sql.DB
	state StateStore
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &PostgresStore{db: db, state: NewMemoryStore()}, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) SaveState(state OAuthState) {
	s.state.SaveState(state)
}

func (s *PostgresStore) ConsumeState(stateValue string) (OAuthState, bool) {
	return s.state.ConsumeState(stateValue)
}

func (s *PostgresStore) SaveToken(token Token) {
	_, _ = s.db.ExecContext(context.Background(), `
		INSERT INTO ad_sync.ad_oauth_tokens (
			id, tenant_id, platform, external_account_id, token_ref,
			access_token_ciphertext, refresh_token_ciphertext,
			access_token_expires_at, refresh_token_expires_at, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
		ON CONFLICT (tenant_id, platform, external_account_id) DO UPDATE SET
			token_ref = EXCLUDED.token_ref,
			access_token_ciphertext = EXCLUDED.access_token_ciphertext,
			refresh_token_ciphertext = EXCLUDED.refresh_token_ciphertext,
			access_token_expires_at = EXCLUDED.access_token_expires_at,
			refresh_token_expires_at = EXCLUDED.refresh_token_expires_at,
			updated_at = now()
	`, token.ID, token.TenantID, token.Platform, token.ExternalAccountID, token.TokenRef, token.AccessTokenCiphertext, token.RefreshTokenCiphertext, token.AccessTokenExpiresAt, token.RefreshTokenExpiresAt, token.CreatedAt)
}

func (s *PostgresStore) SaveAccount(account Account) {
	_, _ = s.db.ExecContext(context.Background(), `
		INSERT INTO ad_sync.ad_platform_accounts (
			id, tenant_id, organization_id, channel_id, platform, external_account_id,
			account_name, status, token_ref, last_sync_status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now())
		ON CONFLICT (tenant_id, platform, external_account_id) DO UPDATE SET
			organization_id = EXCLUDED.organization_id,
			channel_id = EXCLUDED.channel_id,
			account_name = EXCLUDED.account_name,
			status = EXCLUDED.status,
			token_ref = EXCLUDED.token_ref,
			last_sync_status = EXCLUDED.last_sync_status,
			last_sync_at = CASE WHEN EXCLUDED.last_sync_status IN ('success', 'failed') THEN now() ELSE ad_sync.ad_platform_accounts.last_sync_at END,
			updated_at = now()
	`, account.ID, account.TenantID, nullableUUID(account.OrganizationID), nullableUUID(account.ChannelID), account.Platform, account.ExternalAccountID, account.AccountName, account.Status, account.TokenRef, account.LastSyncStatus, account.CreatedAt)
}

func (s *PostgresStore) Accounts() []Account {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id::text, tenant_id::text, COALESCE(organization_id::text, ''), COALESCE(channel_id::text, ''),
			platform, external_account_id, COALESCE(account_name, ''), status, COALESCE(token_ref, ''),
			COALESCE(last_sync_status, ''), created_at
		FROM ad_sync.ad_platform_accounts
		ORDER BY created_at DESC
	`)
	if err != nil {
		return []Account{}
	}
	defer rows.Close()
	accounts := []Account{}
	for rows.Next() {
		var account Account
		if err := rows.Scan(&account.ID, &account.TenantID, &account.OrganizationID, &account.ChannelID, &account.Platform, &account.ExternalAccountID, &account.AccountName, &account.Status, &account.TokenRef, &account.LastSyncStatus, &account.CreatedAt); err == nil {
			accounts = append(accounts, account)
		}
	}
	return accounts
}

func (s *PostgresStore) AccountByID(accountID string) (Account, bool) {
	var account Account
	err := s.db.QueryRowContext(context.Background(), `
		SELECT id::text, tenant_id::text, COALESCE(organization_id::text, ''), COALESCE(channel_id::text, ''),
			platform, external_account_id, COALESCE(account_name, ''), status, COALESCE(token_ref, ''),
			COALESCE(last_sync_status, ''), created_at
		FROM ad_sync.ad_platform_accounts
		WHERE id = $1
	`, accountID).Scan(&account.ID, &account.TenantID, &account.OrganizationID, &account.ChannelID, &account.Platform, &account.ExternalAccountID, &account.AccountName, &account.Status, &account.TokenRef, &account.LastSyncStatus, &account.CreatedAt)
	return account, err == nil
}

func (s *PostgresStore) SaveSyncJob(job SyncJob) {
	resultMeta := marshalJSON(job.ResultMeta)
	errorMeta := marshalJSON(job.ErrorMeta)
	_, _ = s.db.ExecContext(context.Background(), `
		INSERT INTO ad_sync.ad_sync_jobs (
			id, tenant_id, account_id, platform, report_type, date_from, date_to,
			status, result_meta, error_meta, started_at, finished_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10::jsonb, $11, $12, $13)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			result_meta = EXCLUDED.result_meta,
			error_meta = EXCLUDED.error_meta,
			started_at = COALESCE(EXCLUDED.started_at, ad_sync.ad_sync_jobs.started_at),
			finished_at = EXCLUDED.finished_at
	`, job.ID, job.TenantID, job.AccountID, job.Platform, job.ReportType, job.DateFrom, job.DateTo, job.Status, resultMeta, errorMeta, zeroTimeToNil(job.StartedAt), zeroTimeToNil(job.FinishedAt), job.CreatedAt)
}

func (s *PostgresStore) SyncJobs() []SyncJob {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id::text, tenant_id::text, account_id::text, platform, report_type, date_from::text, date_to::text,
			status, COALESCE(result_meta, '{}'::jsonb), COALESCE(error_meta, '{}'::jsonb), started_at, finished_at, created_at
		FROM ad_sync.ad_sync_jobs
		ORDER BY created_at DESC
		LIMIT 100
	`)
	if err != nil {
		return []SyncJob{}
	}
	defer rows.Close()
	jobs := []SyncJob{}
	for rows.Next() {
		var job SyncJob
		var resultBytes []byte
		var errorBytes []byte
		var startedAt sql.NullTime
		var finishedAt sql.NullTime
		if err := rows.Scan(&job.ID, &job.TenantID, &job.AccountID, &job.Platform, &job.ReportType, &job.DateFrom, &job.DateTo, &job.Status, &resultBytes, &errorBytes, &startedAt, &finishedAt, &job.CreatedAt); err != nil {
			continue
		}
		job.ResultMeta = unmarshalMap(resultBytes)
		job.ErrorMeta = unmarshalMap(errorBytes)
		if startedAt.Valid {
			job.StartedAt = startedAt.Time
		}
		if finishedAt.Valid {
			job.FinishedAt = finishedAt.Time
		}
		jobs = append(jobs, job)
	}
	return jobs
}

func (s *PostgresStore) SaveAdEntities(entities []AdEntity) {
	for _, entity := range entities {
		_, _ = s.db.ExecContext(context.Background(), `
			INSERT INTO ad_sync.ad_entities (
				id, tenant_id, account_id, platform, entity_type, external_id,
				parent_external_id, name, status, raw, synced_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''), $8, $9, $10::jsonb, $11, now())
			ON CONFLICT (account_id, platform, entity_type, external_id) DO UPDATE SET
				parent_external_id = EXCLUDED.parent_external_id,
				name = EXCLUDED.name,
				status = EXCLUDED.status,
				raw = EXCLUDED.raw,
				synced_at = EXCLUDED.synced_at,
				updated_at = now()
		`, entity.ID, entity.TenantID, entity.AccountID, entity.Platform, entity.EntityType, entity.ExternalID, entity.ParentExternalID, entity.Name, entity.Status, marshalJSON(entity.Raw), entity.SyncedAt)
	}
}

func (s *PostgresStore) AdEntities(filter AdEntityFilter) []AdEntity {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id::text, tenant_id::text, account_id::text, platform, entity_type, external_id,
			COALESCE(parent_external_id, ''), COALESCE(name, ''), COALESCE(status, ''), raw, synced_at
		FROM ad_sync.ad_entities
		WHERE ($1 = '' OR account_id = $1::uuid)
			AND ($2 = '' OR platform = $2)
			AND ($3 = '' OR entity_type = $3)
		ORDER BY entity_type ASC, external_id ASC
		LIMIT $4
	`, filter.AccountID, filter.Platform, filter.EntityType, positiveLimit(filter.Limit))
	if err != nil {
		return []AdEntity{}
	}
	defer rows.Close()
	entities := []AdEntity{}
	for rows.Next() {
		var entity AdEntity
		var rawBytes []byte
		if err := rows.Scan(&entity.ID, &entity.TenantID, &entity.AccountID, &entity.Platform, &entity.EntityType, &entity.ExternalID, &entity.ParentExternalID, &entity.Name, &entity.Status, &rawBytes, &entity.SyncedAt); err == nil {
			entity.Raw = unmarshalMap(rawBytes)
			entities = append(entities, entity)
		}
	}
	return entities
}

func (s *PostgresStore) SaveRawReportRows(rows []RawReportRow) {
	for _, row := range rows {
		_, _ = s.db.ExecContext(context.Background(), `
			INSERT INTO ad_sync.ad_raw_report_rows (
				id, tenant_id, account_id, platform, report_type, granularity, stat_date, stat_hour,
				entity_type, external_entity_id, metrics, raw, synced_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12::jsonb, $13)
			ON CONFLICT (account_id, report_type, stat_date, stat_hour, external_entity_id) DO UPDATE SET
				metrics = EXCLUDED.metrics,
				raw = EXCLUDED.raw,
				synced_at = EXCLUDED.synced_at
		`, row.ID, row.TenantID, row.AccountID, row.Platform, row.ReportType, row.Granularity, row.StatDate, row.StatHour, row.EntityType, row.ExternalEntityID, marshalJSON(row.Metrics), marshalJSON(row.Raw), row.SyncedAt)
	}
}

func (s *PostgresStore) RawReportRows(filter RawReportFilter) []RawReportRow {
	rows, err := s.db.QueryContext(context.Background(), `
		SELECT id::text, tenant_id::text, account_id::text, platform, report_type, granularity,
			stat_date::text, stat_hour, entity_type, external_entity_id, metrics, raw, synced_at
		FROM ad_sync.ad_raw_report_rows
		WHERE ($1 = '' OR account_id = $1::uuid)
			AND ($2 = '' OR platform = $2)
			AND ($3 = '' OR report_type = $3)
		ORDER BY stat_date DESC, report_type ASC, external_entity_id ASC
		LIMIT $4
	`, filter.AccountID, filter.Platform, filter.ReportType, positiveLimit(filter.Limit))
	if err != nil {
		return []RawReportRow{}
	}
	defer rows.Close()
	reportRows := []RawReportRow{}
	for rows.Next() {
		var row RawReportRow
		var metricsBytes []byte
		var rawBytes []byte
		if err := rows.Scan(&row.ID, &row.TenantID, &row.AccountID, &row.Platform, &row.ReportType, &row.Granularity, &row.StatDate, &row.StatHour, &row.EntityType, &row.ExternalEntityID, &metricsBytes, &rawBytes, &row.SyncedAt); err != nil {
			continue
		}
		_ = json.Unmarshal(metricsBytes, &row.Metrics)
		row.Raw = unmarshalMap(rawBytes)
		reportRows = append(reportRows, row)
	}
	return reportRows
}

func marshalJSON(value any) string {
	if value == nil {
		return "{}"
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func unmarshalMap(bytes []byte) map[string]any {
	out := map[string]any{}
	_ = json.Unmarshal(bytes, &out)
	return out
}

func zeroTimeToNil(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func positiveLimit(limit int) int {
	if limit <= 0 {
		return 100
	}
	if limit > 1000 {
		return 1000
	}
	return limit
}

func nullableUUID(value string) any {
	value = strings.TrimSpace(value)
	if !isUUIDLike(value) {
		return nil
	}
	return value
}

func isUUIDLike(value string) bool {
	if len(value) != 32 && len(value) != 36 {
		return false
	}
	for _, char := range value {
		if char == '-' {
			continue
		}
		if char >= '0' && char <= '9' {
			continue
		}
		if char >= 'a' && char <= 'f' {
			continue
		}
		if char >= 'A' && char <= 'F' {
			continue
		}
		return false
	}
	return true
}

var _ Repository = (*PostgresStore)(nil)
