package planning

import "testing"

func dec(id, statement, scope string, affected []string) DecisionProposal {
	return DecisionProposal{
		ID:             id,
		Statement:      statement,
		Classification: ClassificationExplicitDecision,
		Scope:          scope,
		Authority:      AuthorityUserDecision,
		Confidence:     ConfidenceUnknown,
		Status:         StatusAccepted,
		Affected:       affected,
	}
}

func TestDetectContradictionsSameContractDifferentStatement(t *testing.T) {
	proposals := []DecisionProposal{
		dec("DP-1", "use Postgres", "v0.4", []string{"contract:db"}),
		dec("DP-2", "use SQLite", "v0.4", []string{"contract:db"}),
	}
	contradictions := DetectContradictions(proposals)
	if len(contradictions) != 1 {
		t.Fatalf("expected 1 contradiction, got %d", len(contradictions))
	}
	c := contradictions[0]
	if c.Contract != "contract:db" || c.Scope != "v0.4" {
		t.Fatalf("unexpected contradiction: %#v", c)
	}
}

func TestDetectContradictionsDistinctContracts(t *testing.T) {
	proposals := []DecisionProposal{
		dec("DP-1", "use Postgres", "v0.4", []string{"contract:db"}),
		dec("DP-2", "use React", "v0.4", []string{"contract:ui"}),
	}
	if got := DetectContradictions(proposals); len(got) != 0 {
		t.Fatalf("expected no contradictions for distinct contracts, got %#v", got)
	}
}

func TestDetectContradictionsIgnoresSameStatement(t *testing.T) {
	proposals := []DecisionProposal{
		dec("DP-1", "use Postgres", "v0.4", []string{"contract:db"}),
		dec("DP-2", "use Postgres", "v0.4", []string{"contract:db"}),
	}
	if got := DetectContradictions(proposals); len(got) != 0 {
		t.Fatalf("expected no contradiction for identical statements, got %#v", got)
	}
}

func TestDetectContradictionsSkipsNonDecisions(t *testing.T) {
	hypothesis := DecisionProposal{
		ID:             "DP-1",
		Statement:      "maybe Postgres",
		Classification: ClassificationHypothesis,
		Scope:          "v0.4",
		Authority:      AuthorityInferredState,
		Confidence:     ConfidenceLow,
		Status:         StatusProposed,
		Affected:       []string{"contract:db"},
	}
	accepted := dec("DP-2", "use SQLite", "v0.4", []string{"contract:db"})
	if got := DetectContradictions([]DecisionProposal{hypothesis, accepted}); len(got) != 0 {
		t.Fatalf("a hypothesis and a decision must not contradict, got %#v", got)
	}
}

