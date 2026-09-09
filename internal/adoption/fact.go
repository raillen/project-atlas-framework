package adoption

import (
	"fmt"
	"strings"
)

// FactKind classifies what category of repository evidence a fact represents.
type FactKind string

const (
	FactFile        FactKind = "file"
	FactManifest    FactKind = "manifest"
	FactConfig      FactKind = "config"
	FactDoc         FactKind = "doc"
	FactCI          FactKind = "ci"
	FactTest        FactKind = "test"
	FactSchema      FactKind = "schema"
	FactADR         FactKind = "adr"
	FactAgentRules  FactKind = "agent-rules"
	FactGitMetadata FactKind = "git-metadata"
	FactOther       FactKind = "other"
)

// Valid reports whether the kind is a recognised fact category.
func (k FactKind) Valid() bool {
	switch k {
	case FactFile, FactManifest, FactConfig, FactDoc, FactCI, FactTest,
		FactSchema, FactADR, FactAgentRules, FactGitMetadata, FactOther:
		return true
	}
	return false
}

// Extraction describes the method used to derive a fact from its source.
type Extraction string

const (
	ExtractionStaticPresence     Extraction = "static-presence"
	ExtractionFilenamePattern    Extraction = "filename-pattern"
	ExtractionContentParse       Extraction = "content-parse"
	ExtractionManifestField      Extraction = "manifest-field"
	ExtractionGitCommand         Extraction = "git-command"
	ExtractionDirectoryStructure Extraction = "directory-structure"
)

// Valid reports whether the extraction method is recognised.
func (e Extraction) Valid() bool {
	switch e {
	case ExtractionStaticPresence, ExtractionFilenamePattern,
		ExtractionContentParse, ExtractionManifestField,
		ExtractionGitCommand, ExtractionDirectoryStructure:
		return true
	}
	return false
}

// FactConfidence indicates how certain the fact extraction is.
// "factual" means deterministically verifiable (file exists, field parsed).
type FactConfidence string

const (
	ConfidenceFactual FactConfidence = "factual"
	ConfidenceHigh    FactConfidence = "high"
	ConfidenceMedium  FactConfidence = "medium"
	ConfidenceLow     FactConfidence = "low"
)

// Valid reports whether the confidence level is recognised.
func (c FactConfidence) Valid() bool {
	switch c {
	case ConfidenceFactual, ConfidenceHigh, ConfidenceMedium, ConfidenceLow:
		return true
	}
	return false
}

// ObservedFact is a single verifiable piece of repository evidence.
// It never contains inferences — only deterministically extractable data.
type ObservedFact struct {
	ID         string         `json:"id"`
	Kind       FactKind       `json:"kind"`
	Key        string         `json:"key"`
	Value      string         `json:"value,omitempty"`
	Source     string         `json:"source"`
	SourceHash string         `json:"source_hash,omitempty"`
	Location   string         `json:"location,omitempty"`
	Extraction Extraction     `json:"extraction"`
	Confidence FactConfidence `json:"confidence"`
	Revision   string         `json:"revision,omitempty"`
	ScannedAt  string         `json:"scanned_at,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// Validate enforces the invariants required by the F-G01 spec.
func (f ObservedFact) Validate() error {
	if strings.TrimSpace(f.ID) == "" {
		return fmt.Errorf("observed fact requires an id")
	}
	if !f.Kind.Valid() {
		return fmt.Errorf("unknown fact kind %q", f.Kind)
	}
	if strings.TrimSpace(f.Key) == "" {
		return fmt.Errorf("observed fact requires a key")
	}
	if strings.TrimSpace(f.Source) == "" {
		return fmt.Errorf("observed fact requires a source")
	}
	if !f.Extraction.Valid() {
		return fmt.Errorf("unknown extraction method %q", f.Extraction)
	}
	if !f.Confidence.Valid() {
		return fmt.Errorf("unknown fact confidence %q", f.Confidence)
	}
	return nil
}
