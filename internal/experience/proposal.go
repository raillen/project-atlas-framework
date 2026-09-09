package experience

import (
	"fmt"
	"strings"
	"time"
)

// ExperienceProposalStatus models the review status of an experience pattern.
type ExperienceProposalStatus string

const (
	ProposalStatusProposed ExperienceProposalStatus = "proposed"
	ProposalStatusApproved ExperienceProposalStatus = "approved"
	ProposalStatusRejected ExperienceProposalStatus = "rejected"
)

// ExperienceProposal captures candidate patterns and procedural heuristics learned from sessions.
// Invariant: Proposals are governed by the Unified Review Queue before being promoted to canonical skills/rules.
type ExperienceProposal struct {
	ID             string                   `json:"id"`
	Pattern        string                   `json:"pattern"`
	Observation    string                   `json:"observation"`
	ProposedAction string                   `json:"proposed_action"`
	SourceSession  string                   `json:"source_session"`
	Status         ExperienceProposalStatus `json:"status"`
	ReviewRequired bool                     `json:"review_required"`
	ReviewNotes    string                   `json:"review_notes,omitempty"`
	ApprovedBy     string                   `json:"approved_by,omitempty"`
	CreatedAt      string                   `json:"created_at"`
}

// ProposeExperience constructs a review-governed experience proposal.
func ProposeExperience(id, pattern, observation, action, sessionID string) (ExperienceProposal, error) {
	if strings.TrimSpace(id) == "" {
		return ExperienceProposal{}, fmt.Errorf("experience proposal requires an ID")
	}
	if strings.TrimSpace(pattern) == "" {
		return ExperienceProposal{}, fmt.Errorf("experience proposal requires a pattern name")
	}
	if strings.TrimSpace(observation) == "" {
		return ExperienceProposal{}, fmt.Errorf("experience proposal requires an observation")
	}
	return ExperienceProposal{
		ID:             id,
		Pattern:        pattern,
		Observation:    observation,
		ProposedAction: action,
		SourceSession:  sessionID,
		Status:         ProposalStatusProposed,
		ReviewRequired: true,
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// Approve accepts an experience proposal via human or governance review.
func (p *ExperienceProposal) Approve(actor, notes string) error {
	if p.Status != ProposalStatusProposed {
		return fmt.Errorf("cannot approve proposal in status %q", p.Status)
	}
	p.Status = ProposalStatusApproved
	p.ApprovedBy = actor
	p.ReviewNotes = notes
	return nil
}

// Reject discards an experience proposal with rationale.
func (p *ExperienceProposal) Reject(actor, notes string) error {
	if p.Status != ProposalStatusProposed {
		return fmt.Errorf("cannot reject proposal in status %q", p.Status)
	}
	p.Status = ProposalStatusRejected
	p.ApprovedBy = actor
	p.ReviewNotes = notes
	return nil
}
