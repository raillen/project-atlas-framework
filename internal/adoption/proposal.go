package adoption

import (
	"fmt"
	"sort"
	"strings"
)

// CapabilityProposal represents a candidate capability inferred from repository classification,
// awaiting user confirmation before becoming canonical.
type CapabilityProposal struct {
	ID         string         `json:"id"`
	Capability string         `json:"capability"`
	Source     string         `json:"source"`
	Confidence FactConfidence `json:"confidence"`
	Evidence   []string       `json:"evidence"`
	Rationale  string         `json:"rationale,omitempty"`
	Status     string         `json:"status"` // "proposed", "accepted", "rejected"
}

// Validate checks invariants for a CapabilityProposal.
func (p CapabilityProposal) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("capability proposal requires an id")
	}
	if strings.TrimSpace(p.Capability) == "" {
		return fmt.Errorf("capability proposal requires a capability identifier")
	}
	if !p.Confidence.Valid() {
		return fmt.Errorf("invalid confidence: %q", p.Confidence)
	}
	if len(p.Evidence) == 0 {
		return fmt.Errorf("capability proposal requires at least one evidence fact ID")
	}
	switch p.Status {
	case "proposed", "accepted", "rejected":
	default:
		return fmt.Errorf("invalid capability proposal status: %q", p.Status)
	}
	return nil
}

// ProfileCandidate represents a candidate Documentation or Project Profile
// proposed based on inferred capabilities and app types.
type ProfileCandidate struct {
	ID           string         `json:"id"`
	Profile      string         `json:"profile"`
	Confidence   FactConfidence `json:"confidence"`
	Capabilities []string       `json:"capabilities"`
	Contracts    []string       `json:"contracts"`
	Evidence     []string       `json:"evidence"`
	Rationale    string         `json:"rationale,omitempty"`
}

// Validate checks invariants for a ProfileCandidate.
func (p ProfileCandidate) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("profile candidate requires an id")
	}
	if strings.TrimSpace(p.Profile) == "" {
		return fmt.Errorf("profile candidate requires a profile name")
	}
	if !p.Confidence.Valid() {
		return fmt.Errorf("invalid confidence: %q", p.Confidence)
	}
	if len(p.Evidence) == 0 {
		return fmt.Errorf("profile candidate requires at least one evidence fact ID")
	}
	return nil
}

// ProposalsResult is the complete container for candidate capability and profile proposals.
type ProposalsResult struct {
	Version      int                  `json:"version"`
	Capabilities []CapabilityProposal `json:"capabilities"`
	Profiles     []ProfileCandidate   `json:"profiles"`
}

// builtinProfileDefinition defines standard profile mappings matching docs/profiles/builtin.json.
type builtinProfileDefinition struct {
	Capabilities []string
	Contracts    []string
	Description  string
}

var standardProfiles = map[string]builtinProfileDefinition{
	"core-software": {
		Capabilities: []string{"core"},
		Contracts: []string{
			"product.vision",
			"project.scope",
			"architecture.system",
			"testing.strategy",
			"security.trust",
			"cli.reference",
			"installation.lifecycle",
		},
		Description: "Common engineering foundation",
	},
	"cli": {
		Capabilities: []string{"cli"},
		Contracts:    []string{"cli.reference"},
		Description:  "Command-line projects",
	},
	"web-application": {
		Capabilities: []string{"web-application", "ui"},
		Contracts:    []string{"ui.documentation"},
		Description:  "Web applications",
	},
	"desktop-gui": {
		Capabilities: []string{"desktop-gui", "ui"},
		Contracts:    []string{"ui.documentation"},
		Description:  "Desktop graphical applications",
	},
	"api-service": {
		Capabilities: []string{"api-service"},
		Contracts: []string{
			"architecture.system",
			"testing.strategy",
			"security.trust",
		},
		Description: "HTTP/API backend services",
	},
	"compiler": {
		Capabilities: []string{"compiler"},
		Contracts: []string{
			"architecture.system",
			"testing.strategy",
		},
		Description: "Compilers and language tooling",
	},
	"library": {
		Capabilities: []string{"library"},
		Contracts: []string{
			"architecture.system",
			"testing.strategy",
		},
		Description: "Libraries, SDKs and reusable packages",
	},
}

