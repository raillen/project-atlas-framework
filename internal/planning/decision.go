package planning

import (
	"fmt"
	"strings"
)

// DecisionStatus models the lifecycle of a decision proposal or record.
type DecisionStatus string

const (
	StatusProposed   DecisionStatus = "proposed"
	StatusAccepted   DecisionStatus = "accepted"
	StatusRejected   DecisionStatus = "rejected"
	StatusSuperseded DecisionStatus = "superseded"
)

func (s DecisionStatus) Valid() bool {
	switch s {
	case StatusProposed, StatusAccepted, StatusRejected, StatusSuperseded:
		return true
	}
	return false
}

// DecisionProposal is the canonical decision/proposal model for the Living
// Plan. It distinguishes statement, classification, scope, actor/source,
// authority, confidence, status, affected contracts/docs, rationale summary,
// alternatives (rejected candidates), and evidence/refs.
//
// If Rationale is populated it must be an explicit summary; hidden
// chain-of-thought is never persisted as rationale.
type DecisionProposal struct {
	ID             string               `json:"id"`
	Statement      string               `json:"statement"`
	Classification AnswerClassification `json:"classification"`
	Scope          string               `json:"scope,omitempty"`
	Actor          string               `json:"actor,omitempty"`
	Source         string               `json:"source,omitempty"`
	Authority      string               `json:"authority"`
	Confidence     Confidence           `json:"confidence"`
	Status         DecisionStatus       `json:"status"`
	Affected       []string             `json:"affected,omitempty"`
	Rationale      string               `json:"rationale,omitempty"`
	Alternatives   []string             `json:"alternatives,omitempty"`
	Evidence       []string             `json:"evidence,omitempty"`
	Supersedes     string               `json:"supersedes,omitempty"`
	CreatedAt      string               `json:"created_at,omitempty"`
	ResolvedAt     string               `json:"resolved_at,omitempty"`
}

// Validate enforces the invariants of the canonical Living Plan decision model.
// An agent suggestion is never silently promoted to a user decision.
func (d DecisionProposal) Validate() error {
	if strings.TrimSpace(d.ID) == "" {
		return fmt.Errorf("decision proposal requires an id")
	}
	if strings.TrimSpace(d.Statement) == "" {
		return fmt.Errorf("decision proposal requires a statement")
	}
	if !d.Classification.Valid() {
		return fmt.Errorf("unknown classification %q", d.Classification)
	}
	if AuthorityRank(d.Authority) == 0 {
		return fmt.Errorf("unknown or missing authority %q", d.Authority)
	}
	if !d.Confidence.Valid() {
		return fmt.Errorf("unknown confidence %q", d.Confidence)
	}
	if !d.Status.Valid() {
		return fmt.Errorf("unknown status %q", d.Status)
	}
	if d.Classification == ClassificationAgentSuggestion && d.Status == StatusAccepted {
		return fmt.Errorf("agent suggestion cannot be promoted to accepted without an explicit user decision")
	}
	if d.Classification == ClassificationAgentSuggestion && d.Authority != AuthorityAgentSuggestion {
		return fmt.Errorf("agent suggestion must carry agent-suggestion authority")
	}
	return nil
}

// AgentSuggestion produces a proposed, non-binding decision proposal from
// model content. It always carries the lowest non-external authority and is
// never accepted silently.
func AgentSuggestion(id, statement, scope string) DecisionProposal {
	return DecisionProposal{
		ID:             id,
		Classification: ClassificationAgentSuggestion,
		Authority:      AuthorityAgentSuggestion,
		Confidence:     ConfidenceUnknown,
		Status:         StatusProposed,
		Statement:      statement,
		Scope:          scope,
	}
}

// Resolve promotes a proposal to an explicit user decision when the user has
// confirmed it. It records the source actor and the authority of the user
// decision, replacing the suggestion's own low authority.
func (d DecisionProposal) Resolve(actor string) DecisionProposal {
	d.Classification = ClassificationExplicitDecision
	d.Authority = AuthorityUserDecision
	d.Confidence = ConfidenceUnknown
	d.Status = StatusAccepted
	d.Actor = actor
	return d
}
