package adoption

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRunAdoptionAudit(t *testing.T) {
	root := t.TempDir()

	writeTestFile(t, root, "go.mod", "module example.com/myservice\n\ngo 1.22\n")
	writeTestFile(t, root, "cmd/server/main.go", "package main\n\nfunc main() {}\n")
	writeTestFile(t, root, "README.md", "# MyService\n\nA demonstration microservice.\n")
	writeTestFile(t, root, "docs/design.md", "# Architecture\n\nSystem architecture overview.\n")
	writeTestFile(t, root, "Dockerfile", "FROM golang:1.22\n")
	mkdirTest(t, root, ".github/workflows")
	writeTestFile(t, root, ".github/workflows/ci.yml", "name: CI\non: push\n")
	mkdirTest(t, root, "tests")
	writeTestFile(t, root, "tests/service_test.go", "package tests\n")

	opts := ScanOptions{
		Budget:         DefaultBudget(),
		ParseManifests: true,
	}

	report, err := RunAdoptionAudit(root, opts)
	if err != nil {
		t.Fatalf("RunAdoptionAudit failed: %v", err)
	}

	if report.Version != 1 {
		t.Errorf("expected version 1, got %d", report.Version)
	}
	if report.Repository.Root != root {
		t.Errorf("expected repository root %s, got %s", root, report.Repository.Root)
	}
	if report.FactsSummary.TotalFacts == 0 {
		t.Errorf("expected observed facts in summary, got 0")
	}

	// Verify classification
	hasGo := false
	for _, l := range report.Classification.Languages {
		if l == "go" {
			hasGo = true
		}
	}
	if !hasGo {
		t.Errorf("expected 'go' in classification languages")
	}

	// Verify profiles include core-software and cli
	profMap := make(map[string]bool)
	for _, p := range report.Profiles {
		profMap[p.Profile] = true
	}
	if !profMap["core-software"] || !profMap["cli"] {
		t.Errorf("expected core-software and cli profiles, got %v", report.Profiles)
	}

	// Verify doc coverage has bound contracts
	if len(report.DocCoverage.BoundContracts) == 0 {
		t.Errorf("expected bound contracts in doc coverage")
	}

	// Verify Confidence Ledger is populated
	if report.Ledger.Summary.Total == 0 {
		t.Errorf("expected confidence ledger entries in report")
	}

	// Verify human readable rendering
	human := RenderHumanReport(report)
	if !strings.Contains(human, "REPOSITORY ADOPTION REPORT") {
		t.Errorf("human report missing header")
	}
	if !strings.Contains(human, "1. Classification") {
		t.Errorf("human report missing classification section")
	}
	if !strings.Contains(human, "2. Recommended Profiles & Capabilities") {
		t.Errorf("human report missing profiles section")
	}
	if !strings.Contains(human, "3. Documentation Mapping & Coverage") {
		t.Errorf("human report missing doc mapping section")
	}
	if !strings.Contains(human, "4. Confidence Ledger Summary") {
		t.Errorf("human report missing ledger section")
	}

	// Verify JSON serialization round-trip
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("failed to marshal AdoptionReport: %v", err)
	}
	var unmarshaled AdoptionReport
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal AdoptionReport: %v", err)
	}
	if unmarshaled.FactsSummary.TotalFacts != report.FactsSummary.TotalFacts {
		t.Errorf("mismatched facts count after unmarshal: %d vs %d", unmarshaled.FactsSummary.TotalFacts, report.FactsSummary.TotalFacts)
	}
}
