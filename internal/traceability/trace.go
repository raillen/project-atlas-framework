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
	Summary     string   `json:"summary"`
	Evidence    []string `json:"evidence,omitempty"`
	ChangedDocs []string `json:"changed_docs,omitempty"`
}
