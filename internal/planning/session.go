package planning

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// SessionCheckpointVersion is the wire version of the persisted planning
// session snapshot. Bump it on incompatible shape changes.
const SessionCheckpointVersion = 1

// Token costs are coarse, stable per-source estimates for context compilation.
// They are packing units, not measurements of actual model tokens; keep the
// values stable so manifests are reproducible across runs.
const (
	TokenCostGoal     = 40
	TokenCostContract = 80
	TokenCostDecision = 120
	TokenCostQuestion = 60
	TokenCostPreview  = 100
)

// SessionSource is one deterministic input to context compilation for a resumed
// planning session. TokenCost is a coarse, stable estimate used to pack sources
// under a budget; it is not a measurement of actual model tokens.
type SessionSource struct {
	Ref       string `json:"ref"`
	Authority string `json:"authority"`
	TokenCost int    `json:"token_cost"`
}

// PlanningSession is the transcript-free structured state of a planning run.
// It holds only what a resumed turn may see: scope/Goal, accepted decisions,
// still-open questions, and the last structured planning state (preview).
//
// Raw conversation and secrets are explicitly never part of the session; the
// checkpoint is built exclusively from canonical structured records.
type PlanningSession struct {
	ID          string             `json:"id"`
	RunID       string             `json:"run_id"`
	Scope       string             `json:"scope"`
	Goal        string             `json:"goal,omitempty"`
	Decisions   []DecisionProposal `json:"decisions,omitempty"`
	Open        []OpenQuestion     `json:"open_questions,omitempty"`
	LastPreview Preview            `json:"last_preview,omitempty"`
	UpdatedAt   string             `json:"updated_at,omitempty"`
}

// NewPlanningSession creates a blank structured session. No transcript is
// attached: the session only carries canonical planning state from the start.
func NewPlanningSession(id, runID, scope, goal string) PlanningSession {
	return PlanningSession{
		ID:        id,
		RunID:     runID,
		Scope:     scope,
		Goal:      goal,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
}

// Validate enforces the invariants of a persisted planning session.
func (s PlanningSession) Validate() error {
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("planning session requires an id")
	}
	if strings.TrimSpace(s.RunID) == "" {
		return fmt.Errorf("planning session requires a run id")
	}
	if strings.TrimSpace(s.Scope) == "" {
		return fmt.Errorf("planning session requires a scope")
	}
	return nil
}

// SessionCheckpoint is the serializable resume snapshot of a PlanningSession.
// It is the durable interchange contract used to continue a run; it never
// contains a raw conversation.
type SessionCheckpoint struct {
	Version   int                `json:"version"`
	SessionID string             `json:"session_id"`
	RunID     string             `json:"run_id"`
	Scope     string             `json:"scope"`
	Goal      string             `json:"goal,omitempty"`
	Decisions []DecisionProposal `json:"decisions,omitempty"`
	Open      []OpenQuestion     `json:"open_questions,omitempty"`
	Preview   Preview            `json:"preview,omitempty"`
	UpdatedAt string             `json:"updated_at,omitempty"`
}

// Checkpoint returns the stable structured snapshot used to resume the session.
func (s PlanningSession) Checkpoint() SessionCheckpoint {
	return SessionCheckpoint{
		Version:   SessionCheckpointVersion,
		SessionID: s.ID,
		RunID:     s.RunID,
		Scope:     s.Scope,
		Goal:      s.Goal,
		Decisions: s.Decisions,
		Open:      s.Open,
		Preview:   s.LastPreview,
		UpdatedAt: s.UpdatedAt,
	}
}

// WithResults folds one resolved batch into the session: accepted decisions are
// kept, resolved questions move out of Open, and the last structured planning
// state is recomputed from the resolver preview. existing holds the accepted
// decisions already canonical before this batch.
func (s PlanningSession) WithResults(results []ResolveResult, existing []DecisionProposal) (PlanningSession, error) {
	next := s
	next.Decisions = dedupeDecisions(append(next.Decisions, existing...))
	resolved := map[string]bool{}
	for _, r := range results {
		if r.Proposal.ID == "" {
			continue
		}
		if r.Proposal.Status == StatusAccepted {
			next.Decisions = dedupeDecisions(append(next.Decisions, r.Proposal))
		}
		if r.Resolved {
			resolved[r.Question.ID] = true
		}
	}
	open := []OpenQuestion{}
	for _, q := range next.Open {
		if !resolved[q.ID] {
			open = append(open, q)
		}
	}
	for _, r := range results {
		if resolved[r.Question.ID] {
			continue
		}
		if r.Question.IsOpen() {
			if !containsQuestion(open, r.Question.ID) {
				open = append(open, r.Question)
			}
		}
	}
	next.Open = open
	next.LastPreview = BuildPreview(results, next.Decisions)
	next.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return next, next.Validate()
}

