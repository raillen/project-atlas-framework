package planning

import (
	"fmt"
	"sort"
)

// Contradiction is a detected semantic conflict between decisions that affect
// the same contract/scope but carry different statements. It is detected, not
// resolved silently: the authority model decides who wins.
type Contradiction struct {
	Scope             string `json:"scope"`
	Contract          string `json:"contract,omitempty"`
	ExistingID        string `json:"existing_id"`
	ExistingAuthority string `json:"existing_authority"`
	ExistingStatement string `json:"existing_statement"`
	NewID             string `json:"new_id,omitempty"`
	NewAuthority      string `json:"new_authority,omitempty"`
	NewStatement      string `json:"new_statement"`
}

// AuthorityResolution is the outcome of applying the authority model to a
// detected contradiction. Only a strictly higher authority can override an
// accepted decision; everything else keeps the status quo.
type AuthorityResolution struct {
	Contradiction Contradiction `json:"contradiction"`
	WinnerID      string        `json:"winner_id"`
	Authority     string        `json:"authority"`
	// Rejected indicates the proposed decision was rejected and the existing
	// decision stays authoritative.
	Rejected bool `json:"rejected"`
	// Superseded indicates the proposed decision wins and the existing
	// decision is marked superseded.
	Superseded bool `json:"superseded"`
	// Equal indicates both carry the same authority and neither can override
	// without an explicit amendment.
	Equal bool `json:"equal"`
}

// DetectContradictions scans the given proposals for conflicting decisions:
// sharing an affected contract, both accepted or both binding decisions, with
// materially different statements. Contradictions are keyed deterministically
// by the pair of decision IDs so repeated detection is stable.
func DetectContradictions(proposals []DecisionProposal) []Contradiction {
	var out []Contradiction
	accepted := filterBinding(proposals)
	for i := 0; i < len(accepted); i++ {
		for j := i + 1; j < len(accepted); j++ {
			a, b := accepted[i], accepted[j]
			if a.Scope != b.Scope {
				continue
			}
			if !sharedContract(a, b) {
				continue
			}
			if a.Statement == b.Statement {
				continue
			}
			out = append(out, Contradiction{
				Scope:             a.Scope,
				Contract:          sharedAffected(a, b),
				ExistingID:        lowerID(a.ID, b.ID),
				ExistingStatement: statementFor(a, b, lowerID(a.ID, b.ID)),
				NewID:             higherID(a.ID, b.ID),
				NewStatement:      statementFor(a, b, higherID(a.ID, b.ID)),
			})
		}
	}
	return out
}

// sharedContract reports whether two decisions affect at least one common
// contract/document.
func sharedContract(a, b DecisionProposal) bool {
	for _, x := range a.Affected {
		for _, y := range b.Affected {
			if x == y {
				return true
			}
		}
	}
	return false
}

// sharedAffected returns the first contract that both decisions affect.
func sharedAffected(a, b DecisionProposal) string {
	for _, x := range a.Affected {
		for _, y := range b.Affected {
			if x == y {
				return x
			}
		}
	}
	return ""
}

// ResolveContradiction applies the canonical authority order. winner is the
// decision that stays authoritative; the other side reports whether it was
// superseded or rejected. An equal authority can never unilaterally win and is
// reported as Equal so the caller can raise an explicit amendment decision.
func ResolveContradiction(c Contradiction, existing, proposed DecisionProposal) (AuthorityResolution, error) {
	ra, rb := AuthorityRank(existing.Authority), AuthorityRank(proposed.Authority)
	if ra == 0 || rb == 0 {
		return AuthorityResolution{}, fmt.Errorf("cannot resolve contradiction with unknown authority")
	}
	if ra == rb {
		return AuthorityResolution{
			Contradiction: c,
			WinnerID:      existing.ID,
			Authority:     existing.Authority,
			Equal:         true,
		}, nil
	}
	if ra < rb {
		return AuthorityResolution{
			Contradiction: c,
			WinnerID:      existing.ID,
			Authority:     existing.Authority,
			Rejected:      true,
		}, nil
	}
	return AuthorityResolution{
		Contradiction: c,
		WinnerID:      proposed.ID,
		Authority:     proposed.Authority,
		Superseded:    true,
	}, nil
}

// filterBinding keeps only decisions that are authoritative or already
// accepted; guesses and proposals cannot contradict each other.
func filterBinding(proposals []DecisionProposal) []DecisionProposal {
	var out []DecisionProposal
	for _, p := range proposals {
		if !p.Classification.IsDecision() {
			continue
		}
		if p.Status == StatusAccepted || p.Status == StatusSuperseded {
			out = append(out, p)
		}
	}
	return out
}

func lowerID(a, b string) string {
	if a < b {
		return a
	}
	return b
}

func higherID(a, b string) string {
	if a > b {
		return a
	}
	return b
}

func statementFor(a, b DecisionProposal, id string) string {
	if a.ID == id {
		return a.Statement
	}
	return b.Statement
}

// Preview summarizes the outcome of resolving a batch of open questions
// before any Documentation Delta is produced, covering extracted decisions,
// affected contracts, contradictions/supersessions, resolved blockers and
// remaining open questions.
type Preview struct {
	Extracted        []DecisionProposal    `json:"extracted_decisions"`
	Contradictions   []Contradiction       `json:"contradictions,omitempty"`
	Supersessions    []AuthorityResolution `json:"supersessions,omitempty"`
	BlockersResolved int                   `json:"blockers_resolved"`
	OpenRemaining    int                   `json:"open_remaining"`
	// Mandatory is true when the preview must gate the operation (high-impact
	// or ambiguous), per the M6 spec.
	Mandatory bool `json:"preview_mandatory"`
}

// BuildPreview produces a Preview from a batch of resolution results. Existing
// decisions are included so newly accepted decisions can be checked against
// what already holds. OpenRemaining counts questions still open after the run;
// BlockersResolved counts blocking questions now resolved.
func BuildPreview(results []ResolveResult, existing []DecisionProposal) Preview {
	var extracted []DecisionProposal
	var supersessions []AuthorityResolution
	resolved := 0
	for _, r := range results {
		if r.Proposal.ID == "" {
			continue
		}
		extracted = append(extracted, r.Proposal)
		if r.Resolved && r.BlockedNow {
			resolved++
		}
		for _, c := range DetectContradictions(append(existing, r.Proposal)) {
			for _, e := range existing {
				if e.ID == c.ExistingID {
					res, _ := ResolveContradiction(c, e, r.Proposal)
					supersessions = append(supersessions, res)
					break
				}
			}
		}
	}

	open := 0
	for _, r := range results {
		if r.Question.IsOpen() {
			open++
		}
	}

	preview := Preview{
		Extracted:        extracted,
		Supersessions:    supersessions,
		BlockersResolved: resolved,
		OpenRemaining:    open,
	}
	preview.Contradictions = DetectContradictions(append(existing, extracted...))
	// High-impact or ambiguous operations must gate: any blocker resolved, any
	// supersession/equal authority, anything that contradicts existing state.
	preview.Mandatory = resolved > 0
	for _, s := range supersessions {
		if s.Superseded || s.Equal || s.Rejected {
			preview.Mandatory = true
		}
	}
	sort.Slice(preview.Extracted, func(i, j int) bool {
		return preview.Extracted[i].ID < preview.Extracted[j].ID
	})
	return preview
}
