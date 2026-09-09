package experience

import (
	"fmt"
	"strings"
	"time"
)

// HandoffStatus models the lifecycle state of a structured agent/session handoff.
type HandoffStatus string

const (
	HandoffPending      HandoffStatus = "pending"
	HandoffTransferred  HandoffStatus = "transferred"
	HandoffAcknowledged HandoffStatus = "acknowledged"
	HandoffRejected     HandoffStatus = "rejected"
)

func (s HandoffStatus) Valid() bool {
	switch s {
	case HandoffPending, HandoffTransferred, HandoffAcknowledged, HandoffRejected:
		return true
	}
	return false
}

// CreateHandoff constructs and validates an agent-to-agent state handoff.
func CreateHandoff(id, from, to, runID, goalID string, summary Summary, decisions, questions, evidence []string) (Handoff, error) {
	if strings.TrimSpace(id) == "" {
		return Handoff{}, fmt.Errorf("handoff requires an ID")
	}
	if strings.TrimSpace(from) == "" {
		return Handoff{}, fmt.Errorf("handoff requires a sender ('from')")
	}
	if strings.TrimSpace(to) == "" {
		return Handoff{}, fmt.Errorf("handoff requires a recipient ('to')")
	}

	h := Handoff{
		ID:              id,
		From:            from,
		To:              to,
		RunID:           runID,
		GoalID:          goalID,
		Status:          HandoffPending,
		Summary:         summary,
		ActiveDecisions: decisions,
		OpenQuestions:   questions,
		Evidence:        evidence,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339),
	}
	return h, nil
}

// Acknowledge marks the handoff as accepted by the receiving agent.
func (h *Handoff) Acknowledge(actor string) error {
	if h.Status != HandoffPending && h.Status != HandoffTransferred {
		return fmt.Errorf("cannot acknowledge handoff in status %q", h.Status)
	}
	h.Status = HandoffAcknowledged
	h.AcknowledgedAt = time.Now().UTC().Format(time.RFC3339)
	return nil
}

// Reject marks the handoff as rejected by the recipient.
func (h *Handoff) Reject(actor, reason string) error {
	if h.Status != HandoffPending && h.Status != HandoffTransferred {
		return fmt.Errorf("cannot reject handoff in status %q", h.Status)
	}
	h.Status = HandoffRejected
	return nil
}
