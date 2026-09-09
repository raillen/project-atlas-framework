package planning

import (
	"fmt"
	"sort"
	"strings"
)

// Priority ranks a question by impact, not by model curiosity. Lower rank is
// higher impact. Ordering follows the M6 spec: implementation blockers first,
// then irreversible/high-cost decisions, high-risk architecture/security/data,
// user behavior/product semantics, important details, and nice-to-have.
type Priority int

const (
	PriorityBlocker Priority = iota
	PriorityIrreversible
	PriorityHighRisk
	PriorityProduct
	PriorityImportantDetail
	PriorityNiceToHave
)

// String returns the canonical wire value for a Priority.
func (p Priority) String() string {
	switch p {
	case PriorityBlocker:
		return "blocker"
	case PriorityIrreversible:
		return "irreversible"
	case PriorityHighRisk:
		return "high-risk"
	case PriorityProduct:
		return "product"
	case PriorityImportantDetail:
		return "important-detail"
	case PriorityNiceToHave:
		return "nice-to-have"
	default:
		return "unknown"
	}
}

// ParsePriority converts a canonical wire value into a Priority.
func ParsePriority(value string) (Priority, error) {
	switch value {
	case "blocker":
		return PriorityBlocker, nil
	case "irreversible":
		return PriorityIrreversible, nil
	case "high-risk":
		return PriorityHighRisk, nil
	case "product":
		return PriorityProduct, nil
	case "important-detail":
		return PriorityImportantDetail, nil
	case "nice-to-have":
		return PriorityNiceToHave, nil
	default:
		return -1, fmt.Errorf("unknown question priority: %q", value)
	}
}

// OpenQuestion is a still-relevant question for the Interview Engine. It may be
// created without having been sent to the user yet.
type OpenQuestion struct {
	ID              string   `json:"id"`
	Version         int      `json:"version,omitempty"`
	Scope           string   `json:"scope"`
	Contract        string   `json:"contract,omitempty"`
	Topic           string   `json:"topic,omitempty"`
	Priority        string   `json:"priority"`
	Blocking        bool     `json:"blocking,omitempty"`
	Reason          string   `json:"reason,omitempty"`
	SuggestedAnswer []string `json:"suggested_answers,omitempty"`
	Status          string   `json:"status"`
	Owner           string   `json:"owner,omitempty"`
	EligibleOwners  []string `json:"eligible_owners,omitempty"`
	Question        string   `json:"question,omitempty"`
	Answer          string   `json:"answer,omitempty"`
	AskedAt         string   `json:"asked_at,omitempty"`
	ResolvedAt      string   `json:"resolved_at,omitempty"`
	DecisionRef     string   `json:"decision_ref,omitempty"`
	SupersededBy    string   `json:"superseded_by,omitempty"`
	Evidence        []string `json:"evidence,omitempty"`
}

// IsOpen reports whether the question is still awaiting resolution.
func (q OpenQuestion) IsOpen() bool {
	return q.Status == "open" || q.Status == "asked"
}

// Validate checks that the question has the invariants required by M6.
func (q OpenQuestion) Validate() error {
	if strings.TrimSpace(q.ID) == "" {
		return fmt.Errorf("question id is required")
	}
	if strings.TrimSpace(q.Scope) == "" {
		return fmt.Errorf("question scope is required")
	}
	if _, err := ParsePriority(q.Priority); err != nil {
		return err
	}
	switch q.Status {
	case "open", "asked", "resolved", "superseded":
	default:
		return fmt.Errorf("unknown question status: %q", q.Status)
	}
	if q.Blocking && q.Status != "open" && q.Status != "asked" {
		return fmt.Errorf("blocking question %q resolved without an evidence/decision reference", q.ID)
	}
	if (q.Status == "resolved" || q.Status == "superseded") && q.Answer == "" && q.DecisionRef == "" {
		return fmt.Errorf("resolved question %q must carry an answer or a decision reference", q.ID)
	}
	return nil
}

// PriorityOrder returns a copy of questions sorted by impact (highest first),
// stable within the same priority by preserving input order.
func PriorityOrder(questions []OpenQuestion) []OpenQuestion {
	sorted := append([]OpenQuestion{}, questions...)
	sort.SliceStable(sorted, func(i, j int) bool {
		ri, erri := ParsePriority(sorted[i].Priority)
		rj, errj := ParsePriority(sorted[j].Priority)
		if erri != nil && errj != nil {
			return sorted[i].ID < sorted[j].ID
		}
		if erri != nil {
			return false
		}
		if errj != nil {
			return true
		}
		return ri < rj
	})
	return sorted
}

// NextOpen returns the highest-priority question that is still open or asked.
func NextOpen(questions []OpenQuestion) (OpenQuestion, bool) {
	for _, q := range PriorityOrder(questions) {
		if q.IsOpen() {
			return q, true
		}
	}
	return OpenQuestion{}, false
}

// BlockingOpenCount counts implementation-blocking questions still awaiting
// resolution, for coverage reporting.
func BlockingOpenCount(questions []OpenQuestion) int {
	count := 0
	for _, q := range questions {
		if q.Blocking && q.IsOpen() {
			count++
		}
	}
	return count
}
