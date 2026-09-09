package adoption

import (
	"encoding/json"
	"testing"
)

func TestBuildConfidenceLedger(t *testing.T) {
	facts := []ObservedFact{
		{ID: "fact-0001", Kind: FactManifest, Key: "manifest:go.mod", Source: "go.mod", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0002", Kind: FactFile, Key: "source:cmd/atlas/main.go", Source: "cmd/atlas/main.go", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
		{ID: "fact-0003", Kind: FactDoc, Key: "doc:README.md", Source: "README.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0004", Kind: FactDoc, Key: "doc:docs/design.md", Source: "docs/design.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
	}

	class := Classify(facts)
	prop := Propose(class)
	mapRes := MapDocumentation(facts, class)

	ledger := BuildConfidenceLedger(class, prop, mapRes)

	if ledger.Version != 1 {
		t.Fatalf("expected version 1, got %d", ledger.Version)
	}

	if len(ledger.Entries) == 0 {
		t.Fatal("expected ledger entries, got none")
	}

	// Verify all entries pass validation
	for _, e := range ledger.Entries {
		if err := e.Validate(); err != nil {
			t.Errorf("invalid ledger entry %s: %v", e.ID, err)
		}
	}

	// Verify summary matches counts
	if ledger.Summary.Total != len(ledger.Entries) {
		t.Errorf("expected summary total %d, got %d", len(ledger.Entries), ledger.Summary.Total)
	}
	if ledger.Summary.Unresolved != len(ledger.Entries) {
		t.Errorf("expected all entries unresolved initially")
	}

	// Verify categories are represented
	catMap := make(map[LedgerCategory]bool)
	for _, e := range ledger.Entries {
		catMap[e.Category] = true
	}

	if !catMap[CategoryClassification] {
		t.Errorf("expected classification category entries in ledger")
	}
	if !catMap[CategoryCapability] {
		t.Errorf("expected capability category entries in ledger")
	}
	if !catMap[CategoryProfile] {
		t.Errorf("expected profile category entries in ledger")
	}
	if !catMap[CategoryMapping] {
		t.Errorf("expected mapping category entries in ledger")
	}

	// Test UpdateStatus transition
	firstID := ledger.Entries[0].ID
	if err := ledger.UpdateStatus(firstID, LedgerStatusAccepted); err != nil {
		t.Fatalf("failed to update status: %v", err)
	}
	if ledger.Entries[0].Status != LedgerStatusAccepted {
		t.Errorf("expected status accepted, got %q", ledger.Entries[0].Status)
	}
	if ledger.Summary.Unresolved != len(ledger.Entries)-1 {
		t.Errorf("expected unresolved count to decrease by 1")
	}

	// JSON round-trip
	data, err := json.Marshal(ledger)
	if err != nil {
		t.Fatalf("failed to marshal ledger: %v", err)
	}
	var unmarshaled ConfidenceLedger
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal ledger: %v", err)
	}
	if len(unmarshaled.Entries) != len(ledger.Entries) {
		t.Errorf("entry count mismatch after round-trip: %d vs %d", len(unmarshaled.Entries), len(ledger.Entries))
	}
}

func TestLedgerEntryValidation(t *testing.T) {
	tests := []struct {
		name    string
		entry   LedgerEntry
		wantErr bool
	}{
		{
			name: "valid high confidence",
			entry: LedgerEntry{
				ID: "cld-001", Claim: "Archetype is cli", Category: CategoryClassification,
				Confidence: ConfidenceHigh, Score: 0.85, Evidence: []string{"f1"},
				Status: LedgerStatusUnresolved, Authority: "inferred-state",
			},
		},
		{
			name: "valid low confidence with confirmation",
			entry: LedgerEntry{
				ID: "cld-002", Claim: "May be library", Category: CategoryClassification,
				Confidence: ConfidenceLow, Score: 0.30, Evidence: []string{"f1"},
				Status: LedgerStatusUnresolved, Authority: "inferred-state",
				RequiresConfirmation: true,
			},
		},
		{
			name: "invalid low confidence without confirmation",
			entry: LedgerEntry{
				ID: "cld-003", Claim: "May be library", Category: CategoryClassification,
				Confidence: ConfidenceLow, Score: 0.30, Evidence: []string{"f1"},
				Status: LedgerStatusUnresolved, Authority: "inferred-state",
				RequiresConfirmation: false,
			},
			wantErr: true,
		},
		{
			name: "missing id",
			entry: LedgerEntry{
				Claim: "Claim", Category: CategoryClassification, Confidence: ConfidenceHigh,
				Score: 0.85, Evidence: []string{"f1"}, Status: LedgerStatusUnresolved, Authority: "inferred-state",
			},
			wantErr: true,
		},
		{
			name: "score out of bounds",
			entry: LedgerEntry{
				ID: "cld-004", Claim: "Claim", Category: CategoryClassification,
				Confidence: ConfidenceHigh, Score: 1.5, Evidence: []string{"f1"},
				Status: LedgerStatusUnresolved, Authority: "inferred-state",
			},
			wantErr: true,
		},
		{
			name: "missing evidence",
			entry: LedgerEntry{
				ID: "cld-005", Claim: "Claim", Category: CategoryClassification,
				Confidence: ConfidenceHigh, Score: 0.85, Evidence: []string{},
				Status: LedgerStatusUnresolved, Authority: "inferred-state",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestLedgerFiltering(t *testing.T) {
	ledger := ConfidenceLedger{
		Version: 1,
		Entries: []LedgerEntry{
			{
				ID: "cld-001", Claim: "Claim 1", Category: CategoryClassification,
				Confidence: ConfidenceHigh, Score: 0.85, Evidence: []string{"f1"},
				Status: LedgerStatusUnresolved, Authority: "inferred-state", RequiresConfirmation: false,
			},
			{
				ID: "cld-002", Claim: "Claim 2", Category: CategoryClassification,
				Confidence: ConfidenceLow, Score: 0.30, Evidence: []string{"f2"},
				Status: LedgerStatusUnresolved, Authority: "inferred-state", RequiresConfirmation: true,
			},
			{
				ID: "cld-003", Claim: "Claim 3", Category: CategoryCapability,
				Confidence: ConfidenceHigh, Score: 0.85, Evidence: []string{"f3"},
				Status: LedgerStatusAccepted, Authority: "inferred-state", RequiresConfirmation: false,
			},
		},
	}
	ledger.RecomputeSummary()

	unresolved := ledger.UnresolvedEntries()
	if len(unresolved) != 2 {
		t.Errorf("expected 2 unresolved entries, got %d", len(unresolved))
	}

	requiringConfirm := ledger.RequiringConfirmation()
	if len(requiringConfirm) != 1 {
		t.Errorf("expected 1 entry requiring confirmation, got %d", len(requiringConfirm))
	}
}
