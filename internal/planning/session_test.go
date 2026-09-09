package planning

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func sessionFixture() PlanningSession {
	return PlanningSession{
		ID:    "S1",
		RunID: "R1",
		Scope: "v0.4",
		Goal:  "M6 living plan ready",
		Decisions: []DecisionProposal{
			{
				ID:             "DP-D1",
				Statement:      "adopt NTP as canonical clock",
				Classification: ClassificationExplicitDecision,
				Scope:          "v0.4",
				Authority:      AuthorityUserDecision,
				Confidence:     ConfidenceUnknown,
				Status:         StatusAccepted,
				Affected:       []string{"contract:chrono:id"},
			},
		},
		Open: []OpenQuestion{
			openQ("Q1"),
			openQ("Q2"),
		},
		LastPreview: BuildPreview(nil, nil),
		UpdatedAt:   "2026-01-01T00:00:00Z",
	}
}

func TestPlanningSessionValidate(t *testing.T) {
	if err := sessionFixture().Validate(); err != nil {
		t.Fatalf("fixture should validate: %v", err)
	}
	for name, mutate := range map[string]func(*PlanningSession){
		"missing id":    func(s *PlanningSession) { s.ID = "" },
		"missing run":   func(s *PlanningSession) { s.RunID = " " },
		"missing scope": func(s *PlanningSession) { s.Scope = "" },
	} {
		t.Run(name, func(t *testing.T) {
			s := sessionFixture()
			mutate(&s)
			if err := s.Validate(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestCheckpointRoundTrip(t *testing.T) {
	s := sessionFixture()
	cp := s.Checkpoint()
	if cp.Version != SessionCheckpointVersion {
		t.Fatalf("expected version %d, got %d", SessionCheckpointVersion, cp.Version)
	}
	if cp.SessionID != s.ID || cp.RunID != s.RunID || cp.Scope != s.Scope || cp.Goal != s.Goal {
		t.Fatal("checkpoint did not preserve identity")
	}
	if !reflect.DeepEqual(cp.Decisions, s.Decisions) {
		t.Fatal("checkpoint did not preserve decisions")
	}
	if !reflect.DeepEqual(cp.Open, s.Open) {
		t.Fatal("checkpoint did not preserve open questions")
	}
	data, err := json.Marshal(cp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var back SessionCheckpoint
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(back, cp) {
		t.Fatal("checkpoint did not round-trip through JSON")
	}
}

func TestResumeFromCanonicalState(t *testing.T) {
	s := sessionFixture()
	canonical := []DecisionProposal{{
		ID:             "DP-D9",
		Statement:      "canonical documented decision received before resume",
		Classification: ClassificationRequirement,
		Scope:          "v0.4",
		Authority:      AuthorityProjectDecision,
		Confidence:     ConfidenceUnknown,
		Status:         StatusAccepted,
	}}
	open := []OpenQuestion{openQ("Q9")}

	resumed, err := Resume(s.Checkpoint(), canonical, open)
	if err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	if resumed.ID != s.ID || resumed.RunID != s.RunID || resumed.Scope != s.Scope || resumed.Goal != s.Goal {
		t.Fatal("resume must keep identity, scope and goal from the checkpoint")
	}
	if !reflect.DeepEqual(resumed.Decisions, canonical) {
		t.Fatal("resume must rebuild decisions from canonical state, not the checkpoint")
	}
	if !reflect.DeepEqual(resumed.Open, open) {
		t.Fatal("resume must rebuild open questions from canonical state")
	}
	if !reflect.DeepEqual(resumed.LastPreview, s.LastPreview) {
		t.Fatal("resume must keep the last structured planning state from the checkpoint")
	}
}

func TestResumeRejectsUnknownVersion(t *testing.T) {
	cp := sessionFixture().Checkpoint()
	cp.Version = 99
	if _, err := Resume(cp, nil, nil); err == nil {
		t.Fatal("expected error for unsupported checkpoint version")
	}
}

func TestResumeNeverCarriesConversation(t *testing.T) {
	cp := sessionFixture().Checkpoint()
	resumed, err := Resume(cp, nil, nil)
	if err != nil {
		t.Fatalf("resume failed: %v", err)
	}
	data, err := json.Marshal(resumed)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	text := strings.ToLower(string(data))
	for _, forbidden := range []string{"transcript", "rawconversation", "conversation"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("resumed session leaked a %q field", forbidden)
		}
	}
}

func TestWithResultsFoldsAcceptedBatch(t *testing.T) {
	s := sessionFixture()
	result, err := Resolve(ResolveRequest{
		Question:       s.Open[0],
		Statement:      "user keeps NTP as canonical time source",
		Classification: ClassificationExplicitDecision,
		Source:         SourceUser,
		Actor:          "raillen",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	next, err := s.WithResults([]ResolveResult{result}, nil)
	if err != nil {
		t.Fatalf("with results failed: %v", err)
	}
	if err := next.Validate(); err != nil {
		t.Fatalf("folded session must validate: %v", err)
	}
	found := false
	for _, d := range next.Decisions {
		if d.ID == result.Proposal.ID && d.Status == StatusAccepted {
			found = true
		}
	}
	if !found {
		t.Fatalf("accepted decision %q missing after fold", result.Proposal.ID)
	}
	countExisting := 0
	for _, d := range next.Decisions {
		if d.ID == s.Decisions[0].ID {
			countExisting++
		}
	}
	if countExisting != 1 {
		t.Fatalf("existing decision %q must be preserved exactly once, got %d", s.Decisions[0].ID, countExisting)
	}
	for _, q := range next.Open {
		if q.ID == s.Open[0].ID {
			t.Fatal("resolved question must leave the open set")
		}
	}
	if next.LastPreview.BlockersResolved != 1 {
		t.Fatalf("preview should report 1 blocker resolved, got %d", next.LastPreview.BlockersResolved)
	}
	if next.UpdatedAt == s.UpdatedAt {
		t.Fatal("updated_at must be refreshed after a fold")
	}
}

func TestWithResultsDoesNotPromoteSuggestion(t *testing.T) {
	s := sessionFixture()
	suggestion, err := Resolve(ResolveRequest{
		Question:       s.Open[1],
		Statement:      "the model suggests reusing the existing contract",
		Classification: ClassificationAgentSuggestion,
		Source:         SourceModel,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	next, err := s.WithResults([]ResolveResult{suggestion}, nil)
	if err != nil {
		t.Fatalf("with results failed: %v", err)
	}
	for _, d := range next.Decisions {
		if d.ID == suggestion.Proposal.ID {
			t.Fatalf("agent suggestion %q must not be folded as an accepted decision", d.ID)
		}
		if d.Classification == ClassificationAgentSuggestion {
			t.Fatalf("agent suggestion must not be promoted into decisions")
		}
	}
}

func TestAffectedContracts(t *testing.T) {
	s := PlanningSession{
		Decisions: []DecisionProposal{
			{
				ID:         "DP-A",
				Status:     StatusAccepted,
				Authority:  AuthorityProjectDecision,
				Confidence: ConfidenceUnknown,
				Affected:   []string{"contract:z", "contract:a", "contract:a"}, // duplicates collapse
			},
			{ID: "DP-B", Status: StatusProposed, Authority: AuthorityAgentSuggestion, Affected: []string{"contract:ignored"}},
		},
		Open: []OpenQuestion{{ID: "Q1", Contract: "contract:m", Scope: "v0.4", Priority: "blocker", Status: "open"}},
	}
	got := s.AffectedContracts()
	want := []string{"contract:a", "contract:m", "contract:z"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestContextSourcesDeterministicAndComplete(t *testing.T) {
	s := sessionFixture()
	first := s.ContextSources()
	second := s.ContextSources()
	if !reflect.DeepEqual(first, second) {
		t.Fatal("context sources must be deterministic")
	}

	refs := map[string]bool{}
	for _, src := range first {
		if refs[src.Ref] {
			t.Fatalf("duplicate ref %q", src.Ref)
		}
		refs[src.Ref] = true
	}
	for _, want := range []string{"goal:v0.4", "contract:contract:chrono:id", "decision:DP-D1", "question:Q1", "question:Q2", "preview:S1"} {
		if !refs[want] {
			t.Fatalf("missing context ref %q", want)
		}
	}

	for i := 1; i < len(first); i++ {
		prev, cur := first[i-1], first[i]
		prevRank, curRank := AuthorityRank(prev.Authority), AuthorityRank(cur.Authority)
		if prevRank > curRank {
			t.Fatalf("sources out of authority order: %q (%d) before %q (%d)", prev.Ref, prevRank, cur.Ref, curRank)
		}
		if prevRank == curRank && prev.Ref > cur.Ref {
			t.Fatalf("sources out of ref order within authority: %q before %q", cur.Ref, prev.Ref)
		}
	}
}

func TestContextSourcesOrderByRank(t *testing.T) {
	s := sessionFixture()
	sources := s.ContextSources()
	for _, src := range sources {
		if src.TokenCost <= 0 {
			t.Fatalf("ref %q has no token cost", src.Ref)
		}
	}
}
