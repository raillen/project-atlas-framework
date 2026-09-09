package planning

import "testing"

func openQ(id string) OpenQuestion {
	return OpenQuestion{
		ID:              id,
		Scope:           "v0.4",
		Priority:        "high-risk",
		Blocking:        true,
		Status:          "open",
		Contract:        "contract:chrono:id",
		Question:        "which time source is canonical?",
		SuggestedAnswer: []string{"clock", "NTP"},
	}
}

func TestResolveAcceptsUserDecision(t *testing.T) {
	result, err := Resolve(ResolveRequest{
		Question:       openQ("Q1"),
		Statement:      "user selected NTP as canonical clock",
		Classification: ClassificationExplicitDecision,
		Source:         SourceUser,
		Actor:          "raillen",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Resolved {
		t.Fatal("explicit user decision should resolve the question")
	}
	if result.Proposal.Authority != AuthorityUserDecision {
		t.Fatalf("expected user-decision authority, got %q", result.Proposal.Authority)
	}
	if result.Proposal.Confidence != ConfidenceUnknown {
		t.Fatalf("explicit decisions must not carry confidence, got %q", result.Proposal.Confidence)
	}
	if result.Proposal.Status != StatusAccepted {
		t.Fatalf("expected accepted status, got %q", result.Proposal.Status)
	}
	if result.BlockedNow != true {
		t.Fatal("resolving a blocking question should unblock")
	}
	if result.Question.Status != "resolved" || result.Question.DecisionRef == "" {
		t.Fatalf("question should be resolved with a decision ref: %#v", result.Question)
	}
}

func TestResolveDocumentedConstraint(t *testing.T) {
	result, err := Resolve(ResolveRequest{
		Question:       openQ("Q2"),
		Statement:      "repository docs mandate Go for core protocol",
		Classification: ClassificationConstraint,
		Source:         SourceRepo,
		Evidence:       []string{"docs/architecture/dependency-rules.md"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Resolved {
		t.Fatal("documented constraint should resolve the question")
	}
	if result.Proposal.Authority != AuthorityProjectDecision {
		t.Fatalf("expected project-decision authority from repository source, got %q", result.Proposal.Authority)
	}
	if len(result.Proposal.Evidence) != 1 {
		t.Fatalf("evidence should be carried over: %#v", result.Proposal.Evidence)
	}
}

func TestResolveHypothesisStaysProposed(t *testing.T) {
	result, err := Resolve(ResolveRequest{
		Question:       openQ("Q3"),
		Statement:      "model guesses mTLS on local sockets is sufficient",
		Classification: ClassificationHypothesis,
		Source:         SourceModel,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Resolved {
		t.Fatal("a hypothesis must not resolve the question")
	}
	if result.Proposal.Authority != AuthorityInferredState {
		t.Fatalf("expected inferred-state authority, got %q", result.Proposal.Authority)
	}
	if result.Proposal.Confidence != ConfidenceUnknown {
		t.Fatalf("hypothesis without evidence must be unknown confidence, got %q", result.Proposal.Confidence)
	}
	if result.Proposal.Status != StatusProposed {
		t.Fatalf("hypothesis must stay proposed, got %q", result.Proposal.Status)
	}
	if result.Question.Status != "open" {
		t.Fatalf("question must remain open, got %q", result.Question.Status)
	}
}

func TestResolveSuggestionNeverPromoted(t *testing.T) {
	result, err := Resolve(ResolveRequest{
		Question:       openQ("Q4"),
		Statement:      "agent suggests wiring a local event bus",
		Classification: ClassificationAgentSuggestion,
		Source:         SourceModel,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Resolved {
		t.Fatal("a suggestion must not resolve the question")
	}
	if result.Proposal.Authority != AuthorityAgentSuggestion {
		t.Fatalf("expected agent-suggestion authority, got %q", result.Proposal.Authority)
	}
	if result.Proposal.Status != StatusProposed {
		t.Fatalf("suggestion must stay proposed, got %q", result.Proposal.Status)
	}
	if err := result.Proposal.Validate(); err != nil {
		t.Fatalf("suggestion proposal must be valid: %v", err)
	}
}

func TestResolveUnresolvedKeepsOpen(t *testing.T) {
	result, err := Resolve(ResolveRequest{
		Question:       openQ("Q5"),
		Statement:      "cannot determine yet",
		Classification: ClassificationUnresolved,
		Source:         SourceUser,
		Actor:          "raillen",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Resolved {
		t.Fatal("unresolved answer must leave the question open")
	}
	if result.Proposal.ID != "" {
		t.Fatalf("unresolved answer must not create a proposal, got %#v", result.Proposal)
	}
}

func TestResolveRejectsDecisionWithoutSource(t *testing.T) {
	_, err := Resolve(ResolveRequest{
		Question:       openQ("Q6"),
		Statement:      "some decision",
		Classification: ClassificationExplicitDecision,
	})
	if err == nil {
		t.Fatal("explicit decision without user or repository source must fail")
	}
}

func TestResolveRejectsUserDecisionWithoutActor(t *testing.T) {
	_, err := Resolve(ResolveRequest{
		Question:       openQ("Q7"),
		Statement:      "some decision",
		Classification: ClassificationExplicitDecision,
		Source:         SourceUser,
	})
	if err == nil {
		t.Fatal("user decision without actor must fail")
	}
}

func TestResolveRejectsModelDecision(t *testing.T) {
	_, err := Resolve(ResolveRequest{
		Question:       openQ("Q8"),
		Statement:      "model self-decides",
		Classification: ClassificationExplicitDecision,
		Source:         SourceModel,
		Actor:          "assistant",
	})
	if err == nil {
		t.Fatal("model cannot originate an explicit decision")
	}
}

func TestResolveValidatesQuestion(t *testing.T) {
	q := openQ("Q9")
	q.Status = "bogus"
	_, err := Resolve(ResolveRequest{
		Question:       q,
		Statement:      "x",
		Classification: ClassificationExplicitDecision,
		Source:         SourceUser,
		Actor:          "raillen",
	})
	if err == nil {
		t.Fatal("invalid question must fail resolution")
	}
}

func TestResolveManyPairsQuestionsAndRequests(t *testing.T) {
	_, err := ResolveMany([]OpenQuestion{openQ("Q10"), openQ("Q11")}, []ResolveRequest{{
		Question: openQ("Q10"),
	}})
	if err == nil {
		t.Fatal("mismatched pairs must fail")
	}
}
