package app

import (
	"fmt"
	"time"

	docengine "github.com/raillen/prumo/internal/documentation"
	"github.com/raillen/prumo/internal/planning"
)

// FeedbackReport is the outcome of running the Living Plan feedback loop for a
// batch of accepted decisions: the impacts they imply, the Documentation Delta
// produced, and the recomputed readiness after applying it.
type FeedbackReport struct {
	Impacts   []docengine.Impact        `json:"impacts"`
	Delta     docengine.Delta           `json:"delta"`
	Readiness docengine.ReadinessReport `json:"readiness"`
}

// LivingPlanLoop implements the Documentation Delta / readiness feedback loop:
//
//	accepted decision → documentation impact → documentation delta →
//	governance/review policy → canonical docs → readiness recompute
//
// M5 (docengine) remains the authority on applicability/coverage/readiness;
// planning only decides which decisions are accepted and which contracts they
// affect. The loop is read-only from the planning perspective: it never writes
// canonical documents itself, only the delta record under the governance flow.
type LivingPlanLoop struct {
	Root string
}

// PlanDelta maps accepted decisions onto documentation impacts and produces a
// proposed Documentation Delta. Only accepted binding decisions affect
// documentation: suggestions, hypotheses and rejected/superseded records are
// excluded so the model cannot self-author canonical docs.
func (l LivingPlanLoop) PlanDelta(goal string, accepted []planning.DecisionProposal, now time.Time) (docengine.Delta, error) {
	bindings, err := docengine.LoadBindings(l.Root)
	if err != nil {
		return docengine.Delta{}, err
	}
	impacts := ImpactsFromDecisions(bindings, accepted)
	delta := docengine.NewDelta(goal, impacts, now)
	if err := delta.Validate(); err != nil {
		return docengine.Delta{}, err
	}
	return delta, nil
}

// ApplyGovernance walks the Documentation Delta through the M5 review policy:
// proposed → reviewed → accepted → applied, recording the decision ids as
// evidence. It fails if the delta cannot legally reach the applied state.
func (l LivingPlanLoop) ApplyGovernance(delta docengine.Delta, evidence []string, now time.Time) (docengine.Delta, error) {
	current := delta
	for _, target := range []string{"reviewed", "accepted", "applied"} {
		next, err := current.Transition(target, evidence, now)
		if err != nil {
			return docengine.Delta{}, err
		}
		current = next
	}
	return current, nil
}

// Run executes the full loop: accept → impact → delta → governance → applied →
// readiness recompute. It returns the proposed delta as built plus the applied
// delta and the readiness after the loop, so callers can persist deliberately
// without the loop writing canonical docs on its own.
func (l LivingPlanLoop) Run(goal string, accepted []planning.DecisionProposal, now time.Time) (FeedbackReport, error) {
	bound, err := docengine.LoadBindings(l.Root)
	if err != nil {
		return FeedbackReport{}, err
	}
	impacts := ImpactsFromDecisions(bound, accepted)
	proposal, err := l.PlanDelta(goal, accepted, now)
	if err != nil {
		return FeedbackReport{}, err
	}
	delta := proposal
	if len(impacts) > 0 {
		applied, err := l.ApplyGovernance(proposal, decisionIDs(accepted), now)
		if err != nil {
			return FeedbackReport{}, err
		}
		delta = applied
	}
	readiness, err := docengine.Readiness(l.Root, goal)
	if err != nil {
		return FeedbackReport{}, err
	}
	return FeedbackReport{Impacts: impacts, Delta: delta, Readiness: readiness}, nil
}

// ImpactsFromDecisions converts accepted decisions into documentation impacts.
// Each decision that affects at least one contract produces one impact per
// contract, carrying the bound documents. Unbound contracts still produce an
// impact so the delta records the gap explicitly.
func ImpactsFromDecisions(bindings []docengine.Binding, accepted []planning.DecisionProposal) []docengine.Impact {
	docsByContract := map[string][]string{}
	for _, b := range bindings {
		docsByContract[b.ContractID] = append([]string{}, b.Sources...)
	}
	seen := map[string]bool{}
	impacts := []docengine.Impact{}
	for _, decision := range accepted {
		if decision.Status != planning.StatusAccepted {
			continue
		}
		for _, contract := range decision.Affected {
			if contract == "" {
				continue
			}
			if seen[contract] {
				continue
			}
			seen[contract] = true
			impacts = append(impacts, docengine.Impact{
				ContractID: contract,
				Documents:  docsByContract[contract],
				Reason:     "living plan decision " + decision.ID,
				Severity:   "medium",
			})
		}
	}
	return impacts
}

func decisionIDs(decisions []planning.DecisionProposal) []string {
	ids := []string{}
	for _, d := range decisions {
		if d.Status == planning.StatusAccepted {
			ids = append(ids, d.ID)
		}
	}
	return ids
}

// ValidateFeedback ensures a loop result is coherent: applied deltas must have
// evidence, and a delta touching contracts must reflect decisions that were
// actually accepted.
func ValidateFeedback(report FeedbackReport, accepted []planning.DecisionProposal) error {
	if err := report.Delta.Validate(); err != nil {
		return err
	}
	if len(accepted) > 0 && len(report.Impacts) == 0 {
		return fmt.Errorf("accepted decisions produced no documentation impact")
	}
	return nil
}
