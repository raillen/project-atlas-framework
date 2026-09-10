package experience

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSessionEvents(t *testing.T) {
	ev := NewEvent("ev-001", "sess-01", EventSessionStarted, "Session started for Goal P00-G01", nil)
	if err := ev.Validate(); err != nil {
		t.Fatalf("expected valid event: %v", err)
	}

	// Invariant: Invalid type rejected
	badEv := NewEvent("ev-002", "sess-01", "invalid_type", "Bad event", nil)
	if err := badEv.Validate(); err == nil {
		t.Errorf("expected error validating event with invalid type")
	}

	// Invariant: Missing ID or summary rejected
	if err := (SessionEvent{SessionID: "sess-01", Type: EventGoalSelected}).Validate(); err == nil {
		t.Errorf("expected error on missing ID and summary")
	}
}

func TestSynthesizeSummaries(t *testing.T) {
	events := []SessionEvent{
		NewEvent("e1", "sess-01", EventSessionStarted, "Started work", nil),
		NewEvent("e2", "sess-01", EventDecisionMade, "Use Go standard library", nil),
		NewEvent("e3", "sess-01", EventBlockerOccurred, "Missing database credentials", nil),
		NewEvent("e4", "sess-01", EventToolExecuted, "Ran tests", map[string]any{"status": "success"}),
		NewEvent("e5", "sess-01", EventSessionCompleted, "Completed core module", nil),
	}

	sessSummary := SynthesizeSessionSummary("sess-01", "run-100", events)
	if len(sessSummary.Completed) != 2 { // Completed module + Decision
		t.Errorf("expected 2 completed items, got %d", len(sessSummary.Completed))
	}
	if len(sessSummary.Blockers) != 1 || sessSummary.Blockers[0] != "Missing database credentials" {
		t.Errorf("unexpected blockers: %v", sessSummary.Blockers)
	}
	if len(sessSummary.Current) != 1 || sessSummary.Current[0] != "Ran tests" {
		t.Errorf("unexpected current items: %v", sessSummary.Current)
	}

	// Goal summary synthesis
	goalSummary := SynthesizeGoalSummary("G-01", []Summary{sessSummary})
	if goalSummary.Progress != "blocked" {
		t.Errorf("expected progress 'blocked' due to open blockers, got %s", goalSummary.Progress)
	}
	if len(goalSummary.OpenBlockers) != 1 {
		t.Errorf("expected 1 open blocker in goal summary")
	}
}

func TestHandoffProtocol(t *testing.T) {
	summary := Summary{
		SessionID: "sess-planner",
		Completed: []string{"Architecture defined", "Schemas created"},
		Current:   []string{"Drafting engine implementation"},
		NextSteps: []string{"Implement Go engine", "Write tests"},
	}

	handoff, err := CreateHandoff(
		"ho-001",
		"planner-agent",
		"developer-agent",
		"run-001",
		"P00-G01",
		summary,
		[]string{"dec-clean-code"},
		[]string{"q-perf-budget"},
		[]string{"test-evidence-hash"},
	)
	if err != nil {
		t.Fatalf("failed creating handoff: %v", err)
	}

	if handoff.Status != HandoffPending {
		t.Errorf("expected pending status, got %s", handoff.Status)
	}
	if len(handoff.ActiveDecisions) != 1 || handoff.ActiveDecisions[0] != "dec-clean-code" {
		t.Errorf("expected active decisions in handoff")
	}
	if len(handoff.OpenQuestions) != 1 || handoff.OpenQuestions[0] != "q-perf-budget" {
		t.Errorf("expected open questions in handoff")
	}

	// Acknowledge
	err = handoff.Acknowledge("developer-agent")
	if err != nil {
		t.Fatalf("failed acknowledging handoff: %v", err)
	}
	if handoff.Status != HandoffAcknowledged {
		t.Errorf("expected acknowledged status, got %s", handoff.Status)
	}
	if handoff.AcknowledgedAt == "" {
		t.Errorf("expected non-empty acknowledged_at timestamp")
	}
}