// Propose evaluates a RepositoryClassification and generates candidate CapabilityProposals
// and ProfileCandidates.
func Propose(class RepositoryClassification) ProposalsResult {
	res := ProposalsResult{
		Version:      1,
		Capabilities: make([]CapabilityProposal, 0),
		Profiles:     make([]ProfileCandidate, 0),
	}

	capEvidenceMap := make(map[string][]string)
	capConfidenceMap := make(map[string]FactConfidence)

	// Collect evidence and highest confidence for each capability
	for _, s := range class.Signals {
		if s.Dimension == DimCapability {
			capEvidenceMap[s.Value] = append(capEvidenceMap[s.Value], s.Evidence...)
			if higherConfidence(s.Confidence, capConfidenceMap[s.Value]) {
				capConfidenceMap[s.Value] = s.Confidence
			}
		}
	}

	// Stable capability proposal generation
	sortedCaps := make([]string, 0, len(capEvidenceMap))
	for capName := range capEvidenceMap {
		sortedCaps = append(sortedCaps, capName)
	}
	sort.Strings(sortedCaps)

	capCounter := 1
	for _, capName := range sortedCaps {
		ev := uniqueStrings(capEvidenceMap[capName])
		conf := capConfidenceMap[capName]
		if conf == "" {
			conf = ConfidenceHigh
		}

		rationale := fmt.Sprintf("Inferred capability %q supported by %d repository evidence fact(s)", capName, len(ev))
		res.Capabilities = append(res.Capabilities, CapabilityProposal{
			ID:         fmt.Sprintf("prop-cap-%04d", capCounter),
			Capability: capName,
			Source:     "inferred",
			Confidence: conf,
			Evidence:   ev,
			Rationale:  rationale,
			Status:     "proposed",
		})
		capCounter++
	}

	// Generate Profile Candidates
	// 1. If any signals exist, propose core-software base profile
	var allSignalEvidence []string
	for _, s := range class.Signals {
		allSignalEvidence = append(allSignalEvidence, s.Evidence...)
	}
	allSignalEvidence = uniqueStrings(allSignalEvidence)

	profCounter := 1
	if len(allSignalEvidence) > 0 {
		baseDef := standardProfiles["core-software"]
		res.Profiles = append(res.Profiles, ProfileCandidate{
			ID:           fmt.Sprintf("prop-prof-%04d", profCounter),
			Profile:      "core-software",
			Confidence:   ConfidenceFactual,
			Capabilities: baseDef.Capabilities,
			Contracts:    baseDef.Contracts,
			Evidence:     allSignalEvidence,
			Rationale:    "Baseline engineering foundation profile for any managed software repository",
		})
		profCounter++
	}

	// 2. Map archetypes / capabilities to specialized profiles
	profileTriggers := []struct {
		profile    string
		triggerCap string
		triggerApp string
	}{
		{profile: "cli", triggerCap: "cli", triggerApp: "cli"},
		{profile: "web-application", triggerCap: "web-application", triggerApp: "web-application"},
		{profile: "desktop-gui", triggerCap: "desktop-gui", triggerApp: "desktop-gui"},
		{profile: "api-service", triggerCap: "api-service", triggerApp: "api-service"},
		{profile: "compiler", triggerCap: "compiler", triggerApp: "compiler"},
		{profile: "library", triggerCap: "library", triggerApp: "library"},
	}

	for _, pt := range profileTriggers {
		var matchedEvidence []string
		var bestConf FactConfidence

		// Check if capability or app type matched
		for _, s := range class.Signals {
			if (s.Dimension == DimCapability && s.Value == pt.triggerCap) ||
				(s.Dimension == DimAppType && s.Value == pt.triggerApp) {
				matchedEvidence = append(matchedEvidence, s.Evidence...)
				if higherConfidence(s.Confidence, bestConf) {
					bestConf = s.Confidence
				}
			}
		}

		matchedEvidence = uniqueStrings(matchedEvidence)
		if len(matchedEvidence) > 0 {
			def := standardProfiles[pt.profile]
			if bestConf == "" {
				bestConf = ConfidenceHigh
			}
			res.Profiles = append(res.Profiles, ProfileCandidate{
				ID:           fmt.Sprintf("prop-prof-%04d", profCounter),
				Profile:      pt.profile,
				Confidence:   bestConf,
				Capabilities: def.Capabilities,
				Contracts:    def.Contracts,
				Evidence:     matchedEvidence,
				Rationale:    fmt.Sprintf("Candidate profile %q suggested by %d piece(s) of evidence matching capability/app-type", pt.profile, len(matchedEvidence)),
			})
			profCounter++
		}
	}

	return res
}

func higherConfidence(a, b FactConfidence) bool {
	rank := func(c FactConfidence) int {
		switch c {
		case ConfidenceFactual:
			return 4
		case ConfidenceHigh:
			return 3
		case ConfidenceMedium:
			return 2
		case ConfidenceLow:
			return 1
		default:
			return 0
		}
	}
	return rank(a) > rank(b)
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range in {
		if s != "" && !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}
