package adoption

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// MappingCandidate represents a candidate mapping from existing brownfield repository
// documents to an Prumo Documentation Contract, without forcing filename changes.
type MappingCandidate struct {
	ID         string         `json:"id"`
	ContractID string         `json:"contract_id"`
	Sources    []string       `json:"sources"`
	Confidence FactConfidence `json:"confidence"`
	Evidence   []string       `json:"evidence"`
	Rationale  string         `json:"rationale,omitempty"`
	Authority  string         `json:"authority"` // "inferred-state"
	Status     string         `json:"status"`    // "candidate", "accepted", "rejected"
}

// Validate checks invariants for a MappingCandidate.
func (m MappingCandidate) Validate() error {
	if strings.TrimSpace(m.ID) == "" {
		return fmt.Errorf("mapping candidate requires an id")
	}
	if strings.TrimSpace(m.ContractID) == "" {
		return fmt.Errorf("mapping candidate requires a contract id")
	}
	if len(m.Sources) == 0 {
		return fmt.Errorf("mapping candidate requires at least one source file")
	}
	for _, s := range m.Sources {
		if strings.TrimSpace(s) == "" {
			return fmt.Errorf("source path cannot be empty")
		}
	}
	if !m.Confidence.Valid() {
		return fmt.Errorf("invalid confidence: %q", m.Confidence)
	}
	if len(m.Evidence) == 0 {
		return fmt.Errorf("mapping candidate requires at least one evidence fact ID")
	}
	if strings.TrimSpace(m.Authority) == "" {
		return fmt.Errorf("mapping candidate requires an authority")
	}
	switch m.Status {
	case "candidate", "accepted", "rejected":
	default:
		return fmt.Errorf("invalid mapping candidate status: %q", m.Status)
	}
	return nil
}

// CandidateBinding models a proposed documentation binding ready for consumption
// by the M5 documentation coverage engine.
type CandidateBinding struct {
	ContractID string   `json:"contract_id"`
	Sources    []string `json:"sources"`
	Ownership  string   `json:"ownership"`
	Authority  string   `json:"authority"`
	Evidence   []string `json:"evidence,omitempty"`
}

// MappingResult contains all semantic documentation mapping candidates and candidate bindings.
type MappingResult struct {
	Version    int                `json:"version"`
	Candidates []MappingCandidate `json:"candidates"`
	Bindings   []CandidateBinding `json:"bindings"`
}

// contractMappingRule defines how to match documents to a contract.
type contractMappingRule struct {
	ContractID string
	Keywords   []string
	ExactFiles []string
	DocRoles   []string
	Rationale  string
}

var standardContractRules = []contractMappingRule{
	{
		ContractID: "product.vision",
		ExactFiles: []string{"readme.md"},
		DocRoles:   []string{"product-vision"},
		Rationale:  "Primary project readme introduces vision, problem statement, and goals",
	},
	{
		ContractID: "project.scope",
		Keywords:   []string{"scope", "non-goals", "roadmap", "goals"},
		Rationale:  "Document contains project scope, non-goals, or roadmap boundaries",
	},
	{
		ContractID: "architecture.system",
		Keywords:   []string{"architecture", "system", "design", "overview", "internals", "topology"},
		DocRoles:   []string{"system-architecture"},
		Rationale:  "Document covers system design, components, boundaries, or architecture",
	},
	{
		ContractID: "testing.strategy",
		Keywords:   []string{"testing", "test-strategy", "qa", "verification", "conformance"},
		ExactFiles: []string{"testing.md"},
		DocRoles:   []string{"testing-strategy"},
		Rationale:  "Document details test suite structure, quality gates, or testing strategy",
	},
	{
		ContractID: "security.trust",
		Keywords:   []string{"security", "trust", "threat", "trust-model"},
		ExactFiles: []string{"security.md"},
		DocRoles:   []string{"security-trust"},
		Rationale:  "Document addresses security policy, trust boundaries, or permissions",
	},
	{
		ContractID: "installation.lifecycle",
		Keywords:   []string{"install", "installation", "setup", "getting-started", "lifecycle"},
		ExactFiles: []string{"install.md"},
		DocRoles:   []string{"installation"},
		Rationale:  "Document outlines installation, configuration, or environment setup",
	},
	{
		ContractID: "cli.reference",
		Keywords:   []string{"cli", "commands", "cli-reference", "usage", "flags"},
		DocRoles:   []string{"cli-reference"},
		Rationale:  "Document provides CLI usage, command reference, or flag documentation",
	},
	{
		ContractID: "ui.documentation",
		Keywords:   []string{"ui", "components", "design-system", "screens", "wireframes", "storybook"},
		Rationale:  "Document describes UI screens, design tokens, or component states",
	},
}

