package adoption

import (
	"fmt"
	"sort"
	"strings"
)

// UncertaintyQuestion represents an adoption uncertainty derived from the Confidence Ledger,
// formatted for operator interview or automated policy resolution.
type UncertaintyQuestion struct {
	ID            string         `json:"id"`
	LedgerEntryID string         `json:"ledger_entry_id"`
	Question      string         `json:"question"`
	Context       string         `json:"context,omitempty"`
	Options       []string       `json:"options"`
	DefaultChoice string         `json:"default_choice"`
	Confidence    FactConfidence `json:"confidence"`
}

// ResolutionChoice captures an operator or policy answer resolving an UncertaintyQuestion.
type ResolutionChoice struct {
	QuestionID     string `json:"question_id"`
	SelectedOption string `json:"selected_option"`
	Actor          string `json:"actor"` // e.g. "human", "policy-non-interactive"
	Rationale      string `json:"rationale,omitempty"`
}

// ResolutionSession tracks questions and progress in an uncertainty resolution pass.
type ResolutionSession struct {
	Version         int                   `json:"version"`
	Questions       []UncertaintyQuestion `json:"questions"`
	ResolvedCount   int                   `json:"resolved_count"`
	UnresolvedCount int                   `json:"unresolved_count"`
}

// GenerateQuestions transforms all ConfidenceLedger entries requiring confirmation
// into structured interview questions.
func GenerateQuestions(ledger ConfidenceLedger) []UncertaintyQuestion {
	unconfirmed := ledger.RequiringConfirmation()
	questions := make([]UncertaintyQuestion, 0, len(unconfirmed))

	for i, entry := range unconfirmed {
		qID := fmt.Sprintf("adopt-q-%04d", i+1)
		var qText string
		var options []string
		var defaultChoice string

		switch entry.Category {
		case CategoryClassification:
			qText = fmt.Sprintf("Confirm inferred characteristic: %s?", entry.Claim)
			options = []string{"accept", "reject"}
			if len(entry.Alternatives) > 0 {
				options = append(options, entry.Alternatives...)
			}
			defaultChoice = "accept"

		case CategoryCapability:
			qText = fmt.Sprintf("Enable proposed capability: %s?", entry.Claim)
			options = []string{"yes", "no"}
			defaultChoice = "yes"

		case CategoryProfile:
			qText = fmt.Sprintf("Adopt candidate profile: %s?", entry.Claim)
			options = []string{"accept", "reject"}
			if len(entry.Alternatives) > 0 {
				options = append(options, entry.Alternatives...)
			}
			defaultChoice = "accept"

		case CategoryMapping:
			qText = fmt.Sprintf("Accept documentation binding: %s?", entry.Claim)
			options = []string{"accept", "reject"}
			defaultChoice = "accept"

		default:
			qText = fmt.Sprintf("Review inference: %s", entry.Claim)
			options = []string{"accept", "reject"}
			defaultChoice = "accept"
		}

		questions = append(questions, UncertaintyQuestion{
			ID:            qID,
			LedgerEntryID: entry.ID,
			Question:      qText,
			Context:       entry.Rationale,
			Options:       options,
			DefaultChoice: defaultChoice,
			Confidence:    entry.Confidence,
		})
	}

	sort.Slice(questions, func(i, j int) bool {
		return questions[i].ID < questions[j].ID
	})

	return questions
}

// ApplyChoice applies a resolution choice to the Confidence Ledger,
// upgrading authority to "user-decision" and transitioning the status.
func ApplyChoice(ledger *ConfidenceLedger, q UncertaintyQuestion, choice ResolutionChoice) error {
	var targetEntry *LedgerEntry
	for i := range ledger.Entries {
		if ledger.Entries[i].ID == q.LedgerEntryID {
			targetEntry = &ledger.Entries[i]
			break
		}
	}

	if targetEntry == nil {
		return fmt.Errorf("ledger entry %q not found for question %q", q.LedgerEntryID, q.ID)
	}

	lowerChoice := strings.ToLower(strings.TrimSpace(choice.SelectedOption))

	switch lowerChoice {
	case "accept", "yes":
		targetEntry.Status = LedgerStatusAccepted
		targetEntry.RequiresConfirmation = false
		targetEntry.Authority = "user-decision"
		if choice.Rationale != "" {
			targetEntry.Rationale = fmt.Sprintf("%s (Confirmed by %s: %s)", targetEntry.Rationale, choice.Actor, choice.Rationale)
		} else {
			targetEntry.Rationale = fmt.Sprintf("%s (Confirmed by %s)", targetEntry.Rationale, choice.Actor)
		}
	case "reject", "no":
		targetEntry.Status = LedgerStatusRejected
		targetEntry.RequiresConfirmation = false
		targetEntry.Authority = "user-decision"
		if choice.Rationale != "" {
			targetEntry.Rationale = fmt.Sprintf("%s (Rejected by %s: %s)", targetEntry.Rationale, choice.Actor, choice.Rationale)
		} else {
			targetEntry.Rationale = fmt.Sprintf("%s (Rejected by %s)", targetEntry.Rationale, choice.Actor)
		}
	default:
		// Operator selected an alternative interpretation
		targetEntry.Status = LedgerStatusAccepted
		targetEntry.RequiresConfirmation = false
		targetEntry.Authority = "user-decision"
		targetEntry.Claim = fmt.Sprintf("Alternative selected: %s", choice.SelectedOption)
		targetEntry.Rationale = fmt.Sprintf("Operator selected alternative %q (by %s)", choice.SelectedOption, choice.Actor)
	}

	ledger.RecomputeSummary()
	return nil
}

// ResolveNonInteractive automatically resolves all high and factual confidence entries,
// leaving low/medium confidence entries unresolved (to uphold the canonical rule that
// low-confidence inferences are never silently promoted to canonical).
func ResolveNonInteractive(ledger *ConfidenceLedger) ResolutionSession {
	questions := GenerateQuestions(*ledger)
	resolved := 0
	unresolved := 0

	for _, q := range questions {
		// Only auto-accept if factual or high confidence
		if q.Confidence == ConfidenceFactual || q.Confidence == ConfidenceHigh {
			choice := ResolutionChoice{
				QuestionID:     q.ID,
				SelectedOption: q.DefaultChoice,
				Actor:          "policy-non-interactive",
				Rationale:      "Auto-accepted high/factual confidence inference in non-interactive mode",
			}
			_ = ApplyChoice(ledger, q, choice)
			resolved++
		} else {
			unresolved++
		}
	}

	return ResolutionSession{
		Version:         1,
		Questions:       questions,
		ResolvedCount:   resolved,
		UnresolvedCount: unresolved,
	}
}
