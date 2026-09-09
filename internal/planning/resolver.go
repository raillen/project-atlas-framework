package planning

import (
	"fmt"
	"strings"
)

// Source identifies where a piece of planning content came from. It is used to
// decide which authorities are legitimate, not to take the answer at face value.
type Source string

const (
	SourceUser  Source = "user"
	SourceRepo  Source = "repository"
	SourceModel Source = "model"
)

// ResolveRequest carries a classified answer to an open question together with
// the context needed to turn it into a decision proposal.
type ResolveRequest struct {
	Question       OpenQuestion
	Statement      string
	Classification AnswerClassification
	Source         Source
	Actor          string
	Evidence       []string
}

// ResolveResult is the outcome of resolving one open question. Proposal is
// empty when the answer does not produce a decision (unresolved answer).
type ResolveResult struct {
	Proposal   DecisionProposal
	Resolved   bool
	Question   OpenQuestion
	BlockedNow bool
}

// ConfidenceReason explains an inferred confidence with a short reason code so
// it is explainable instead of being an opaque model number.
type ConfidenceReason struct {
	Confidence Confidence `json:"confidence"`
	Reason     string     `json:"reason"`
	Evidence   []string   `json:"evidence,omitempty"`
}

// Resolve maps a classified answer onto the authority model, decides a
// confidence (only for inferences), and either accepts, keeps proposed, or
// keeps a question open.
//
// Invariants enforced:
//   - an agent suggestion is never promoted to accepted here;
//   - confidence is only produced for inferences (explicit decisions carry none);
//   - a user decision requires an explicit user source/actor.
func Resolve(req ResolveRequest) (ResolveResult, error) {
	var result ResolveResult
	result.Question = req.Question

	if err := req.Question.Validate(); err != nil {
		return result, err
	}
	if !req.Classification.Valid() {
		return result, fmt.Errorf("unknown classification %q", req.Classification)
	}
	if err := validateSourceAuthority(req); err != nil {
		return result, err
	}

	switch req.Classification {
	case ClassificationUnresolved:
		return result, nil

	case ClassificationHypothesis:
		result.Proposal = hypothesisProposal(req)
		return result, nil

	case ClassificationAgentSuggestion:
		result.Proposal = suggestionProposal(req)
		return result, nil

	default:
		proposal, err := acceptedProposal(req)
		if err != nil {
			return result, err
		}
		result.Proposal = proposal
		result.Resolved = true
		result.Question.Status = "resolved"
		result.Question.Answer = req.Statement
		result.Question.DecisionRef = proposal.ID
		result.Question.ResolvedAt = proposal.ResolvedAt
		result.BlockedNow = result.Question.Blocking
		return result, nil
	}
}

func validateSourceAuthority(req ResolveRequest) error {
	if req.Classification.IsDecision() {
		if req.Source != SourceUser && req.Source != SourceRepo {
			return fmt.Errorf("decision classification %q requires a user or repository source", req.Classification)
		}
		if req.Source == SourceUser && strings.TrimSpace(req.Actor) == "" {
			return fmt.Errorf("user decision requires an actor")
		}
		return nil
	}
	// Inferences may come from the model; repository content cannot answer a
	// user-authority question on its own but can be the basis of evidence.
	if req.Source != SourceUser && req.Source != SourceModel && req.Source != SourceRepo {
		return fmt.Errorf("unknown source %q", req.Source)
	}
	return nil
}

func hypothesisProposal(req ResolveRequest) DecisionProposal {
	reason := ConfidenceReason{Confidence: inferConfidence(req.Evidence), Reason: "inference"}
	return DecisionProposal{
		ID:             proposalID(req, "hyp"),
		Statement:      req.Statement,
		Classification: ClassificationHypothesis,
		Scope:          req.Question.Scope,
		Source:         string(req.Source),
		Actor:          req.Actor,
		Authority:      AuthorityInferredState,
		Confidence:     reason.Confidence,
		Status:         StatusProposed,
		Evidence:       req.Evidence,
	}
}

func suggestionProposal(req ResolveRequest) DecisionProposal {
	d := AgentSuggestion(proposalID(req, "sug"), req.Statement, req.Question.Scope)
	d.Evidence = req.Evidence
	return d
}

func acceptedProposal(req ResolveRequest) (DecisionProposal, error) {
	authority := AuthorityProjectDecision
	if req.Source == SourceUser {
		authority = AuthorityUserDecision
	}
	d := DecisionProposal{
		ID:             proposalID(req, "dec"),
		Statement:      req.Statement,
		Classification: req.Classification,
		Scope:          req.Question.Scope,
		Source:         string(req.Source),
		Actor:          req.Actor,
		Authority:      authority,
		Confidence:     ConfidenceUnknown,
		Status:         StatusAccepted,
		Evidence:       req.Evidence,
	}
	if req.Question.Contract != "" {
		d.Affected = []string{req.Question.Contract}
	}
	if err := d.Validate(); err != nil {
		return d, err
	}
	return d, nil
}

// proposalID builds a deterministic, human-readable proposal id from the
// question id and a short suffix. Collisions are avoided by the caller by
// choosing suffixes per classification.
func proposalID(req ResolveRequest, suffix string) string {
	return fmt.Sprintf("DP-%s-%s", req.Question.ID, suffix)
}

// inferConfidence avoids false precision: no evidence means unknown; evidence
// pointers support a low-to-medium inference confidence. It never applies to
// explicit decisions.
func inferConfidence(evidence []string) Confidence {
	if len(evidence) == 0 {
		return ConfidenceUnknown
	}
	return ConfidenceLow
}

// ResolveMany resolves a batch of questions, respecting the priority order.
// The resulting proposals keep the resolver invariant: no agent suggestion is
// silently promoted.
func ResolveMany(questions []OpenQuestion, reqs []ResolveRequest) ([]ResolveResult, error) {
	if len(questions) != len(reqs) {
		return nil, fmt.Errorf("questions and requests must be paired")
	}
	results := make([]ResolveResult, 0, len(questions))
	for i := range questions {
		res, err := Resolve(reqs[i])
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}
	return results, nil
}
