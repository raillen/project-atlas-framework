package adoption

import (
	"fmt"
	"sort"
	"strings"
)

// LedgerCategory classifies the domain of an inference in the Confidence Ledger.
type LedgerCategory string

const (
	CategoryClassification LedgerCategory = "classification"
	CategoryCapability     LedgerCategory = "capability"
	CategoryProfile        LedgerCategory = "profile"
	CategoryMapping        LedgerCategory = "mapping"
	CategoryOther          LedgerCategory = "other"
)

func (c LedgerCategory) Valid() bool {
	switch c {
	case CategoryClassification, CategoryCapability, CategoryProfile, CategoryMapping, CategoryOther:
		return true
	}
	return false
}

// LedgerStatus represents the review/acceptance state of an inference.
type LedgerStatus string

const (
	LedgerStatusUnresolved LedgerStatus = "unresolved"
	LedgerStatusAccepted   LedgerStatus = "accepted"
	LedgerStatusRejected   LedgerStatus = "rejected"
)

func (s LedgerStatus) Valid() bool {
	switch s {
	case LedgerStatusUnresolved, LedgerStatusAccepted, LedgerStatusRejected:
		return true
	}
	return false
}

// LedgerEntry represents a single non-trivial inference tracked in the Confidence Ledger.
type LedgerEntry struct {
	ID                   string         `json:"id"`
	Claim                string         `json:"claim"`
	Category             LedgerCategory `json:"category"`
	Confidence           FactConfidence `json:"confidence"`
	Score                float64        `json:"score"`
	Evidence             []string       `json:"evidence"`
	Alternatives         []string       `json:"alternatives,omitempty"`
	Status               LedgerStatus   `json:"status"`
	Authority            string         `json:"authority"` // "inferred-state"
	Rationale            string         `json:"rationale,omitempty"`
	RequiresConfirmation bool           `json:"requires_confirmation"`
}

// Validate checks invariants for a LedgerEntry.
func (e LedgerEntry) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return fmt.Errorf("ledger entry requires an id")
	}
	if strings.TrimSpace(e.Claim) == "" {
		return fmt.Errorf("ledger entry requires a claim statement")
	}
	if !e.Category.Valid() {
		return fmt.Errorf("invalid ledger category: %q", e.Category)
	}
	if !e.Confidence.Valid() {
		return fmt.Errorf("invalid confidence: %q", e.Confidence)
	}
	if e.Score < 0.0 || e.Score > 1.0 {
		return fmt.Errorf("score must be between 0.0 and 1.0, got %f", e.Score)
	}
	if len(e.Evidence) == 0 {
		return fmt.Errorf("ledger entry requires at least one supporting evidence fact ID")
	}
	if !e.Status.Valid() {
		return fmt.Errorf("invalid ledger status: %q", e.Status)
	}
	if strings.TrimSpace(e.Authority) == "" {
		return fmt.Errorf("ledger entry requires an authority")
	}
	// Invariant: low confidence inferences must require confirmation
	if (e.Confidence == ConfidenceLow || e.Confidence == ConfidenceMedium) && !e.RequiresConfirmation && e.Status != LedgerStatusAccepted {
		return fmt.Errorf("medium/low confidence inference must require confirmation")
	}
	return nil
}

// LedgerSummary aggregates metric counts for entries in the Confidence Ledger.
type LedgerSummary struct {
	Total                int `json:"total"`
	Factual              int `json:"factual"`
	High                 int `json:"high"`
	Medium               int `json:"medium"`
	Low                  int `json:"low"`
	Unresolved           int `json:"unresolved"`
	RequiresConfirmation int `json:"requires_confirmation"`
}

// ConfidenceLedger is the canonical container recording all non-trivial inferences
// produced during the Adoption Engine evaluation.
type ConfidenceLedger struct {
	Version int           `json:"version"`
	Entries []LedgerEntry `json:"entries"`
	Summary LedgerSummary `json:"summary"`
}

// ScoreFromConfidence maps a FactConfidence to a default numeric confidence score.
func ScoreFromConfidence(conf FactConfidence) float64 {
	switch conf {
	case ConfidenceFactual:
		return 1.0
	case ConfidenceHigh:
		return 0.85
	case ConfidenceMedium:
		return 0.60
	case ConfidenceLow:
		return 0.30
	default:
		return 0.0
	}
}

