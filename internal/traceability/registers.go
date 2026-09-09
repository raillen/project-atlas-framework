package traceability

import (
	"fmt"
	"strings"
	"time"
)

// ExperimentEntry tracks hypotheses tested, validation results, and architectural decisions.
type ExperimentEntry struct {
	ID         string   `json:"id"`
	Goal       string   `json:"goal"`
	Hypothesis string   `json:"hypothesis"`
	Method     string   `json:"method"`
	Outcome    string   `json:"outcome"`
	Conclusion string   `json:"conclusion"`
	Evidence   []string `json:"evidence,omitempty"`
	CreatedAt  string   `json:"created_at"`
}

// RejectionEntry tracks discarded proposals, alternative paths, and explicit rationale for rejection.
type RejectionEntry struct {
	ID          string   `json:"id"`
	ProposalID  string   `json:"proposal_id"`
	Title       string   `json:"title"`
	Alternative string   `json:"alternative"`
	Rationale   string   `json:"rationale"`
	RejectedBy  string   `json:"rejected_by"`
	Evidence    []string `json:"evidence,omitempty"`
	CreatedAt   string   `json:"created_at"`
}

// DebtSeverity indicates the operational or maintainability impact of technical debt.
type DebtSeverity string

const (
	DebtSeverityLow      DebtSeverity = "low"
	DebtSeverityMedium   DebtSeverity = "medium"
	DebtSeverityHigh     DebtSeverity = "high"
	DebtSeverityCritical DebtSeverity = "critical"
)

// DebtEntry records known architectural or implementation debt, impacted contracts, and planned remediation.
type DebtEntry struct {
	ID                string       `json:"id"`
	Title             string       `json:"title"`
	Description       string       `json:"description"`
	ImpactedContracts []string     `json:"impacted_contracts,omitempty"`
	Severity          DebtSeverity `json:"severity"`
	RemediationPlan   string       `json:"remediation_plan"`
	Status            string       `json:"status"` // "open", "remediated", "accepted"
	CreatedAt         string       `json:"created_at"`
}

// Registers aggregates Experiment, Rejection, and Debt records.
type Registers struct {
	Experiments []ExperimentEntry `json:"experiments"`
	Rejections  []RejectionEntry  `json:"rejections"`
	Debt        []DebtEntry       `json:"debt"`
}

// NewRegisters initializes empty registers.
func NewRegisters() *Registers {
	return &Registers{
		Experiments: make([]ExperimentEntry, 0),
		Rejections:  make([]RejectionEntry, 0),
		Debt:        make([]DebtEntry, 0),
	}
}

// RecordExperiment appends a validated experiment record.
func (r *Registers) RecordExperiment(e ExperimentEntry) error {
	if strings.TrimSpace(e.ID) == "" {
		return fmt.Errorf("experiment entry requires an ID")
	}
	if strings.TrimSpace(e.Hypothesis) == "" {
		return fmt.Errorf("experiment entry requires a hypothesis")
	}
	if e.CreatedAt == "" {
		e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	r.Experiments = append(r.Experiments, e)
	return nil
}

// RecordRejection appends an explicit rejection record.
func (r *Registers) RecordRejection(rej RejectionEntry) error {
	if strings.TrimSpace(rej.ID) == "" {
		return fmt.Errorf("rejection entry requires an ID")
	}
	if strings.TrimSpace(rej.Rationale) == "" {
		return fmt.Errorf("rejection entry requires a rationale")
	}
	if rej.CreatedAt == "" {
		rej.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	r.Rejections = append(r.Rejections, rej)
	return nil
}

// RecordDebt appends a tracked debt item.
func (r *Registers) RecordDebt(d DebtEntry) error {
	if strings.TrimSpace(d.ID) == "" {
		return fmt.Errorf("debt entry requires an ID")
	}
	if strings.TrimSpace(d.Title) == "" {
		return fmt.Errorf("debt entry requires a title")
	}
	if d.Status == "" {
		d.Status = "open"
	}
	if d.CreatedAt == "" {
		d.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	r.Debt = append(r.Debt, d)
	return nil
}