// Resume restores a planning session from a checkpoint using only canonical
// structured state: the checkpoint (identity, scope/goal, last preview) plus
// the caller-provided canonical accepted decisions and still-open questions.
// It never reads or depends on a raw transcript.
func Resume(cp SessionCheckpoint, canonicalDecisions []DecisionProposal, openQuestions []OpenQuestion) (PlanningSession, error) {
	if cp.Version != SessionCheckpointVersion {
		return PlanningSession{}, fmt.Errorf("unsupported session checkpoint version %d", cp.Version)
	}
	session := PlanningSession{
		ID:          cp.SessionID,
		RunID:       cp.RunID,
		Scope:       cp.Scope,
		Goal:        cp.Goal,
		Decisions:   canonicalDecisions,
		Open:        openQuestions,
		LastPreview: cp.Preview,
		UpdatedAt:   cp.UpdatedAt,
	}
	if err := session.Validate(); err != nil {
		return PlanningSession{}, err
	}
	return session, nil
}

// AffectedContracts returns the sorted, deduplicated union of contracts touched
// by accepted decisions (affected) and still-open questions (contract). Empty
// contracts and non-accepted decision records are ignored.
func (s PlanningSession) AffectedContracts() []string {
	seen := map[string]bool{}
	for _, d := range s.Decisions {
		if d.Status != StatusAccepted {
			continue
		}
		for _, contract := range d.Affected {
			contract = strings.TrimSpace(contract)
			if contract != "" {
				seen[contract] = true
			}
		}
	}
	for _, q := range s.Open {
		contract := strings.TrimSpace(q.Contract)
		if contract != "" {
			seen[contract] = true
		}
	}
	contracts := []string{}
	for contract := range seen {
		contracts = append(contracts, contract)
	}
	sort.Strings(contracts)
	return contracts
}

// ContextSources enumerates, deterministically, every input a resumed turn is
// allowed to see under the Context Compiler integration policy: scope/Goal,
// affected contracts (gaps), accepted decisions, open questions, and the last
// structured planning state. Duplicate refs are collapsed and the list is
// ordered by authority rank and ref so packing is stable across runs.
func (s PlanningSession) ContextSources() []SessionSource {
	sources := []SessionSource{}
	if goal := strings.TrimSpace(s.Goal); goal != "" {
		sources = append(sources, SessionSource{
			Ref:       "goal:" + s.Scope,
			Authority: AuthorityUserDecision,
			TokenCost: TokenCostGoal,
		})
	}
	for _, contract := range s.AffectedContracts() {
		sources = append(sources, SessionSource{
			Ref:       "contract:" + contract,
			Authority: AuthorityDocumentedEvidence,
			TokenCost: TokenCostContract,
		})
	}
	for _, d := range s.Decisions {
		if d.Status != StatusAccepted {
			continue
		}
		authority := d.Authority
		if AuthorityRank(authority) == 0 {
			authority = AuthorityProjectDecision
		}
		sources = append(sources, SessionSource{
			Ref:       "decision:" + d.ID,
			Authority: authority,
			TokenCost: TokenCostDecision,
		})
	}
	for _, q := range s.Open {
		sources = append(sources, SessionSource{
			Ref:       "question:" + q.ID,
			Authority: AuthorityInferredState,
			TokenCost: TokenCostQuestion,
		})
	}
	sources = append(sources, SessionSource{
		Ref:       "preview:" + s.ID,
		Authority: AuthorityInferredState,
		TokenCost: TokenCostPreview,
	})
	return dedupeSortSources(sources)
}

func dedupeSortSources(sources []SessionSource) []SessionSource {
	seen := map[string]bool{}
	out := []SessionSource{}
	for _, s := range sources {
		if seen[s.Ref] {
			continue
		}
		seen[s.Ref] = true
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		ri, rj := AuthorityRank(out[i].Authority), AuthorityRank(out[j].Authority)
		if ri != rj {
			return ri < rj
		}
		return out[i].Ref < out[j].Ref
	})
	return out
}

func dedupeDecisions(decisions []DecisionProposal) []DecisionProposal {
	seen := map[string]bool{}
	out := []DecisionProposal{}
	for _, d := range decisions {
		if d.ID == "" || seen[d.ID] {
			continue
		}
		seen[d.ID] = true
		out = append(out, d)
	}
	return out
}

func containsQuestion(questions []OpenQuestion, id string) bool {
	for _, q := range questions {
		if q.ID == id {
			return true
		}
	}
	return false
}
