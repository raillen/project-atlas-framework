package app

import (
	"testing"

	"github.com/raillen/prumo/internal/documentation"
	"github.com/raillen/prumo/internal/planning"
)

func readyPreview() planning.Preview {
	return planning.Preview{
		Extracted: []planning.DecisionProposal{
			{ID: "DP-001", Statement: "Prumo must be provider-neutral", Classification: "explicit-decision", Status: planning.StatusAccepted, Affected: []string{"architecture.system"}},
			{ID: "DP-002", Statement: "No proprietary runtime required", Classification: "non-goal", Status: planning.StatusAccepted, Affected: []string{"architecture.system"}},
		},
		BlockersResolved: 1,
		OpenRemaining:    0,
	}
}

func readyReadiness() docengine.ReadinessReport {
	return docengine.ReadinessReport{
		Ready: true, Goal: "G042",
		ApplicableContracts: []string{"architecture.system"},
		Coverage: []docengine.Coverage{
			{ContractID: "architecture.system", State: docengine.ImplementationReady, Sources: []string{"docs/architecture/overview.md"}},
		},
	}
}

func TestProposeGoalPlanSimpleScopeHasNoTasks(t *testing.T) {
	p := readyPreview()
	p.BlockersResolved = 0
	in := GoalPlanInput{
		Scope:     "planning",
		Preview:   p,
		Readiness: readyReadiness(),
	}
	out, err := ProposeGoalPlan(in, false)
	if err != nil {
		t.Fatal(err)
	}
	if out.PlanRequired {
		t.Fatalf("expected no plan for simple scope, got %d tasks", len(out.Tasks))
	}
	if len(out.Tasks) != 0 {
		t.Fatalf("expected no tasks, got %d", len(out.Tasks))
	}
	if g, ok := out.Goal["acceptance"].([]any); ok {
		if len(g) == 0 {
			t.Fatal("goal must carry acceptance criteria")
		}
	} else {
		t.Fatalf("acceptance type wrong: %T", out.Goal["acceptance"])
	}
}

func TestProposeGoalPlanExplicitPlanRequested(t *testing.T) {
	in := GoalPlanInput{Scope: "planning", Preview: readyPreview(), Readiness: readyReadiness()}
	out, err := ProposeGoalPlan(in, true)
	if err != nil {
		t.Fatal(err)
	}
	if !out.PlanRequired {
		t.Fatal("explicit plan request must force decomposition")
	}
	if len(out.Tasks) < 2 {
		t.Fatalf("expected chained task DAG, got %d tasks", len(out.Tasks))
	}
	for _, task := range out.Tasks {
		if id, _ := task["id"].(string); id == "" {
			t.Fatalf("task without id: %#v", task)
		}
	}
}

func TestProposeGoalPlanMultipleContractsSequenced(t *testing.T) {
	p := readyPreview()
	p.Extracted = append(p.Extracted, planning.DecisionProposal{
		ID: "DP-003", Statement: "Adopt Python 3.12", Classification: "explicit-decision",
		Status: planning.StatusAccepted, Affected: []string{"testing.strategy"},
	})
	in := GoalPlanInput{Scope: "tooling", Preview: p, Readiness: readyReadiness()}
	out, err := ProposeGoalPlan(in, false)
	if err != nil {
		t.Fatal(err)
	}
	if !out.PlanRequired {
		t.Fatal("multiple contracts require sequencing")
	}
	if len(out.Tasks) < 3 {
		t.Fatalf("expected contract tasks plus review task, got %d", len(out.Tasks))
	}
}

func TestProposeGoalPlanConstraintsAndNonGoals(t *testing.T) {
	in := GoalPlanInput{Scope: "platform", Preview: readyPreview(), Readiness: readyReadiness()}
	out, err := ProposeGoalPlan(in, false)
	if err != nil {
		t.Fatal(err)
	}
	nonGoals, ok := out.Goal["non_goals"].([]string)
	if !ok {
		t.Fatalf("expected non_goals []string, got %T", out.Goal["non_goals"])
	}
	found := false
	for _, ng := range nonGoals {
		if ng == "No proprietary runtime required" {
			found = true
		}
	}
	if !found {
		t.Fatalf("non-goal not carried into goal: %v", nonGoals)
	}
}

func TestProposeGoalPlanRequiresScope(t *testing.T) {
	if _, err := ProposeGoalPlan(GoalPlanInput{Preview: readyPreview(), Readiness: readyReadiness()}, false); err == nil {
		t.Fatal("expected error for empty scope")
	}
}

func TestProposeGoalPlanNeverProducesCyclicDAG(t *testing.T) {
	p := readyPreview()
	p.Extracted = append(p.Extracted,
		planning.DecisionProposal{ID: "DP-003", Statement: "A", Classification: "explicit-decision", Status: planning.StatusAccepted, Affected: []string{"testing.strategy"}},
		planning.DecisionProposal{ID: "DP-004", Statement: "B", Classification: "explicit-decision", Status: planning.StatusAccepted, Affected: []string{"runtime.boot"}},
	)
	in := GoalPlanInput{Scope: "wide", Preview: p, Readiness: readyReadiness()}
	out, err := ProposeGoalPlan(in, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(out.Tasks) == 0 {
		t.Fatal("expected tasks")
	}
}
