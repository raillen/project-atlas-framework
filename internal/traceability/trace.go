package traceability

type Edge struct {
	ID       string   `json:"id"`
	From     string   `json:"from"`
	To       string   `json:"to"`
	Kind     string   `json:"kind"`
	Evidence []string `json:"evidence,omitempty"`
}
type JournalEntry struct {
	ID          string   `json:"id"`
	Goal        string   `json:"goal,omitempty"`
	Title       string   `json:"title,omitempty"`
	Summary     string   `json:"summary"`
	Decisions   []string `json:"decisions,omitempty"`
	CodeChanges []string `json:"code_changes,omitempty"`
	Tests       []string `json:"tests,omitempty"`
	Evidence    []string `json:"evidence,omitempty"`
	ChangedDocs []string `json:"changed_docs,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
}
