package experience

import (
	"time"
)

// GoalSummary captures high-level execution synthesis across all sessions addressing a Goal.
type GoalSummary struct {
	GoalID         string   `json:"goal_id"`
	Sessions       []string `json:"sessions"`
	Progress       string   `json:"progress"`
	CompletedItems []string `json:"completed_items"`
	OpenBlockers   []string `json:"open_blockers"`
	Decisions      []string `json:"decisions,omitempty"`
	Evidence       []string `json:"evidence,omitempty"`
	UpdatedAt      string   `json:"updated_at"`
}

// SynthesizeSessionSummary aggregates a stream of session events into a concise session summary.
func SynthesizeSessionSummary(sessionID, runID string, events []SessionEvent) Summary {
	summary := Summary{
		SessionID: sessionID,
		RunID:     runID,
		Completed: make([]string, 0),
		Current:   make([]string, 0),
		Blockers:  make([]string, 0),
		NextSteps: make([]string, 0),
	}

	for _, e := range events {
		switch e.Type {
		case EventSessionCompleted:
			summary.Completed = append(summary.Completed, e.Summary)
		case EventBlockerOccurred:
			summary.Blockers = append(summary.Blockers, e.Summary)
		case EventDecisionMade:
			summary.Completed = append(summary.Completed, "Decision: "+e.Summary)
		case EventToolExecuted:
			// Only record completed actions
			if status, ok := e.Payload["status"].(string); ok && status == "success" {
				summary.Current = append(summary.Current, e.Summary)
			}
		}
	}

	return summary
}

// SynthesizeGoalSummary aggregates session summaries into a unified GoalSummary.
func SynthesizeGoalSummary(goalID string, sessions []Summary) GoalSummary {
	gs := GoalSummary{
		GoalID:         goalID,
		Sessions:       make([]string, 0, len(sessions)),
		CompletedItems: make([]string, 0),
		OpenBlockers:   make([]string, 0),
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
	}

	blockerSet := make(map[string]bool)
	completedSet := make(map[string]bool)

	for _, s := range sessions {
		gs.Sessions = append(gs.Sessions, s.SessionID)
		for _, c := range s.Completed {
			if !completedSet[c] {
				completedSet[c] = true
				gs.CompletedItems = append(gs.CompletedItems, c)
			}
		}
		for _, b := range s.Blockers {
			if !blockerSet[b] {
				blockerSet[b] = true
				gs.OpenBlockers = append(gs.OpenBlockers, b)
			}
		}
	}

	if len(gs.OpenBlockers) > 0 {
		gs.Progress = "blocked"
	} else if len(gs.CompletedItems) > 0 {
		gs.Progress = "in_progress"
	} else {
		gs.Progress = "planned"
	}

	return gs
}
