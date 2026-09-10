package planning

import (
	"path/filepath"
	"testing"

	"github.com/raillen/prumo/internal/validation"
)

func TestClassificationValidAndDecision(t *testing.T) {
	for _, c := range []AnswerClassification{
		ClassificationExplicitDecision, ClassificationPreference, ClassificationConstraint,
		ClassificationRequirement, ClassificationNonGoal, ClassificationUnresolved,
		ClassificationHypothesis, ClassificationAgentSuggestion,
	} {
		if !c.Valid() {
			t.Fatalf("%s should be valid", c)
		}
	}
	decision := []AnswerClassification{
		ClassificationExplicitDecision, ClassificationPreference, ClassificationConstraint,
		ClassificationRequirement, ClassificationNonGoal,
	}
	for _, c := range decision {
		if !c.IsDecision() {
			t.Fatalf("%s should be a decision", c)
		}
	}
	for _, c := range []AnswerClassification{ClassificationHypothesis, ClassificationAgentSuggestion, ClassificationUnresolved} {
		if c.IsDecision() {
			t.Fatalf("%s should not be a decision", c)
		}
	}
}

func TestConfidenceValid(t *testing.T) {
	for _, c := range []Confidence{ConfidenceHigh, ConfidenceMedium, ConfidenceLow, ConfidenceUnknown} {
		if !c.Valid() {
			t.Fatalf("%s should be valid confidence", c)
		}
	}
	if (Confidence("opaque-num")).Valid() {
		t.Fatal("opaque confidence should be rejected")
	}
}

func TestAuthorityRankOrder(t *testing.T) {
	order := []struct {
		authority string
		want      int
	}{
		{AuthorityInvariant, 1},
		{AuthorityUserDecision, 2},
		{AuthorityProjectDecision, 3},
		{AuthorityDocumentedEvidence, 4},
		{AuthorityInferredState, 5},
		{AuthorityAgentSuggestion, 6},
		{AuthorityExternalContent, 7},
		{"unknown", 0},
	}
	for _, tc := range order {
		if got := AuthorityRank(tc.authority); got != tc.want {
			t.Fatalf("AuthorityRank(%s) = %d, want %d", tc.authority, got, tc.want)
		}
	}
}

func TestDecisionProposalValidate(t *testing.T) {
	valid := DecisionProposal{
		ID: "DP-1", Statement: "Control Plane owns canonical state",
		Classification: ClassificationExplicitDecision, Authority: AuthorityUserDecision,
		Confidence: ConfidenceUnknown, Status: StatusAccepted,
	}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid proposal rejected: %v", err)
	}
	if err := (DecisionProposal{Statement: "x", Classification: ClassificationPreference, Authority: AuthorityProjectDecision, Confidence: ConfidenceUnknown, Status: StatusAccepted}).Validate(); err == nil {
		t.Fatal("expected missing id to fail")
	}
	if err := (DecisionProposal{ID: "DP-2", Statement: "x", Classification: "opinion", Authority: AuthorityUserDecision, Confidence: ConfidenceUnknown, Status: StatusAccepted}).Validate(); err == nil {
		t.Fatal("expected unknown classification to fail")
	}
	if err := (DecisionProposal{ID: "DP-3", Statement: "x", Classification: ClassificationRequirement, Authority: AuthorityUserDecision, Confidence: "definitely", Status: StatusAccepted}).Validate(); err == nil {
		t.Fatal("expected unknown confidence to fail")
	}
	if err := (DecisionProposal{ID: "DP-4", Statement: "x", Classification: ClassificationExplicitDecision, Authority: AuthorityUserDecision, Confidence: ConfidenceUnknown, Status: "merged"}).Validate(); err == nil {
		t.Fatal("expected unknown status to fail")
	}
}

func TestAgentSuggestionCannotBeSilentlyPromoted(t *testing.T) {
	suggestion := AgentSuggestion("S-001", "use SQLite for local state", "goal:G042")
	assertDecision := func(d DecisionProposal) {
		t.Helper()
		if d.ID != "S-001" ||
			d.Classification != ClassificationAgentSuggestion ||
			d.Authority != AuthorityAgentSuggestion ||
			d.Confidence != ConfidenceUnknown ||
			d.Status != StatusProposed ||
			d.Statement != "use SQLite for local state" ||
			d.Scope != "goal:G042" {
			t.Fatalf("unexpected factory output: %#v", d)
		}
	}
	assertDecision(suggestion)
	if err := suggestion.Validate(); err != nil {
		t.Fatalf("valid suggestion rejected: %v", err)
	}
	accepted := suggestion
	accepted.Status = StatusAccepted
	if err := accepted.Validate(); err == nil {
		t.Fatal("agent suggestion must not be accepted without an explicit user decision")
	}
	wrongAuthority := suggestion
	wrongAuthority.Authority = AuthorityUserDecision
	if err := wrongAuthority.Validate(); err == nil {
		t.Fatal("agent suggestion must carry agent-suggestion authority")
	}
}

func TestResolvePromotesToUserDecision(t *testing.T) {
	resolved := AgentSuggestion("S-002", "use SQLite", "goal:G042").Resolve("user")
	if err := resolved.Validate(); err != nil {
		t.Fatalf("resolved proposal should be valid: %v", err)
	}
	if resolved.Status != StatusAccepted || resolved.Authority != AuthorityUserDecision || !resolved.Classification.IsDecision() || resolved.Actor != "user" {
		t.Fatalf("resolve did not promote correctly: %#v", resolved)
	}
}

func TestDecisionProposalConformanceFixture(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repoRoot, "conformance", "planning", "valid_decision_proposal.json")
	if errors := validation.ValidateFile(path, "decision-proposal.schema.json", filepath.Join(repoRoot, "schemas")); len(errors) != 0 {
		t.Fatalf("decision proposal fixture failed schema validation: %v", errors)
	}
}
