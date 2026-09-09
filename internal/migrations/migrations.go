package migrations

type Contract struct {
	ID                string   `json:"id"`
	FromVersion       string   `json:"from_version"`
	ToVersion         string   `json:"to_version"`
	Preconditions     []string `json:"preconditions"`
	BackupStrategy    string   `json:"backup_strategy"`
	AffectedArtifacts []string `json:"affected_artifacts"`
	Reversible        bool     `json:"reversible"`
	Rollback          string   `json:"rollback"`
}
type JournalEntry struct {
	MigrationID string `json:"migration_id"`
	AppliedAt   string `json:"applied_at"`
	Result      string `json:"result"`
	Hash        string `json:"hash"`
}
