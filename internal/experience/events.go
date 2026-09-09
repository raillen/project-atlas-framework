package experience

import (
	"fmt"
	"strings"
	"time"
)

// EventType represents the semantic classification of a structured session event.
type EventType string

const (
	EventSessionStarted   EventType = "session_started"
	EventGoalSelected     EventType = "goal_selected"
	EventDecisionMade     EventType = "decision_made"
	EventCheckpointSaved  EventType = "checkpoint_saved"
	EventToolExecuted     EventType = "tool_executed"
	EventBlockerOccurred  EventType = "blocker_occurred"
	EventSessionCompleted EventType = "session_completed"
	EventHandoffCreated   EventType = "handoff_created"
)

func (t EventType) Valid() bool {
	switch t {
	case EventSessionStarted, EventGoalSelected, EventDecisionMade,
		EventCheckpointSaved, EventToolExecuted, EventBlockerOccurred,
		EventSessionCompleted, EventHandoffCreated:
		return true
	}
	return false
}

// SessionEvent represents a structured, semantic unit of episodic execution history.
// Invariant: Structured events only; conversational transcript or raw prompt dump is prohibited.
type SessionEvent struct {
	ID        string         `json:"id"`
	SessionID string         `json:"session_id"`
	Type      EventType      `json:"type"`
	Summary   string         `json:"summary"`
	Payload   map[string]any `json:"payload,omitempty"`
	Timestamp string         `json:"timestamp"`
}

// Validate verifies invariants for a SessionEvent.
func (e SessionEvent) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return fmt.Errorf("session event requires an ID")
	}
	if strings.TrimSpace(e.SessionID) == "" {
		return fmt.Errorf("session event requires a session_id")
	}
	if !e.Type.Valid() {
		return fmt.Errorf("invalid event type %q", e.Type)
	}
	if strings.TrimSpace(e.Summary) == "" {
		return fmt.Errorf("session event requires a summary")
	}
	return nil
}

// NewEvent creates a timestamped SessionEvent.
func NewEvent(id, sessionID string, eventType EventType, summary string, payload map[string]any) SessionEvent {
	return SessionEvent{
		ID:        id,
		SessionID: sessionID,
		Type:      eventType,
		Summary:   summary,
		Payload:   payload,
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
	}
}
