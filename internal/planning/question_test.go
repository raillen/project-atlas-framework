package planning

import (
	"path/filepath"
	"testing"

	"github.com/raillen/prumo/internal/validation"
)

func TestPriorityParseString(t *testing.T) {
	tests := map[Priority]string{
		PriorityBlocker:         "blocker",
		PriorityIrreversible:    "irreversible",
		PriorityHighRisk:        "high-risk",
		PriorityProduct:         "product",
		PriorityImportantDetail: "important-detail",
		PriorityNiceToHave:      "nice-to-have",
	}
	for priority, want := range tests {
		if got := priority.String(); got != want {
			t.Fatalf("String(%d) = %q, want %q", priority, got, want)
		}
		parsed, err := ParsePriority(want)
		if err != nil {
			t.Fatalf("ParsePriority(%q): %v", want, err)
		}
		if parsed != priority {
			t.Fatalf("ParsePriority(%q) = %d, want %d", want, parsed, priority)
		}
	}
	if _, err := ParsePriority("curiosity"); err == nil {
		t.Fatal("expected unknown priority to fail")
	}
}

func TestOpenQuestionValidate(t *testing.T) {
	valid := OpenQuestion{ID: "OQ-1", Scope: "goal:G042", Priority: "blocker", Blocking: true, Status: "open", Question: "What owns canonical state?"}
	if err := valid.Validate(); err != nil {
		t.Fatalf("valid question rejected: %v", err)
	}
	if err := (OpenQuestion{Scope: "goal:G042", Priority: "blocker", Status: "open"}).Validate(); err == nil {
		t.Fatal("expected missing id to fail")
	}
	if err := (OpenQuestion{ID: "OQ-2", Scope: "goal:G042", Priority: "curiosity", Status: "open"}).Validate(); err == nil {
		t.Fatal("expected unknown priority to fail")
	}
	if err := (OpenQuestion{ID: "OQ-3", Scope: "goal:G042", Priority: "blocker", Status: "resolved", Blocking: true}).Validate(); err == nil {
		t.Fatal("expected blocking resolved question without evidence/decision to fail")
	}
	if err := (OpenQuestion{ID: "OQ-4", Scope: "goal:G042", Priority: "high-risk", Status: "resolved", Answer: "control-plane"}).Validate(); err != nil {
		t.Fatalf("resolved question with answer rejected: %v", err)
	}
}

func TestPriorityOrderIsStableAndImpactRanked(t *testing.T) {
	qs := []OpenQuestion{
		{ID: "OQ-a", Scope: "goal:G042", Priority: "nice-to-have", Status: "open"},
		{ID: "OQ-b", Scope: "goal:G042", Priority: "blocker", Status: "open"},
		{ID: "OQ-c", Scope: "goal:G042", Priority: "high-risk", Status: "open"},
		{ID: "OQ-d", Scope: "goal:G042", Priority: "irreversible", Status: "open"},
		{ID: "OQ-e", Scope: "goal:G042", Priority: "product", Status: "open"},
		{ID: "OQ-f", Scope: "goal:G042", Priority: "important-detail", Status: "open"},
	}
	ordered := PriorityOrder(qs)
	wantIDs := []string{"OQ-b", "OQ-d", "OQ-c", "OQ-e", "OQ-f", "OQ-a"}
	for i, want := range wantIDs {
		if ordered[i].ID != want {
			t.Fatalf("position %d = %s, want %s; got %#v", i, ordered[i].ID, want, ordered)
		}
	}
}

func TestPriorityOrderKeepsInputOrderWithinSamePriority(t *testing.T) {
	first := OpenQuestion{ID: "OQ-x", Scope: "goal:G042", Priority: "blocker", Status: "open"}
	second := OpenQuestion{ID: "OQ-y", Scope: "goal:G042", Priority: "blocker", Status: "open"}
	ordered := PriorityOrder([]OpenQuestion{second, first})
	if ordered[0].ID != "OQ-y" || ordered[1].ID != "OQ-x" {
		t.Fatalf("stable ordering broken: %#v", ordered)
	}
}

func TestNextOpenAndBlockingCount(t *testing.T) {
	qs := []OpenQuestion{
		{ID: "OQ-1", Scope: "goal:G042", Priority: "blocker", Blocking: true, Status: "resolved", Answer: "x"},
		{ID: "OQ-2", Scope: "goal:G042", Priority: "high-risk", Blocking: true, Status: "open"},
		{ID: "OQ-3", Scope: "goal:G042", Priority: "nice-to-have", Blocking: false, Status: "open"},
	}
	next, ok := NextOpen(qs)
	if !ok || next.ID != "OQ-2" {
		t.Fatalf("expected OQ-2 as next open, got %#v ok=%v", next, ok)
	}
	if count := BlockingOpenCount(qs); count != 1 {
		t.Fatalf("expected 1 open blocker, got %d", count)
	}
}

func TestOpenQuestionConformanceFixture(t *testing.T) {
	repoRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repoRoot, "conformance", "planning", "valid_open_question.json")
	if errors := validation.ValidateFile(path, "open-question.schema.json", filepath.Join(repoRoot, "schemas")); len(errors) != 0 {
		t.Fatalf("open question fixture failed schema validation: %v", errors)
	}
}
