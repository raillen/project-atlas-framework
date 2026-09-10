package app

import (
	"reflect"
	"testing"

	"github.com/raillen/prumo/internal/contextcompiler"
	docengine "github.com/raillen/prumo/internal/documentation"
	"github.com/raillen/prumo/internal/planning"
)

func resumeSession(t *testing.T) planning.PlanningSession {
	t.Helper()
	return planning.PlanningSession{
		ID:    "RS1",
		RunID: "RUN-9",
		Scope: "v0.4",
		Goal:  "M6 living plan ready",
		Decisions: []planning.DecisionProposal{
			{
				ID:             "DP-D1",
				Statement:      "NTP is canonical time source",
				Classification: planning.ClassificationExplicitDecision,
				Scope:          "v0.4",
				Authority:      planning.AuthorityUserDecision,
				Confidence:     planning.ConfidenceUnknown,
				Status:         planning.StatusAccepted,
				Affected:       []string{"contract:chrono:id"},
			},
		},
		Open: []planning.OpenQuestion{{
			ID:       "Q2",
			Scope:    "v0.4",
			Contract: "contract:audit:id",
			Priority: "high-risk",
			Status:   "open",
			Question: "where should audit logs live?",
		}},
		LastPreview: planning.BuildPreview(nil, nil),
	}
}

func resumeBindings() []docengine.Binding {
	return []docengine.Binding{
		{
			ContractID: "contract:chrono:id",
			Sources:    []string{"docs/architecture/overview.md"},
			Ownership:  "core",
			Authority:  planning.AuthorityDocumentedEvidence,
		},
		{
			ContractID: "contract:audit:id",
			Sources:    []string{"docs/security/trust-model.md"},
			Authority:  planning.AuthorityInvariant,
		},
	}
}

func TestCompileResumeContextBudgetRespected(t *testing.T) {
	report := CompileResumeContext(resumeSession(t), resumeBindings(), 500)
	if report.Manifest.EstimatedTokens > 500 {
		t.Fatalf("estimated %d exceeds budget 500", report.Manifest.EstimatedTokens)
	}
	used := 0
	for _, s := range report.Manifest.Sources {
		if s.Included {
			used += s.TokenCost
		}
	}
	if used != report.Manifest.EstimatedTokens {
		t.Fatalf("included sources sum %d != manifest estimate %d", used, report.Manifest.EstimatedTokens)
	}
	if report.Manifest.RunID != "RUN-9" {
		t.Fatalf("unexpected run id %q", report.Manifest.RunID)
	}
}

func TestCompileResumeContextDeterministic(t *testing.T) {
	first := CompileResumeContext(resumeSession(t), resumeBindings(), 0)
	second := CompileResumeContext(resumeSession(t), resumeBindings(), 0)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("compiled resume context must be deterministic")
	}
	if first.Manifest.EstimatedTokens == 0 {
		t.Fatal("expected non-empty manifest")
	}
}

func TestCompileResumeContextDedupesDocRefs(t *testing.T) {
	bindings := append(resumeBindings(), docengine.Binding{
		ContractID: "contract:other",
		Sources:    []string{"docs/architecture/overview.md"},
		Authority:  planning.AuthorityDocumentedEvidence,
	})
	report := CompileResumeContext(resumeSession(t), bindings, 5000)
	seen := map[string]bool{}
	for _, s := range report.Manifest.Sources {
		if s.Included && seen[s.Ref] {
			t.Fatalf("duplicate included source %q", s.Ref)
		}
		seen[s.Ref] = true
	}
}

func TestCompileResumeContextPressure(t *testing.T) {
	// Budget equal to a single goal source: the pack is critical because it is
	// fully used, and nothing can be added without an explicit increase.
	report := CompileResumeContext(resumeSession(t), nil, planning.TokenCostGoal)
	if report.Manifest.Pressure != contextcompiler.PressureCritical {
		t.Fatalf("expected critical pressure, got %q", report.Manifest.Pressure)
	}
	if Resumable(report) {
		t.Fatal("critical pressure must stop the session")
	}
}

func TestResumeDocSourcesFromBindings(t *testing.T) {
	got := ResumeDocSources(resumeSession(t), resumeBindings())
	want := []contextcompiler.Source{
		{Ref: "doc:docs/architecture/overview.md", Authority: "documented-evidence", TokenCost: DocTokenCost},
		{Ref: "doc:docs/security/trust-model.md", Authority: "invariant", TokenCost: DocTokenCost},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %#v, got %#v", want, got)
	}
}

func TestResumableAndOpenQuestions(t *testing.T) {
	report := CompileResumeContext(resumeSession(t), resumeBindings(), 500)
	if report.Manifest.Pressure != contextcompiler.PressureHealthy && report.Manifest.Pressure != contextcompiler.PressurePressure {
		t.Fatalf("expected healthy/pressure manifest, got %q", report.Manifest.Pressure)
	}
	if !Resumable(report) {
		t.Fatal("session with open questions and budget left must be resumable")
	}
	isolated := report
	isolated.Session.Open = nil
	if Resumable(isolated) {
		t.Fatal("session without open questions must not be resumable")
	}

	questions := SessionOpenQuestions(resumeSession(t))
	if len(questions) != 1 || questions[0].ID != "Q2" {
		t.Fatalf("expected Q2 as next open question, got %#v", questions)
	}
}