func TestResolveContradictionHigherAuthorityWins(t *testing.T) {
	question := openQ("Q20")
	question.Contract = "contract:db"
	result, err := Resolve(ResolveRequest{
		Question:       question,
		Statement:      "use SQLite",
		Classification: ClassificationExplicitDecision,
		Source:         SourceUser,
		Actor:          "raillen",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	existing := dec("DP-OLD", "use Postgres", "v0.4", []string{"contract:db"})
	existing.Authority = AuthorityDocumentedEvidence

	contradictions := DetectContradictions([]DecisionProposal{existing, result.Proposal})
	if len(contradictions) != 1 {
		t.Fatalf("expected 1 contradiction between existing and new, got %d", len(contradictions))
	}
	resolution, err := ResolveContradiction(contradictions[0], existing, result.Proposal)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resolution.Superseded {
		t.Fatalf("higher-authority new decision should supersede existing, got %#v", resolution)
	}
	if resolution.WinnerID != result.Proposal.ID {
		t.Fatalf("winner should be the user decision %s, got %s", result.Proposal.ID, resolution.WinnerID)
	}
}

func TestResolveContradictionEqualAuthorityNoSilentWin(t *testing.T) {
	a := dec("DP-X", "use A", "v0.4", []string{"contract:io"})
	b := dec("DP-Y", "use B", "v0.4", []string{"contract:io"})
	c := Contradiction{
		Scope:             "v0.4",
		Contract:          "contract:io",
		ExistingID:        "DP-X",
		ExistingStatement: "use A",
		NewID:             "DP-Y",
		NewStatement:      "use B",
	}
	res, err := ResolveContradiction(c, a, b)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Equal || res.Rejected || res.Superseded {
		t.Fatalf("equal authorities must report Equal and never silently win, got %#v", res)
	}
}

func TestResolveContradictionRejectsLowerAuthority(t *testing.T) {
	existing := dec("DP-C", "use NTP", "v0.4", []string{"contract:chrono"})
	proposed := dec("DP-D", "use local clock", "v0.4", []string{"contract:chrono"})
	proposed.Authority = AuthorityInferredState
	proposed.Confidence = ConfidenceLow
	proposed.Status = StatusProposed

	c := Contradiction{
		Scope:             "v0.4",
		Contract:          "contract:chrono",
		ExistingID:        "DP-C",
		ExistingStatement: "use NTP",
		NewID:             "DP-D",
		NewStatement:      "use local clock",
	}
	res, err := ResolveContradiction(c, existing, proposed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Rejected || res.Superseded {
		t.Fatalf("lower-authority proposal must be rejected, got %#v", res)
	}
	if res.WinnerID != existing.ID {
		t.Fatalf("existing decision should stay authoritative, got %s", res.WinnerID)
	}
}

func TestBuildPreviewMandatoryWhenBlockingResolved(t *testing.T) {
	q := openQ("Q30")
	result, err := Resolve(ResolveRequest{
		Question:       q,
		Statement:      "use NTP",
		Classification: ClassificationExplicitDecision,
		Source:         SourceUser,
		Actor:          "raillen",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	preview := BuildPreview([]ResolveResult{result}, nil)
	if preview.BlockersResolved != 1 {
		t.Fatalf("expected 1 blocker resolved, got %d", preview.BlockersResolved)
	}
	if !preview.Mandatory {
		t.Fatal("resolving a blocker must make the preview mandatory")
	}
	if len(preview.Extracted) != 1 {
		t.Fatalf("expected 1 extracted decision, got %d", len(preview.Extracted))
	}
	if preview.OpenRemaining != 0 {
		t.Fatalf("expected 0 open remaining, got %d", preview.OpenRemaining)
	}
}

func TestBuildPreviewKeepsOpenQuestionCounted(t *testing.T) {
	q := openQ("Q31")
	q.Blocking = false
	result, err := Resolve(ResolveRequest{
		Question:       q,
		Statement:      "we can't tell yet",
		Classification: ClassificationUnresolved,
		Source:         SourceUser,
		Actor:          "raillen",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	preview := BuildPreview([]ResolveResult{result}, nil)
	if preview.OpenRemaining != 1 {
		t.Fatalf("expected 1 open remaining, got %d", preview.OpenRemaining)
	}
	if len(preview.Extracted) != 0 {
		t.Fatalf("an unresolved answer must not extract decisions, got %#v", preview.Extracted)
	}
}

func TestBuildPreviewFlagsContradiction(t *testing.T) {
	existing := dec("DP-OLD", "use Postgres for contract:db", "v0.4", []string{"contract:db"})
	q := OpenQuestion{
		ID:       "Q40",
		Scope:    "v0.4",
		Priority: "blocker",
		Blocking: true,
		Status:   "open",
		Contract: "contract:db",
	}
	result, err := Resolve(ResolveRequest{
		Question:       q,
		Statement:      "use SQLite",
		Classification: ClassificationExplicitDecision,
		Source:         SourceUser,
		Actor:          "raillen",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Give the resolved proposal the same affected contract so it pairs with existing.
	result.Proposal.Affected = []string{"contract:db"}

	preview := BuildPreview([]ResolveResult{result}, []DecisionProposal{existing})
	if len(preview.Contradictions) != 1 {
		t.Fatalf("expected 1 contradiction in preview, got %#v", preview.Contradictions)
	}
	if !preview.Mandatory {
		t.Fatal("a contradiction must make the preview mandatory per the specification")
	}
}
