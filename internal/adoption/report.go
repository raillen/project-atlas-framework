package adoption

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FactsSummary aggregates counts of discovered repository evidence.
type FactsSummary struct {
	TotalFacts      int            `json:"total_facts"`
	FilesScanned    int            `json:"files_scanned"`
	BytesScanned    int64          `json:"bytes_scanned"`
	BudgetExhausted bool           `json:"budget_exhausted"`
	ByKind          map[string]int `json:"by_kind"`
}

// DocCoverageSummary compares required contracts from candidate profiles with discovered bindings.
type DocCoverageSummary struct {
	RequiredContracts []string `json:"required_contracts"`
	BoundContracts    []string `json:"bound_contracts"`
	MissingContracts  []string `json:"missing_contracts"`
}

// RepoSummary summarizes the target repository location and git state.
type RepoSummary struct {
	Root     string `json:"root"`
	Branch   string `json:"branch,omitempty"`
	Revision string `json:"revision,omitempty"`
	Dirty    bool   `json:"dirty"`
}

// AdoptionReport is the canonical deliverable of the Adoption Engine discovery pass.
// It clearly separates facts, inferences, proposals, risks, and recommended actions.
type AdoptionReport struct {
	Version            int                      `json:"version"`
	Repository         RepoSummary              `json:"repository"`
	FactsSummary       FactsSummary             `json:"facts_summary"`
	Classification     RepositoryClassification `json:"classification"`
	Profiles           []ProfileCandidate       `json:"profiles"`
	Capabilities       []CapabilityProposal     `json:"capabilities"`
	AtlasArtifacts     []string                 `json:"atlas_artifacts"`
	DocBindings        []CandidateBinding       `json:"doc_bindings"`
	DocCoverage        DocCoverageSummary       `json:"doc_coverage"`
	Contradictions     []string                 `json:"contradictions"`
	Ledger             ConfidenceLedger         `json:"ledger"`
	MigrationProposals []string                 `json:"migration_proposals,omitempty"`
	Risks              []string                 `json:"risks"`
	NextSteps          []string                 `json:"next_steps"`
	GeneratedAt        string                   `json:"generated_at"`
}

