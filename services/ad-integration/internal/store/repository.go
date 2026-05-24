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
	SaveSyncJob(job SyncJob)
	SyncJobs() []SyncJob
}

var _ Repository = (*MemoryStore)(nil)
