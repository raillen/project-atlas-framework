package app

import (
	"sort"

	"github.com/raillen/prumo/internal/contextcompiler"
	docengine "github.com/raillen/prumo/internal/documentation"
	"github.com/raillen/prumo/internal/planning"
)

// DefaultResumeBudget is the default token budget used when no explicit budget
// is given, mirroring the maximum useful context for a planning turn.
const DefaultResumeBudget = 4000

// DocTokenCost is the coarse per-document estimate for compiled context; like
// the planning token costs it is a stable packing unit, not a measurement.
const DocTokenCost = 200

// ResumeReport is the outcome of compiling resume context for a planning
// session: the reconstructed transcript-free session plus the compiled manifest
// with its pressure report.
type ResumeReport struct {
	Session  planning.PlanningSession `json:"session"`
	Manifest contextcompiler.Manifest `json:"manifest"`
}

// ResumeDocSources derives the canonical document sources for a session from
// the docengine bindings that cover its affected contracts/gaps. Each source is
// emitted once per canonical document path and tagged with the binding
// authority; the list is sorted by ref for deterministic packing.
func ResumeDocSources(session planning.PlanningSession, bindings []docengine.Binding) []contextcompiler.Source {
	affected := session.AffectedContracts()
	byContract := map[string]docengine.Binding{}
	for _, b := range bindings {
		byContract[b.ContractID] = b
	}
	seen := map[string]bool{}
	sources := []contextcompiler.Source{}
	for _, contract := range affected {
		binding, ok := byContract[contract]
		if !ok {
			continue
		}
		authority := binding.Authority
		if planning.AuthorityRank(authority) == 0 {
			authority = planning.AuthorityDocumentedEvidence
		}
		for _, doc := range binding.Sources {
			ref := "doc:" + doc
			if seen[ref] {
				continue
			}
			seen[ref] = true
			sources = append(sources, contextcompiler.Source{
				Ref:       ref,
				Authority: authority,
				TokenCost: DocTokenCost,
			})
		}
	}
	sort.Slice(sources, func(i, j int) bool { return sources[i].Ref < sources[j].Ref })
	return sources
}

// CompileResumeContext builds and compiles the resume context manifest for a
// planning session under a token budget. It receives only canonical structured
// state: the session's scope/Goal, affected contracts, accepted decisions, open
// questions, and the last preview, plus the canonical documents that bind those
// contracts. A transcript is never accepted as input. budget <= 0 falls back to
// DefaultResumeBudget.
func CompileResumeContext(session planning.PlanningSession, bindings []docengine.Binding, budget int) ResumeReport {
	if budget <= 0 {
		budget = DefaultResumeBudget
	}
	sources := []contextcompiler.Source{}
	for _, cs := range session.ContextSources() {
		sources = append(sources, contextcompiler.Source{
			Ref:       cs.Ref,
			Authority: cs.Authority,
			TokenCost: cs.TokenCost,
		})
	}
	sources = append(sources, ResumeDocSources(session, bindings)...)
	manifest := contextcompiler.Compile(session.RunID, sources, budget)
	return ResumeReport{Session: session, Manifest: manifest}
}

// Resumable reports whether a session still has a stopping condition artifact:
// the compiled manifest carries a budget, so a session with a healthy or
// pressured manifest can continue and a critical one must not grow silently.
func Resumable(report ResumeReport) bool {
	if report.Manifest.Pressure == contextcompiler.PressureCritical {
		return false
	}
	return HasOpenQuestions(report.Session)
}

// HasOpenQuestions reports whether the session still has questions awaiting
// resolution, so the interview engine knows the run can continue.
func HasOpenQuestions(session planning.PlanningSession) bool {
	for _, q := range session.Open {
		if q.IsOpen() {
			return true
		}
	}
	return false
}

// SessionOpenQuestions returns the highest-impact open questions of a resumed
// session, safe to drive the next interview turn.
func SessionOpenQuestions(session planning.PlanningSession) []planning.OpenQuestion {
	return planning.PriorityOrder(session.Open)
}
