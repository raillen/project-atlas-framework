package adoption

import "testing"

func TestObservedFactValidation(t *testing.T) {
	tests := []struct {
		name    string
		fact    ObservedFact
		wantErr bool
	}{
		{
			name: "valid factual file fact",
			fact: ObservedFact{
				ID: "fact-0001", Kind: FactFile, Key: "source:main.go",
				Source: "main.go", Extraction: ExtractionStaticPresence,
				Confidence: ConfidenceFactual,
			},
		},
		{
			name: "valid manifest fact",
			fact: ObservedFact{
				ID: "fact-0002", Kind: FactManifest, Key: "manifest:go.mod",
				Source: "go.mod", Extraction: ExtractionFilenamePattern,
				Confidence: ConfidenceFactual,
			},
		},
		{
			name:    "missing id",
			fact:    ObservedFact{Kind: FactFile, Key: "x", Source: "x", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
			wantErr: true,
		},
		{
			name:    "whitespace-only id",
			fact:    ObservedFact{ID: "  ", Kind: FactFile, Key: "x", Source: "x", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
			wantErr: true,
		},
		{
			name:    "invalid kind",
			fact:    ObservedFact{ID: "f1", Kind: "bogus", Key: "x", Source: "x", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
			wantErr: true,
		},
		{
			name:    "missing key",
			fact:    ObservedFact{ID: "f1", Kind: FactFile, Source: "x", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
			wantErr: true,
		},
		{
			name:    "missing source",
			fact:    ObservedFact{ID: "f1", Kind: FactFile, Key: "x", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
			wantErr: true,
		},
		{
			name:    "invalid extraction",
			fact:    ObservedFact{ID: "f1", Kind: FactFile, Key: "x", Source: "x", Extraction: "magic", Confidence: ConfidenceFactual},
			wantErr: true,
		},
		{
			name:    "invalid confidence",
			fact:    ObservedFact{ID: "f1", Kind: FactFile, Key: "x", Source: "x", Extraction: ExtractionStaticPresence, Confidence: "maybe"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fact.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestFactKindValid(t *testing.T) {
	valid := []FactKind{
		FactFile, FactManifest, FactConfig, FactDoc, FactCI, FactTest,
		FactSchema, FactADR, FactAgentRules, FactGitMetadata, FactOther,
	}
	for _, k := range valid {
		if !k.Valid() {
			t.Errorf("expected %q to be valid", k)
		}
	}
	if FactKind("unknown").Valid() {
		t.Fatal("unknown kind should be invalid")
	}
}

func TestExtractionValid(t *testing.T) {
	valid := []Extraction{
		ExtractionStaticPresence, ExtractionFilenamePattern,
		ExtractionContentParse, ExtractionManifestField,
		ExtractionGitCommand, ExtractionDirectoryStructure,
	}
	for _, e := range valid {
		if !e.Valid() {
			t.Errorf("expected %q to be valid", e)
		}
	}
	if Extraction("guessing").Valid() {
		t.Fatal("unknown extraction should be invalid")
	}
}

func TestFactConfidenceValid(t *testing.T) {
	valid := []FactConfidence{ConfidenceFactual, ConfidenceHigh, ConfidenceMedium, ConfidenceLow}
	for _, c := range valid {
		if !c.Valid() {
			t.Errorf("expected %q to be valid", c)
		}
	}
	if FactConfidence("uncertain").Valid() {
		t.Fatal("unknown confidence should be invalid")
	}
}