// RunAdoptionAudit runs the full adoption discovery and inference pipeline without mutating disk.
func RunAdoptionAudit(root string, opts ScanOptions) (AdoptionReport, error) {
	// 1. Scan facts
	opts.ParseManifests = true
	scanner := NewScanner(root, opts)
	scanResult := scanner.Scan()

	// 2. Classify
	classification := Classify(scanResult.Facts)

	// 3. Propose capabilities & profiles
	proposals := Propose(classification)

	// 4. Map documentation
	mappingResult := MapDocumentation(scanResult.Facts, classification)

	// 5. Confidence Ledger
	ledger := BuildConfidenceLedger(classification, proposals, mappingResult)

	// 6. Facts summary
	byKind := make(map[string]int)
	for _, f := range scanResult.Facts {
		byKind[string(f.Kind)]++
	}

	factsSum := FactsSummary{
		TotalFacts:      len(scanResult.Facts),
		FilesScanned:    scanResult.Budget.FilesScanned,
		BytesScanned:    scanResult.Budget.BytesScanned,
		BudgetExhausted: scanResult.Budget.Exhausted,
		ByKind:          byKind,
	}

	// 7. Check existing Atlas artifacts
	var atlasArtifacts []string
	candidates := []string{"atlas.json", "AGENTS.md", ".cursorrules", ".clinerules", "copilot-instructions.md"}
	for _, cand := range candidates {
		if _, err := os.Stat(filepath.Join(root, cand)); err == nil {
			atlasArtifacts = append(atlasArtifacts, cand)
		}
	}
	if info, err := os.Stat(filepath.Join(root, ".atlas")); err == nil && info.IsDir() {
		atlasArtifacts = append(atlasArtifacts, ".atlas/")
	}
	sort.Strings(atlasArtifacts)

	// 8. Doc coverage
	requiredContractsMap := make(map[string]bool)
	for _, p := range proposals.Profiles {
		for _, c := range p.Contracts {
			requiredContractsMap[c] = true
		}
	}
	requiredContracts := make([]string, 0, len(requiredContractsMap))
	for c := range requiredContractsMap {
		requiredContracts = append(requiredContracts, c)
	}
	sort.Strings(requiredContracts)

	boundContractsMap := make(map[string]bool)
	for _, b := range mappingResult.Bindings {
		boundContractsMap[b.ContractID] = true
	}
	boundContracts := make([]string, 0, len(boundContractsMap))
	for c := range boundContractsMap {
		boundContracts = append(boundContracts, c)
	}
	sort.Strings(boundContracts)

	var missingContracts []string
	for _, c := range requiredContracts {
		if !boundContractsMap[c] {
			missingContracts = append(missingContracts, c)
		}
	}

	docCoverage := DocCoverageSummary{
		RequiredContracts: requiredContracts,
		BoundContracts:    boundContracts,
		MissingContracts:  missingContracts,
	}

	// 9. Contradictions
	var contradictions []string
	if len(classification.AppTypes) > 2 {
		contradictions = append(contradictions, fmt.Sprintf("Multiple divergent project archetypes inferred: %s", strings.Join(classification.AppTypes, ", ")))
	}

	// 10. Risks
	var risks []string
	if scanResult.Dirty {
		risks = append(risks, "Repository contains uncommitted working tree changes; audit was performed on a dirty state")
	}
	if scanResult.Budget.Exhausted {
		risks = append(risks, "Scanner budget was exhausted; some files were not evaluated")
	}
	if len(missingContracts) > 0 {
		risks = append(risks, fmt.Sprintf("%d required documentation contract(s) have no existing source files", len(missingContracts)))
	}
	if ledger.Summary.RequiresConfirmation > 0 {
		risks = append(risks, fmt.Sprintf("%d inference(s) require confirmation before promotion to canonical", ledger.Summary.RequiresConfirmation))
	}

	// 11. Next steps
	var nextSteps []string
	if ledger.Summary.RequiresConfirmation > 0 {
		nextSteps = append(nextSteps, "Run 'atlas adopt --interactive' to review and confirm open questions and proposals")
	}
	if len(missingContracts) > 0 {
		nextSteps = append(nextSteps, fmt.Sprintf("Author documentation for missing contract(s): %s", strings.Join(missingContracts, ", ")))
	}
	nextSteps = append(nextSteps, "Run 'atlas adopt' without --audit-only to apply accepted configuration")

	report := AdoptionReport{
		Version: 1,
		Repository: RepoSummary{
			Root:     root,
			Branch:   scanResult.Branch,
			Revision: scanResult.Revision,
			Dirty:    scanResult.Dirty,
		},
		FactsSummary:   factsSum,
		Classification: classification,
		Profiles:       proposals.Profiles,
		Capabilities:   proposals.Capabilities,
		AtlasArtifacts: atlasArtifacts,
		DocBindings:    mappingResult.Bindings,
		DocCoverage:    docCoverage,
		Contradictions: contradictions,
		Ledger:         ledger,
		Risks:          risks,
		NextSteps:      nextSteps,
		GeneratedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	migProps := GenerateMigrationProposals(report)
	var propSummaries []string
	for _, mp := range migProps {
		propSummaries = append(propSummaries, fmt.Sprintf("%s: %s", mp.ID, mp.Title))
	}
	if propSummaries == nil {
		propSummaries = []string{}
	}
	report.MigrationProposals = propSummaries

	return report, nil
}

// RenderHumanReport formats an AdoptionReport into a clean, human-readable terminal output.
func RenderHumanReport(r AdoptionReport) string {
	var b strings.Builder

	b.WriteString("================================================================================\n")
	b.WriteString("                    PROJECT ATLAS — REPOSITORY ADOPTION REPORT                 \n")
	b.WriteString("================================================================================\n\n")

	// 1. Repository & Facts
	fmt.Fprintf(&b, "Repository: %s\n", r.Repository.Root)
	if r.Repository.Branch != "" {
		dirty := ""
		if r.Repository.Dirty {
			dirty = " (DIRTY)"
		}
		fmt.Fprintf(&b, "Git State:  %s @ %s%s\n", r.Repository.Branch, r.Repository.Revision, dirty)
	}
	fmt.Fprintf(&b, "Facts:      %d observed facts across %d file(s) (%.1f KB scanned)\n\n",
		r.FactsSummary.TotalFacts, r.FactsSummary.FilesScanned, float64(r.FactsSummary.BytesScanned)/1024.0)

	// 2. Classification
	b.WriteString("--- 1. Classification ---------------------------------------------------------\n")
	if len(r.Classification.Languages) > 0 {
		fmt.Fprintf(&b, "Languages:    %s\n", strings.Join(r.Classification.Languages, ", "))
	}
	if len(r.Classification.Toolchains) > 0 {
		fmt.Fprintf(&b, "Toolchains:   %s\n", strings.Join(r.Classification.Toolchains, ", "))
	}
	if len(r.Classification.AppTypes) > 0 {
		fmt.Fprintf(&b, "Archetypes:   %s\n", strings.Join(r.Classification.AppTypes, ", "))
	}
	if len(r.Classification.Frameworks) > 0 {
		fmt.Fprintf(&b, "Frameworks:   %s\n", strings.Join(r.Classification.Frameworks, ", "))
	}
	if len(r.Classification.Persistence) > 0 {
		fmt.Fprintf(&b, "Persistence:  %s\n", strings.Join(r.Classification.Persistence, ", "))
	}
	if len(r.Classification.Testing) > 0 {
		fmt.Fprintf(&b, "Testing:      %s\n", strings.Join(r.Classification.Testing, ", "))
	}
	if len(r.Classification.Deployment) > 0 {
		fmt.Fprintf(&b, "Deployment:   %s\n", strings.Join(r.Classification.Deployment, ", "))
	}
	b.WriteString("\n")

	// 3. Candidate Profiles & Capabilities
	b.WriteString("--- 2. Recommended Profiles & Capabilities ------------------------------------\n")
	for _, p := range r.Profiles {
		fmt.Fprintf(&b, "• Profile: %s (confidence: %s)\n", p.Profile, p.Confidence)
		if p.Rationale != "" {
			fmt.Fprintf(&b, "  Rationale: %s\n", p.Rationale)
		}
	}
	if len(r.Capabilities) > 0 {
		var caps []string
		for _, c := range r.Capabilities {
			caps = append(caps, c.Capability)
		}
		fmt.Fprintf(&b, "• Inferred Capabilities: %s\n", strings.Join(caps, ", "))
	}
	b.WriteString("\n")

	// 4. Documentation Mapping & Coverage
	b.WriteString("--- 3. Documentation Mapping & Coverage ---------------------------------------\n")
	for _, bnd := range r.DocBindings {
		fmt.Fprintf(&b, "• Contract %-22s <- %s\n", "["+bnd.ContractID+"]", strings.Join(bnd.Sources, ", "))
	}
	if len(r.DocCoverage.MissingContracts) > 0 {
		b.WriteString("\nMissing Contracts (required by recommended profiles):\n")
		for _, mc := range r.DocCoverage.MissingContracts {
			fmt.Fprintf(&b, "  [!] %s (no source file mapped)\n", mc)
		}
	}
	b.WriteString("\n")

	// 5. Confidence Ledger Summary
	b.WriteString("--- 4. Confidence Ledger Summary ----------------------------------------------\n")
	fmt.Fprintf(&b, "Total Inferences: %d | Factual: %d | High: %d | Medium: %d | Low: %d\n",
		r.Ledger.Summary.Total, r.Ledger.Summary.Factual, r.Ledger.Summary.High, r.Ledger.Summary.Medium, r.Ledger.Summary.Low)
	fmt.Fprintf(&b, "Requiring Confirmation: %d\n", r.Ledger.Summary.RequiresConfirmation)
	for _, e := range r.Ledger.RequiringConfirmation() {
		fmt.Fprintf(&b, "  ? [%s] %s (confidence: %s, score: %.2f)\n", e.Category, e.Claim, e.Confidence, e.Score)
	}
	b.WriteString("\n")

	// 6. Risks
	if len(r.Risks) > 0 {
		b.WriteString("--- 5. Risks & Warnings -------------------------------------------------------\n")
		for _, risk := range r.Risks {
			fmt.Fprintf(&b, "  ▲ %s\n", risk)
		}
		b.WriteString("\n")
	}

	// 7. Migration Proposals
	if len(r.MigrationProposals) > 0 {
		b.WriteString("--- 6. Migration Proposals (Review Governed) ----------------------------------\n")
		for _, mp := range r.MigrationProposals {
			fmt.Fprintf(&b, "  → %s\n", mp)
		}
		b.WriteString("\n")
	}

	// 8. Recommended Next Steps
	b.WriteString("--- 7. Recommended Next Steps -------------------------------------------------\n")
	for i, step := range r.NextSteps {
		fmt.Fprintf(&b, "  %d. %s\n", i+1, step)
	}
	b.WriteString("================================================================================\n")

	return b.String()
}
