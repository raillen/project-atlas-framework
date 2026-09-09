package adoption

import (
	"encoding/json"
	"testing"
)

func TestMapDocumentationNonAtlasLayout(t *testing.T) {
	facts := []ObservedFact{
		{ID: "fact-0001", Kind: FactDoc, Key: "doc:README.md", Source: "README.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0002", Kind: FactDoc, Key: "doc:docs/design/system.md", Source: "docs/design/system.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0003", Kind: FactDoc, Key: "doc:TESTING.md", Source: "TESTING.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0004", Kind: FactDoc, Key: "doc:SECURITY.md", Source: "SECURITY.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0005", Kind: FactDoc, Key: "doc:docs/cli-commands.md", Source: "docs/cli-commands.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0006", Kind: FactDoc, Key: "doc:docs/setup-guide.md", Source: "docs/setup-guide.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
	}

	class := Classify(facts)
	res := MapDocumentation(facts, class)

	if res.Version != 1 {
		t.Fatalf("expected version 1, got %d", res.Version)
	}

	candMap := make(map[string]MappingCandidate)
	for _, c := range res.Candidates {
		if err := c.Validate(); err != nil {
			t.Fatalf("invalid mapping candidate: %v", err)
		}
		candMap[c.ContractID] = c
	}

	// Verify README.md mapped to product.vision
	if pv, ok := candMap["product.vision"]; !ok {
		t.Errorf("expected product.vision mapping candidate")
	} else {
		if len(pv.Sources) == 0 || pv.Sources[0] != "README.md" {
			t.Errorf("expected README.md in product.vision sources, got %v", pv.Sources)
		}
	}

	// Verify docs/design/system.md mapped to architecture.system without requiring standard atlas filename
	if arch, ok := candMap["architecture.system"]; !ok {
		t.Errorf("expected architecture.system mapping candidate")
	} else {
		found := false
		for _, s := range arch.Sources {
			if s == "docs/design/system.md" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected docs/design/system.md in architecture.system sources, got %v", arch.Sources)
		}
	}

	// Verify TESTING.md mapped to testing.strategy
	if testStrat, ok := candMap["testing.strategy"]; !ok {
		t.Errorf("expected testing.strategy mapping candidate")
	} else {
		if len(testStrat.Sources) == 0 || testStrat.Sources[0] != "TESTING.md" {
			t.Errorf("expected TESTING.md in testing.strategy sources, got %v", testStrat.Sources)
		}
	}

	// Verify CandidateBindings
	bindMap := make(map[string]CandidateBinding)
	for _, b := range res.Bindings {
		bindMap[b.ContractID] = b
	}

	if b, ok := bindMap["architecture.system"]; !ok {
		t.Errorf("expected architecture.system candidate binding")
	} else {
		if b.Authority != "inferred-state" {
			t.Errorf("expected authority 'inferred-state', got %q", b.Authority)
		}
	}

	// Verify JSON serialization round-trip
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	var unmarshaled MappingResult
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
	if len(unmarshaled.Candidates) != len(res.Candidates) {
		t.Errorf("mismatched candidates count after unmarshal: %d vs %d", len(unmarshaled.Candidates), len(res.Candidates))
	}
}

func TestMapDocumentationEmptyDocs(t *testing.T) {
	facts := []ObservedFact{
		{ID: "fact-0001", Kind: FactFile, Key: "source:main.go", Source: "main.go", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
	}
	class := Classify(facts)
	res := MapDocumentation(facts, class)

	if len(res.Candidates) != 0 {
		t.Errorf("expected 0 candidates for facts without docs, got %d", len(res.Candidates))
	}
	if len(res.Bindings) != 0 {
		t.Errorf("expected 0 bindings for facts without docs, got %d", len(res.Bindings))
	}
}

func TestMappingCandidateValidation(t *testing.T) {
	tests := []struct {
		name    string
		cand    MappingCandidate
		wantErr bool
	}{
		{
			name: "valid",
			cand: MappingCandidate{
				ID: "map-001", ContractID: "architecture.system",
				Sources: []string{"docs/arch.md"}, Confidence: ConfidenceHigh,
				Evidence: []string{"f1"}, Authority: "inferred-state", Status: "candidate",
			},
		},
		{
			name: "missing id",
			cand: MappingCandidate{
				ContractID: "architecture.system", Sources: []string{"docs/arch.md"},
				Confidence: ConfidenceHigh, Evidence: []string{"f1"}, Authority: "inferred-state", Status: "candidate",
			},
			wantErr: true,
		},
		{
			name: "missing contract id",
			cand: MappingCandidate{
				ID: "map-001", Sources: []string{"docs/arch.md"},
				Confidence: ConfidenceHigh, Evidence: []string{"f1"}, Authority: "inferred-state", Status: "candidate",
			},
			wantErr: true,
		},
		{
			name: "empty sources",
			cand: MappingCandidate{
				ID: "map-001", ContractID: "architecture.system",
				Sources: []string{}, Confidence: ConfidenceHigh,
				Evidence: []string{"f1"}, Authority: "inferred-state", Status: "candidate",
			},
			wantErr: true,
		},
		{
			name: "missing evidence",
			cand: MappingCandidate{
				ID: "map-001", ContractID: "architecture.system",
				Sources: []string{"docs/arch.md"}, Confidence: ConfidenceHigh,
				Evidence: []string{}, Authority: "inferred-state", Status: "candidate",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			cand: MappingCandidate{
				ID: "map-001", ContractID: "architecture.system",
				Sources: []string{"docs/arch.md"}, Confidence: ConfidenceHigh,
				Evidence: []string{"f1"}, Authority: "inferred-state", Status: "verified",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cand.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