// BuildConfidenceLedger constructs a comprehensive ConfidenceLedger synthesizing inferences
// from classification, capability/profile proposals, and semantic doc mapping.
func BuildConfidenceLedger(class RepositoryClassification, prop ProposalsResult, mapRes MappingResult) ConfidenceLedger {
	ledger := ConfidenceLedger{
		Version: 1,
		Entries: make([]LedgerEntry, 0),
	}

	counter := 1

	// 1. Inferences from Classification Signals (app-type, frameworks, persistence)
	for _, s := range class.Signals {
		if s.Dimension == DimAppType || s.Dimension == DimFramework || s.Dimension == DimPersistence {
			reqConfirm := s.Confidence == ConfidenceLow || s.Confidence == ConfidenceMedium || s.Value == "unknown"
			claim := fmt.Sprintf("Repository exhibits %s %q", s.Dimension, s.Value)
			if s.Dimension == DimAppType {
				claim = fmt.Sprintf("Candidate project archetype is %q", s.Value)
			}

			var alts []string
			if s.Dimension == DimAppType && s.Value != "unknown" {
				for _, at := range class.AppTypes {
					if at != s.Value && at != "unknown" {
						alts = append(alts, at)
					}
				}
			}

			entry := LedgerEntry{
				ID:                   fmt.Sprintf("cld-%04d", counter),
				Claim:                claim,
				Category:             CategoryClassification,
				Confidence:           s.Confidence,
				Score:                ScoreFromConfidence(s.Confidence),
				Evidence:             s.Evidence,
				Alternatives:         alts,
				Status:               LedgerStatusUnresolved,
				Authority:            "inferred-state",
				Rationale:            fmt.Sprintf("Extracted from %s classification signal with %d supporting evidence item(s)", s.Dimension, len(s.Evidence)),
				RequiresConfirmation: reqConfirm,
			}
			ledger.Entries = append(ledger.Entries, entry)
			counter++
		}
	}

	// 2. Inferences from Capability Proposals
	for _, cp := range prop.Capabilities {
		reqConfirm := cp.Confidence == ConfidenceLow || cp.Confidence == ConfidenceMedium
		entry := LedgerEntry{
			ID:                   fmt.Sprintf("cld-%04d", counter),
			Claim:                fmt.Sprintf("Proposed capability %q", cp.Capability),
			Category:             CategoryCapability,
			Confidence:           cp.Confidence,
			Score:                ScoreFromConfidence(cp.Confidence),
			Evidence:             cp.Evidence,
			Status:               LedgerStatusUnresolved,
			Authority:            "inferred-state",
			Rationale:            cp.Rationale,
			RequiresConfirmation: reqConfirm,
		}
		ledger.Entries = append(ledger.Entries, entry)
		counter++
	}

	// 3. Inferences from Profile Candidates
	var allProfileNames []string
	for _, p := range prop.Profiles {
		allProfileNames = append(allProfileNames, p.Profile)
	}

	for _, p := range prop.Profiles {
		reqConfirm := p.Confidence == ConfidenceLow || p.Confidence == ConfidenceMedium
		var alts []string
		for _, name := range allProfileNames {
			if name != p.Profile && name != "core-software" {
				alts = append(alts, name)
			}
		}

		entry := LedgerEntry{
			ID:                   fmt.Sprintf("cld-%04d", counter),
			Claim:                fmt.Sprintf("Recommended documentation profile %q", p.Profile),
			Category:             CategoryProfile,
			Confidence:           p.Confidence,
			Score:                ScoreFromConfidence(p.Confidence),
			Evidence:             p.Evidence,
			Alternatives:         alts,
			Status:               LedgerStatusUnresolved,
			Authority:            "inferred-state",
			Rationale:            p.Rationale,
			RequiresConfirmation: reqConfirm,
		}
		ledger.Entries = append(ledger.Entries, entry)
		counter++
	}

	// 4. Inferences from Mapping Candidates
	for _, mc := range mapRes.Candidates {
		reqConfirm := mc.Confidence == ConfidenceLow || mc.Confidence == ConfidenceMedium
		entry := LedgerEntry{
			ID:                   fmt.Sprintf("cld-%04d", counter),
			Claim:                fmt.Sprintf("Document %s mapped to contract %q", strings.Join(mc.Sources, ", "), mc.ContractID),
			Category:             CategoryMapping,
			Confidence:           mc.Confidence,
			Score:                ScoreFromConfidence(mc.Confidence),
			Evidence:             mc.Evidence,
			Status:               LedgerStatusUnresolved,
			Authority:            "inferred-state",
			Rationale:            mc.Rationale,
			RequiresConfirmation: reqConfirm,
		}
		ledger.Entries = append(ledger.Entries, entry)
		counter++
	}

	// Stable sort by ID
	sort.Slice(ledger.Entries, func(i, j int) bool {
		return ledger.Entries[i].ID < ledger.Entries[j].ID
	})

	ledger.RecomputeSummary()
	return ledger
}

// RecomputeSummary refreshes summary metric counters from the current entries.
func (l *ConfidenceLedger) RecomputeSummary() {
	s := LedgerSummary{
		Total: len(l.Entries),
	}
	for _, e := range l.Entries {
		switch e.Confidence {
		case ConfidenceFactual:
			s.Factual++
		case ConfidenceHigh:
			s.High++
		case ConfidenceMedium:
			s.Medium++
		case ConfidenceLow:
			s.Low++
		}
		if e.Status == LedgerStatusUnresolved {
			s.Unresolved++
		}
		if e.RequiresConfirmation {
			s.RequiresConfirmation++
		}
	}
	l.Summary = s
}

// UnresolvedEntries returns all entries that have not yet been accepted or rejected.
func (l *ConfidenceLedger) UnresolvedEntries() []LedgerEntry {
	var out []LedgerEntry
	for _, e := range l.Entries {
		if e.Status == LedgerStatusUnresolved {
			out = append(out, e)
		}
	}
	return out
}

// RequiringConfirmation returns all entries that require human or interview verification.
func (l *ConfidenceLedger) RequiringConfirmation() []LedgerEntry {
	var out []LedgerEntry
	for _, e := range l.Entries {
		if e.RequiresConfirmation {
			out = append(out, e)
		}
	}
	return out
}

// UpdateStatus sets the resolution status of an entry and recomputes the summary.
func (l *ConfidenceLedger) UpdateStatus(id string, status LedgerStatus) error {
	if !status.Valid() {
		return fmt.Errorf("invalid status: %q", status)
	}
	for i := range l.Entries {
		if l.Entries[i].ID == id {
			l.Entries[i].Status = status
			if status == LedgerStatusAccepted {
				l.Entries[i].RequiresConfirmation = false
			}
			l.RecomputeSummary()
			return nil
		}
	}
	return fmt.Errorf("ledger entry %q not found", id)
}
