package journal

import "time"

type Kind string

const (
	ReadOnly      Kind = "read-only"
	Reversible    Kind = "reversible"
	SideEffecting Kind = "side-effecting"
	Destructive   Kind = "destructive"
	Unknown       Kind = "unknown"
)

type Entry struct {
	ID             string `json:"id"`
	RunID          string `json:"run_id"`
	Kind           Kind   `json:"kind"`
	Target         string `json:"target"`
	Intent         string `json:"intent"`
	Status         string `json:"status"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
	CreatedAt      string `json:"created_at"`
	Outcome        string `json:"outcome,omitempty"`
}

func New(id, run string, kind Kind, target, intent string) Entry {
	return Entry{ID: id, RunID: run, Kind: kind, Target: target, Intent: intent, Status: "pending", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
}
func (e Entry) Complete(outcome string) Entry { e.Status = "completed"; e.Outcome = outcome; return e }
func (e Entry) UnknownSideEffect() Entry      { e.Status = "unknown_side_effect"; return e }
