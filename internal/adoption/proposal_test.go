package adoption

import (
	"encoding/json"
	"testing"
)

func TestProposeGoCLI(t *testing.T) {
	facts := []ObservedFact{
		{ID: "fact-0001", Kind: FactManifest, Key: "manifest:go.mod", Source: "go.mod", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0002", Kind: FactFile, Key: "source:cmd/atlas/main.go", Source: "cmd/atlas/main.go", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
		{ID: "fact-0003", Kind: FactSchema, Key: "schema:atlas.schema.json", Source: "schemas/atlas.schema.json", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
	}

	class := Classify(facts)
	res := Propose(class)

	if res.Version != 1 {
		t.Fatalf("expected version 1, got %d", res.Version)
	}

	// Verify all capability proposals are valid
	capMap := make(map[string]CapabilityProposal)
	for _, cp := range res.Capabilities {
		if err := cp.Validate(); err != nil {
			t.Fatalf("invalid capability proposal: %v", err)
		}
		capMap[cp.Capability] = cp
	}

	if _, ok := capMap["cli"]; !ok {
		t.Errorf("expected capability 'cli' proposal, got %v", res.Capabilities)
	}
	if _, ok := capMap["go"]; !ok {
		t.Errorf("expected capability 'go' proposal, got %v", res.Capabilities)
	}
	if _, ok := capMap["schema"]; !ok {
		t.Errorf("expected capability 'schema' proposal, got %v", res.Capabilities)
	}

	// Verify profile candidates
	profMap := make(map[string]ProfileCandidate)
	for _, prof := range res.Profiles {
		if err := prof.Validate(); err != nil {
			t.Fatalf("invalid profile candidate: %v", err)
		}
		profMap[prof.Profile] = prof
	}

	// Base profile core-software must be present
	if base, ok := profMap["core-software"]; !ok {
		t.Errorf("expected core-software base profile candidate")
	} else {
		if len(base.Contracts) == 0 {
			t.Errorf("expected contracts in core-software profile candidate")
		}
	}

	// Specialized profile 'cli' must be present
	if cliProf, ok := profMap["cli"]; !ok {
		t.Errorf("expected 'cli' profile candidate")
	} else {
		if len(cliProf.Evidence) == 0 {
			t.Errorf("expected evidence in cli profile candidate")
		}
	}

	// JSON serialization round-trip
	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("failed to marshal ProposalsResult: %v", err)
	}
	var unmarshaled ProposalsResult
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ProposalsResult: %v", err)
	}
	if len(unmarshaled.Capabilities) != len(res.Capabilities) {
		t.Errorf("mismatched capabilities count after unmarshal: %d vs %d", len(unmarshaled.Capabilities), len(res.Capabilities))
	}
}

func TestProposeWebAppAndAPI(t *testing.T) {
	class := RepositoryClassification{
		Version: 1,
		Signals: []ClassificationSignal{
			{Dimension: DimCapability, Value: "web-application", Confidence: ConfidenceHigh, Evidence: []string{"fact-0001"}},
			{Dimension: DimCapability, Value: "api-service", Confidence: ConfidenceHigh, Evidence: []string{"fact-0002"}},
			{Dimension: DimAppType, Value: "web-application", Confidence: ConfidenceHigh, Evidence: []string{"fact-0001"}},
		},
	}

	res := Propose(class)

	profMap := make(map[string]ProfileCandidate)
	for _, p := range res.Profiles {
		profMap[p.Profile] = p
	}

	if _, ok := profMap["web-application"]; !ok {
		t.Errorf("expected web-application profile candidate")
	}
	if _, ok := profMap["api-service"]; !ok {
		t.Errorf("expected api-service profile candidate")
	}
	if _, ok := profMap["core-software"]; !ok {
		t.Errorf("expected core-software profile candidate")
	}
}

func TestProposeEmptyClassification(t *testing.T) {
	class := RepositoryClassification{Version: 1}
	res := Propose(class)

	if len(res.Capabilities) != 0 {
		t.Errorf("expected 0 capabilities for empty classification, got %d", len(res.Capabilities))
	}
	if len(res.Profiles) != 0 {
		t.Errorf("expected 0 profiles for empty classification, got %d", len(res.Profiles))
	}
}

func TestCapabilityProposalValidation(t *testing.T) {
	tests := []struct {
		name    string
		prop    CapabilityProposal
		wantErr bool
	}{
		{
			name: "valid",
			prop: CapabilityProposal{
				ID: "prop-001", Capability: "cli", Confidence: ConfidenceHigh,
				Evidence: []string{"f1"}, Status: "proposed",
			},
		},
		{
			name: "missing id",
			prop: CapabilityProposal{
				Capability: "cli", Confidence: ConfidenceHigh,
				Evidence: []string{"f1"}, Status: "proposed",
			},
			wantErr: true,
		},
		{
			name: "missing capability",
			prop: CapabilityProposal{
				ID: "prop-001", Confidence: ConfidenceHigh,
				Evidence: []string{"f1"}, Status: "proposed",
			},
			wantErr: true,
		},
		{
			name: "invalid confidence",
			prop: CapabilityProposal{
				ID: "prop-001", Capability: "cli", Confidence: "unknown",
				Evidence: []string{"f1"}, Status: "proposed",
			},
			wantErr: true,
		},
		{
			name: "missing evidence",
			prop: CapabilityProposal{
				ID: "prop-001", Capability: "cli", Confidence: ConfidenceHigh,
				Evidence: []string{}, Status: "proposed",
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			prop: CapabilityProposal{
				ID: "prop-001", Capability: "cli", Confidence: ConfidenceHigh,
				Evidence: []string{"f1"}, Status: "active",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prop.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestProfileCandidateValidation(t *testing.T) {
	tests := []struct {
		name    string
		prof    ProfileCandidate
		wantErr bool
	}{
		{
			name: "valid",
			prof: ProfileCandidate{
				ID: "prof-001", Profile: "cli", Confidence: ConfidenceHigh,
				Capabilities: []string{"cli"}, Evidence: []string{"f1"},
			},
		},
		{
			name: "missing id",
			prof: ProfileCandidate{
				Profile: "cli", Confidence: ConfidenceHigh,
				Capabilities: []string{"cli"}, Evidence: []string{"f1"},
			},
			wantErr: true,
		},
		{
			name: "missing profile",
			prof: ProfileCandidate{
				ID: "prof-001", Confidence: ConfidenceHigh,
				Capabilities: []string{"cli"}, Evidence: []string{"f1"},
			},
			wantErr: true,
		},
		{
			name: "missing evidence",
			prof: ProfileCandidate{
				ID: "prof-001", Profile: "cli", Confidence: ConfidenceHigh,
				Capabilities: []string{"cli"}, Evidence: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.prof.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}
