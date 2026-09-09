package adoption

import (
	"testing"
)

func TestUncertaintyQuestionsGeneration(t *testing.T) {
	ledger := ConfidenceLedger{
		Version: 1,
		Entries: []LedgerEntry{
			{
				ID: "cld-0001", Claim: "Candidate project archetype is cli", Category: CategoryClassification,
				Confidence: ConfidenceMedium, Score: 0.60, Evidence: []string{"fact-1"},
				Alternatives: []string{"library", "web-application"}, Status: LedgerStatusUnresolved,
				Authority: "inferred-state", RequiresConfirmation: true,
			},
			{
				ID: "cld-0002", Claim: "Proposed capability cli", Category: CategoryCapability,
				Confidence: ConfidenceLow, Score: 0.30, Evidence: []string{"fact-2"},
				Status: LedgerStatusUnresolved, Authority: "inferred-state", RequiresConfirmation: true,
			},
			{
				ID: "cld-0003", Claim: "Proposed capability go", Category: CategoryCapability,
				Confidence: ConfidenceFactual, Score: 1.0, Evidence: []string{"fact-3"},
				Status: LedgerStatusUnresolved, Authority: "inferred-state", RequiresConfirmation: false,
			},
		},
	}
	ledger.RecomputeSummary()

	questions := GenerateQuestions(ledger)
	if len(questions) != 2 {
		t.Fatalf("expected 2 questions for unconfirmed entries, got %d", len(questions))
	}

	// First question should have options including alternatives
	q1 := questions[0]
	if q1.LedgerEntryID != "cld-0001" {
		t.Errorf("expected question for cld-0001, got %s", q1.LedgerEntryID)
	}
	if len(q1.Options) < 3 {
		t.Errorf("expected options with alternatives, got %v", q1.Options)
	}

	// Test ApplyChoice accept
	err := ApplyChoice(&ledger, q1, ResolutionChoice{
		QuestionID:     q1.ID,
		SelectedOption: "accept",
		Actor:          "tester",
		Rationale:      "Confirmed CLI archetype",
	})
	if err != nil {
		t.Fatalf("ApplyChoice failed: %v", err)
	}

	entry1 := ledger.Entries[0]
	if entry1.Status != LedgerStatusAccepted {
		t.Errorf("expected accepted status, got %s", entry1.Status)
	}
	if entry1.RequiresConfirmation {
		t.Errorf("expected RequiresConfirmation to be false after confirmation")
	}
	if entry1.Authority != "user-decision" {
		t.Errorf("expected authority 'user-decision', got %s", entry1.Authority)
	}

	// Test ApplyChoice reject on question 2
	q2 := questions[1]
	err = ApplyChoice(&ledger, q2, ResolutionChoice{
		QuestionID:     q2.ID,
		SelectedOption: "no",
		Actor:          "tester",
	})
	if err != nil {
		t.Fatalf("ApplyChoice reject failed: %v", err)
	}
	entry2 := ledger.Entries[1]
	if entry2.Status != LedgerStatusRejected {
		t.Errorf("expected rejected status, got %s", entry2.Status)
	}
}

func TestResolveNonInteractive(t *testing.T) {
	ledger := ConfidenceLedger{
		Version: 1,
		Entries: []LedgerEntry{
			{
				ID: "cld-0001", Claim: "Go language", Category: CategoryClassification,
				Confidence: ConfidenceFactual, Score: 1.0, Evidence: []string{"fact-1"},
				Status: LedgerStatusUnresolved, Authority: "inferred-state", RequiresConfirmation: true,
			},
			{
				ID: "cld-0002", Claim: "May be library", Category: CategoryClassification,
				Confidence: ConfidenceLow, Score: 0.30, Evidence: []string{"fact-2"},
				Status: LedgerStatusUnresolved, Authority: "inferred-state", RequiresConfirmation: true,
			},
		},
	}
	ledger.RecomputeSummary()

	session := ResolveNonInteractive(&ledger)
	if session.ResolvedCount != 1 {
		t.Errorf("expected 1 resolved count for factual entry, got %d", session.ResolvedCount)
	}
	if session.UnresolvedCount != 1 {
		t.Errorf("expected 1 unresolved count for low confidence entry, got %d", session.UnresolvedCount)
	}

	// First entry should be accepted
	if ledger.Entries[0].Status != LedgerStatusAccepted {
		t.Errorf("expected entry 1 to be accepted, got %s", ledger.Entries[0].Status)
	}
	// Second entry must remain unresolved
	if ledger.Entries[1].Status != LedgerStatusUnresolved {
		t.Errorf("expected low confidence entry 2 to remain unresolved, got %s", ledger.Entries[1].Status)
	}
}
