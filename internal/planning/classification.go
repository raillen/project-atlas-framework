package planning

// AnswerClassification classifies user/model content produced during a planning
// interview. It is kept distinct from the decision status so an answered piece
// of content can be recorded without being silently promoted.
type AnswerClassification string

const (
	ClassificationExplicitDecision AnswerClassification = "explicit-decision"
	ClassificationPreference       AnswerClassification = "preference"
	ClassificationConstraint       AnswerClassification = "constraint"
	ClassificationRequirement      AnswerClassification = "requirement"
	ClassificationNonGoal          AnswerClassification = "non-goal"
	ClassificationUnresolved       AnswerClassification = "unresolved"
	ClassificationHypothesis       AnswerClassification = "hypothesis"
	ClassificationAgentSuggestion  AnswerClassification = "agent-suggestion"
)

func (c AnswerClassification) Valid() bool {
	switch c {
	case ClassificationExplicitDecision, ClassificationPreference, ClassificationConstraint,
		ClassificationRequirement, ClassificationNonGoal, ClassificationUnresolved,
		ClassificationHypothesis, ClassificationAgentSuggestion:
		return true
	}
	return false
}

// IsDecision reports whether the classification is a binding answer a user can
// state, as opposed to an inference the agent produced.
func (c AnswerClassification) IsDecision() bool {
	return c != ClassificationHypothesis && c != ClassificationAgentSuggestion && c != ClassificationUnresolved
}

// Confidence applies to inferences only; it must be explainable by evidence
// pointers, never an opaque model number.
type Confidence string

const (
	ConfidenceHigh    Confidence = "high"
	ConfidenceMedium  Confidence = "medium"
	ConfidenceLow     Confidence = "low"
	ConfidenceUnknown Confidence = "unknown"
)

func (c Confidence) Valid() bool {
	switch c {
	case ConfidenceHigh, ConfidenceMedium, ConfidenceLow, ConfidenceUnknown:
		return true
	}
	return false
}

// Semantic authority order (highest first) from the canonical Living Plan spec.
const (
	AuthorityInvariant          = "invariant"           // locked canonical invariant / spec / ADR
	AuthorityUserDecision       = "user-decision"       // explicit current user decision within permitted authority
	AuthorityProjectDecision    = "project-decision"    // accepted project decision
	AuthorityDocumentedEvidence = "documented-evidence" // documented constraint / evidence
	AuthorityInferredState      = "inferred-state"      // inferred existing state
	AuthorityAgentSuggestion    = "agent-suggestion"    // agent suggestion
	AuthorityExternalContent    = "external"            // external / untrusted content
)

// AuthorityRank maps an authority to its semantic order. Lower value = higher authority.
func AuthorityRank(authority string) int {
	switch authority {
	case AuthorityInvariant:
		return 1
	case AuthorityUserDecision:
		return 2
	case AuthorityProjectDecision:
		return 3
	case AuthorityDocumentedEvidence:
		return 4
	case AuthorityInferredState:
		return 5
	case AuthorityAgentSuggestion:
		return 6
	case AuthorityExternalContent:
		return 7
	}
	return 0
}