func TestExperienceProposals(t *testing.T) {
	prop, err := ProposeExperience(
		"ep-001",
		"prefer-table-driven-tests",
		"Observed consistent bug reductions when using Go table tests",
		"Add rule to Go coding standards",
		"sess-01",
	)
	if err != nil {
		t.Fatalf("failed creating proposal: %v", err)
	}

	if prop.Status != ProposalStatusProposed {
		t.Errorf("expected proposed status, got %s", prop.Status)
	}
	if !prop.ReviewRequired {
		t.Errorf("expected review_required=true")
	}

	// Approve proposal
	err = prop.Approve("human-operator", "Verified and accepted for codebase")
	if err != nil {
		t.Fatalf("failed to approve proposal: %v", err)
	}
	if prop.Status != ProposalStatusApproved || prop.ApprovedBy != "human-operator" {
		t.Errorf("expected approved proposal state")
	}
}

func TestFileProviderAndRetention(t *testing.T) {
	tmpDir := t.TempDir()
	storeDir := filepath.Join(tmpDir, ".prumo", "experience")

	prov, err := NewFileProvider(storeDir)
	if err != nil {
		t.Fatalf("NewFileProvider failed: %v", err)
	}

	// 1. Record event
	ev := NewEvent("ev-100", "sess-100", EventSessionStarted, "Session 100 started", nil)
	if err := prov.RecordEvent(ev); err != nil {
		t.Fatalf("RecordEvent failed: %v", err)
	}

	events, err := prov.GetEvents("sess-100")
	if err != nil || len(events) != 1 {
		t.Fatalf("GetEvents failed: %v, count=%d", err, len(events))
	}

	// 2. Summary
	s := Summary{
		SessionID: "sess-100",
		Completed: []string{"Task done"},
	}
	if err := prov.SaveSummary(s); err != nil {
		t.Fatalf("SaveSummary failed: %v", err)
	}
	loadedS, err := prov.GetSummary("sess-100")
	if err != nil || loadedS.SessionID != "sess-100" {
		t.Fatalf("GetSummary failed: %v", err)
	}

	// 3. Goal summary
	gs := GoalSummary{
		GoalID:   "G-M9",
		Progress: "completed",
	}
	if err := prov.SaveGoalSummary(gs); err != nil {
		t.Fatalf("SaveGoalSummary failed: %v", err)
	}
	loadedGS, err := prov.GetGoalSummary("G-M9")
	if err != nil || loadedGS.GoalID != "G-M9" {
		t.Fatalf("GetGoalSummary failed: %v", err)
	}

	// 4. Handoff
	h := Handoff{
		ID:      "ho-100",
		From:    "agent-a",
		To:      "agent-b",
		Summary: s,
		Status:  HandoffPending,
	}
	if err := prov.CreateHandoff(h); err != nil {
		t.Fatalf("CreateHandoff failed: %v", err)
	}
	loadedH, err := prov.GetHandoff("ho-100")
	if err != nil || loadedH.From != "agent-a" {
		t.Fatalf("GetHandoff failed: %v", err)
	}

	// 5. Proposal
	p, _ := ProposeExperience("ep-100", "pattern", "obs", "act", "sess-100")
	if err := prov.RecordProposal(p); err != nil {
		t.Fatalf("RecordProposal failed: %v", err)
	}
	props, err := prov.GetProposals()
	if err != nil || len(props) != 1 {
		t.Fatalf("GetProposals failed: %v, count=%d", err, len(props))
	}

	// 6. Retention Pruning
	policy := RetentionPolicy{
		MaxAgeHours:         1,
		MaxEventsPerSession: 10,
		PruneRawEvents:      true,
	}

	// Event from 2 hours ago
	oldEv := SessionEvent{
		ID:        "ev-old",
		SessionID: "sess-old",
		Type:      EventSessionStarted,
		Summary:   "Old session",
		Timestamp: time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339),
	}
	_ = prov.RecordEvent(oldEv)

	pruned, err := prov.Prune(policy)
	if err != nil {
		t.Fatalf("Prune failed: %v", err)
	}
	if pruned != 1 {
		t.Errorf("expected 1 pruned event, got %d", pruned)
	}
}
