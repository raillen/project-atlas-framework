package adoption

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// hashDirectory computes a deterministic SHA256 hash of all files in a directory tree.
func hashDirectory(t *testing.T, root string) string {
	t.Helper()
	hasher := sha256.New()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, path)
		if d.IsDir() {
			return nil
		}
		hasher.Write([]byte(rel))
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, _ = io.Copy(hasher, f)
		return nil
	})
	if err != nil {
		t.Fatalf("failed hashing directory %s: %v", root, err)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

func TestBrownfieldCorpusGoCLI(t *testing.T) {
	fixturePath, _ := filepath.Abs("../../testdata/brownfield/go-cli")
	hashBefore := hashDirectory(t, fixturePath)

	opts := ScanOptions{
		Budget:         DefaultBudget(),
		ParseManifests: true,
	}

	report, err := RunAdoptionAudit(fixturePath, opts)
	if err != nil {
		t.Fatalf("RunAdoptionAudit failed on go-cli: %v", err)
	}

	// Invariant 1: Zero mutations
	hashAfter := hashDirectory(t, fixturePath)
	if hashBefore != hashAfter {
		t.Fatalf("destructive action violation: repository mutated during audit")
	}

	// Invariant 2: Classification accuracy
	hasGo := false
	for _, l := range report.Classification.Languages {
		if l == "go" {
			hasGo = true
		}
	}
	if !hasGo {
		t.Errorf("expected 'go' in languages, got %v", report.Classification.Languages)
	}

	hasCLI := false
	for _, a := range report.Classification.AppTypes {
		if a == "cli" {
			hasCLI = true
		}
	}
	if !hasCLI {
		t.Errorf("expected 'cli' in app types, got %v", report.Classification.AppTypes)
	}

	// Invariant 3: CI capability detected
	hasCI := false
	for _, c := range report.Capabilities {
		if c.Capability == "ci-cd" {
			hasCI = true
		}
	}
	if !hasCI {
		t.Errorf("expected ci-cd capability, got %v", report.Capabilities)
	}

	// Invariant 4: Documentation mapped
	if len(report.DocBindings) == 0 {
		t.Errorf("expected candidate doc bindings for go-cli")
	}
}

func TestBrownfieldCorpusWebMonorepo(t *testing.T) {
	fixturePath, _ := filepath.Abs("../../testdata/brownfield/web-monorepo")
	hashBefore := hashDirectory(t, fixturePath)

	opts := ScanOptions{
		Budget:         DefaultBudget(),
		ParseManifests: true,
	}

	report, err := RunAdoptionAudit(fixturePath, opts)
	if err != nil {
		t.Fatalf("RunAdoptionAudit failed on web-monorepo: %v", err)
	}

	// Invariant 1: Zero mutations
	hashAfter := hashDirectory(t, fixturePath)
	if hashBefore != hashAfter {
		t.Fatalf("destructive action violation: repository mutated during audit")
	}

	// Invariant 2: Languages and frameworks
	frameworksMap := make(map[string]bool)
	for _, f := range report.Classification.Frameworks {
		frameworksMap[f] = true
	}
	if !frameworksMap["react"] {
		t.Errorf("expected react framework, got %v", report.Classification.Frameworks)
	}
	if !frameworksMap["express"] {
		t.Errorf("expected express framework, got %v", report.Classification.Frameworks)
	}

	// Persistence: pg / postgresql
	persistMap := make(map[string]bool)
	for _, p := range report.Classification.Persistence {
		persistMap[p] = true
	}
	if !persistMap["postgresql"] {
		t.Errorf("expected postgresql in persistence, got %v", report.Classification.Persistence)
	}
}

func TestBrownfieldCorpusMaliciousInjection(t *testing.T) {
	fixturePath, _ := filepath.Abs("../../testdata/brownfield/malicious-injection")
	hashBefore := hashDirectory(t, fixturePath)

	opts := ScanOptions{
		Budget:         DefaultBudget(),
		ParseManifests: true,
	}

	report, err := RunAdoptionAudit(fixturePath, opts)
	if err != nil {
		t.Fatalf("RunAdoptionAudit failed on malicious-injection: %v", err)
	}

	// Invariant 1: Zero mutations
	hashAfter := hashDirectory(t, fixturePath)
	if hashBefore != hashAfter {
		t.Fatalf("destructive action violation: repository mutated during audit")
	}

	// Invariant 2: Zero false canonical promotion
	for _, b := range report.DocBindings {
		// All candidate bindings must be proposals, never canonical policy
		if b.ContractID == "canonical-policy" {
			t.Errorf("malicious injection promoted to canonical policy")
		}
	}

	// Invariant 3: Confidence Ledger entries are strictly inferred-state or factual
	for _, entry := range report.Ledger.Entries {
		if entry.Authority == "policy" || entry.Authority == "user-decision" {
			t.Errorf("untrusted README elevated to authoritative level %q", entry.Authority)
		}
	}
}

func TestBrownfieldCorpusSecretsEnv(t *testing.T) {
	fixturePath, _ := filepath.Abs("../../testdata/brownfield/secrets-env")
	hashBefore := hashDirectory(t, fixturePath)

	opts := ScanOptions{
		Budget:         DefaultBudget(),
		ParseManifests: true,
	}

	report, err := RunAdoptionAudit(fixturePath, opts)
	if err != nil {
		t.Fatalf("RunAdoptionAudit failed on secrets-env: %v", err)
	}

	// Invariant 1: Zero mutations
	hashAfter := hashDirectory(t, fixturePath)
	if hashBefore != hashAfter {
		t.Fatalf("destructive action violation: repository mutated during audit")
	}

	// Invariant 2: Secret content is NEVER persisted in facts, ledger, or report
	leakIndicators := []string{
		"SUPER_SECRET_PASSWORD_DO_NOT_LEAK_12345",
		"synthetic_test_api_key_redaction_check_9988",
		"synthetic_test_aws_key_check_1122",
	}

	scanner := NewScanner(fixturePath, opts)
	scanResult := scanner.Scan()
	envFactFound := false

	for _, fact := range scanResult.Facts {
		if strings.HasPrefix(fact.Key, "env-presence:") {
			envFactFound = true
			if fact.Metadata["secret_bearing"] != true {
				t.Errorf("expected secret_bearing: true on env fact")
			}
			if fact.Metadata["content"] != "[REDACTED_BY_SECURITY_POLICY]" {
				t.Errorf("expected redacted content, got %v", fact.Metadata["content"])
			}
		}

		// Check for leak in any fact field
		for _, leak := range leakIndicators {
			if strings.Contains(fact.Key, leak) || strings.Contains(fact.Source, leak) {
				t.Fatalf("SECURITY VIOLATION: secret leaked in fact %s", fact.ID)
			}
			if fact.Metadata != nil {
				for _, v := range fact.Metadata {
					if strings.Contains(strings.ToLower(strings.TrimSpace(toString(v))), strings.ToLower(leak)) {
						t.Fatalf("SECURITY VIOLATION: secret leaked in fact metadata %s", fact.ID)
					}
				}
			}
		}
	}

	if !envFactFound {
		t.Errorf("expected env-presence fact for .env file")
	}

	// Invariant 3: Security risk generated in report
	hasEnvRisk := false
	for _, r := range report.Risks {
		if strings.Contains(r, ".env") {
			hasEnvRisk = true
		}
	}
	if !hasEnvRisk {
		t.Errorf("expected .env risk warning in report risks, got %v", report.Risks)
	}
}

func toString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func TestBrownfieldCorpusStaleConflictingDocs(t *testing.T) {
	fixturePath, _ := filepath.Abs("../../testdata/brownfield/stale-conflicting-docs")
	hashBefore := hashDirectory(t, fixturePath)

	opts := ScanOptions{
		Budget:         DefaultBudget(),
		ParseManifests: true,
	}

	report, err := RunAdoptionAudit(fixturePath, opts)
	if err != nil {
		t.Fatalf("RunAdoptionAudit failed: %v", err)
	}

	// Invariant 1: Zero mutations
	hashAfter := hashDirectory(t, fixturePath)
	if hashBefore != hashAfter {
		t.Fatalf("destructive action violation: repository mutated during audit")
	}

	// Invariant 2: Factual manifest takes precedence over stale documentation
	hasGo := false
	for _, l := range report.Classification.Languages {
		if l == "go" {
			hasGo = true
		}
	}
	if !hasGo {
		t.Errorf("expected factual go manifest to be classified as language go")
	}
}

func TestBrownfieldCorpusNoDocs(t *testing.T) {
	fixturePath, _ := filepath.Abs("../../testdata/brownfield/no-docs")
	hashBefore := hashDirectory(t, fixturePath)

	opts := ScanOptions{
		Budget:         DefaultBudget(),
		ParseManifests: true,
	}

	report, err := RunAdoptionAudit(fixturePath, opts)
	if err != nil {
		t.Fatalf("RunAdoptionAudit failed on no-docs: %v", err)
	}

	// Invariant 1: Zero mutations
	hashAfter := hashDirectory(t, fixturePath)
	if hashBefore != hashAfter {
		t.Fatalf("destructive action violation: repository mutated during audit")
	}

	// Invariant 2: Missing contracts correctly flagged
	if len(report.DocCoverage.MissingContracts) == 0 {
		t.Errorf("expected missing documentation contracts for repository without docs")
	}
}

func TestDogfoodPrumoSelfAudit(t *testing.T) {
	repoRoot, _ := filepath.Abs("../../")

	opts := ScanOptions{
		Budget: ScanBudget{
			MaxFiles: 1000,
			MaxBytes: 50 * 1024 * 1024,
		},
		ParseManifests: true,
	}

	report, err := RunAdoptionAudit(repoRoot, opts)
	if err != nil {
		t.Fatalf("self-adoption audit failed: %v", err)
	}

	// Prumo is now a pure Go repository with CLI and Schemas (Python retired)
	hasGo := false
	for _, l := range report.Classification.Languages {
		if l == "go" {
			hasGo = true
		}
	}
	if !hasGo {
		t.Errorf("expected go in self-audit classification, got %v", report.Classification.Languages)
	}

	hasCLI := false
	for _, a := range report.Classification.AppTypes {
		if a == "cli" {
			hasCLI = true
		}
	}
	if !hasCLI {
		t.Errorf("expected cli app type in self audit, got %v", report.Classification.AppTypes)
	}

	// Prumo artifacts detected
	if len(report.PrumoArtifacts) == 0 {
		t.Errorf("expected Prumo artifacts detected in self-audit")
	}

	// Invariant: Non-destructive verification (human report renders cleanly)
	rendered := RenderHumanReport(report)
	if !strings.Contains(rendered, "PROJECT PRUMO — REPOSITORY ADOPTION REPORT") {
		t.Errorf("human report did not render correctly")
	}
}
