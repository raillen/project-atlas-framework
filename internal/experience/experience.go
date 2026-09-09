package experience

type Summary struct {
	SessionID string   `json:"session_id"`
	RunID     string   `json:"run_id,omitempty"`
	Completed []string `json:"completed"`
	Current   []string `json:"current"`
	Blockers  []string `json:"blockers"`
	NextSteps []string `json:"next_steps"`
}
type Handoff struct {
	ID      string   `json:"id"`
	From    string   `json:"from"`
	To      string   `json:"to"`
	RunID   string   `json:"run_id"`
	Summary Summary  `json:"summary"`
	Claims  []string `json:"claims,omitempty"`
}
