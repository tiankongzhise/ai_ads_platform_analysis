package store

type StateStore interface {
	SaveState(state OAuthState)
	ConsumeState(stateValue string) (OAuthState, bool)
}

type Repository interface {
	StateStore
	SaveToken(token Token)
	SaveAccount(account Account)
	Accounts() []Account
	AccountByID(accountID string) (Account, bool)
	SaveSyncJob(job SyncJob)
	SyncJobs() []SyncJob
	SaveAdEntities(entities []AdEntity)
	AdEntities(filter AdEntityFilter) []AdEntity
	SaveRawReportRows(rows []RawReportRow)
	RawReportRows(filter RawReportFilter) []RawReportRow
}

var _ Repository = (*MemoryStore)(nil)