// MapDocumentation maps discovered documentation and ADR facts to documentation contracts.
func MapDocumentation(facts []ObservedFact, class RepositoryClassification) MappingResult {
	res := MappingResult{
		Version:    1,
		Candidates: make([]MappingCandidate, 0),
		Bindings:   make([]CandidateBinding, 0),
	}

	docFacts := make([]ObservedFact, 0)
	for _, f := range facts {
		if f.Kind == FactDoc || f.Kind == FactADR {
			docFacts = append(docFacts, f)
		}
	}

	counter := 1
	for _, rule := range standardContractRules {
		var matchedSources []string
		var matchedEvidence []string
		bestConf := ConfidenceMedium

		for _, f := range docFacts {
			lowerSource := strings.ToLower(f.Source)
			base := filepath.Base(lowerSource)

			matched := false

			// Check exact file match
			for _, ef := range rule.ExactFiles {
				if base == ef {
					matched = true
					bestConf = ConfidenceHigh
					break
				}
			}

			// Check keyword match in file path
			if !matched {
				for _, kw := range rule.Keywords {
					if strings.Contains(lowerSource, kw) {
						matched = true
						if bestConf != ConfidenceHigh {
							bestConf = ConfidenceHigh
						}
						break
					}
				}
			}

			// Check signal doc-role match
			if !matched && len(rule.DocRoles) > 0 {
				for _, s := range class.Signals {
					if s.Dimension == DimDocRole {
						for _, r := range rule.DocRoles {
							if s.Value == r {
								for _, evID := range s.Evidence {
									if evID == f.ID {
										matched = true
										break
									}
								}
							}
						}
					}
				}
			}

			if matched {
				matchedSources = append(matchedSources, f.Source)
				matchedEvidence = append(matchedEvidence, f.ID)
			}
		}

		matchedSources = uniqueStrings(matchedSources)
		matchedEvidence = uniqueStrings(matchedEvidence)

		if len(matchedSources) > 0 {
			candidateID := fmt.Sprintf("map-%04d", counter)
			candidate := MappingCandidate{
				ID:         candidateID,
				ContractID: rule.ContractID,
				Sources:    matchedSources,
				Confidence: bestConf,
				Evidence:   matchedEvidence,
				Rationale:  rule.Rationale,
				Authority:  "inferred-state",
				Status:     "candidate",
			}

			binding := CandidateBinding{
				ContractID: rule.ContractID,
				Sources:    matchedSources,
				Ownership:  "human-maintained",
				Authority:  "inferred-state",
				Evidence:   matchedEvidence,
			}

			res.Candidates = append(res.Candidates, candidate)
			res.Bindings = append(res.Bindings, binding)
			counter++
		}
	}

	// Sort candidates and bindings stably by ContractID
	sort.Slice(res.Candidates, func(i, j int) bool {
		return res.Candidates[i].ContractID < res.Candidates[j].ContractID
	})
	sort.Slice(res.Bindings, func(i, j int) bool {
		return res.Bindings[i].ContractID < res.Bindings[j].ContractID
	})

	return res
}
