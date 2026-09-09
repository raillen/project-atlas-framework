package adoption

import (
	"encoding/json"
	"testing"
)

func TestClassifyGoCLI(t *testing.T) {
	facts := []ObservedFact{
		{ID: "fact-0001", Kind: FactManifest, Key: "manifest:go.mod", Source: "go.mod", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0002", Kind: FactFile, Key: "source:cmd/atlas/main.go", Source: "cmd/atlas/main.go", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
		{ID: "fact-0003", Kind: FactTest, Key: "test:internal/app/service_test.go", Source: "internal/app/service_test.go", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0004", Kind: FactCI, Key: "ci:ci.yml", Source: ".github/workflows/ci.yml", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0005", Kind: FactDoc, Key: "doc:README.md", Source: "README.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0006", Kind: FactAgentRules, Key: "agent-rules:AGENTS.md", Source: "AGENTS.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0007", Kind: FactConfig, Key: "config:Dockerfile", Source: "Dockerfile", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0008", Kind: FactConfig, Key: "config:Makefile", Source: "Makefile", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0009", Kind: FactSchema, Key: "schema:atlas.schema.json", Source: "schemas/atlas.schema.json", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
	}

	result := Classify(facts)

	if result.Version != 1 {
		t.Fatalf("expected version 1, got %d", result.Version)
	}

	// Verify signals meet validation invariants
	for _, s := range result.Signals {
		if err := s.Validate(); err != nil {
			t.Errorf("invalid signal: %v", err)
		}
	}

	// Verify Languages
	hasGo := false
	for _, lang := range result.Languages {
		if lang == "go" {
			hasGo = true
		}
	}
	if !hasGo {
		t.Errorf("expected 'go' in languages, got %v", result.Languages)
	}

	// Verify AppTypes
	hasCLI := false
	for _, at := range result.AppTypes {
		if at == "cli" {
			hasCLI = true
		}
	}
	if !hasCLI {
		t.Errorf("expected 'cli' in app_types, got %v", result.AppTypes)
	}

	// Verify Capabilities
	capMap := make(map[string]bool)
	for _, cap := range result.Capabilities {
		capMap[cap] = true
	}
	if !capMap["cli"] || !capMap["go"] || !capMap["schema"] {
		t.Errorf("missing expected capabilities: %v", result.Capabilities)
	}

	// Verify Testing
	hasGoTest := false
	for _, tst := range result.Testing {
		if tst == "go-test" {
			hasGoTest = true
		}
	}
	if !hasGoTest {
		t.Errorf("expected 'go-test' in testing, got %v", result.Testing)
	}

	// Verify AIHarness
	hasAgentsMD := false
	for _, ai := range result.AIHarness {
		if ai == "agents-markdown" {
			hasAgentsMD = true
		}
	}
	if !hasAgentsMD {
		t.Errorf("expected 'agents-markdown' in ai_harness, got %v", result.AIHarness)
	}

	// Verify JSON serialization round-trip
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("JSON marshal failed: %v", err)
	}
	var unmarshaled RepositoryClassification
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("JSON unmarshal failed: %v", err)
	}
}

func TestClassifyWebApp(t *testing.T) {
	facts := []ObservedFact{
		{
			ID:         "fact-0001",
			Kind:       FactManifest,
			Key:        "manifest:package.json",
			Source:     "package.json",
			Extraction: ExtractionManifestField,
			Confidence: ConfidenceFactual,
			Metadata: map[string]any{
				"dependencies": []string{"react", "next"},
				"scripts": map[string]string{
					"test": "jest",
				},
			},
		},
		{ID: "fact-0002", Kind: FactConfig, Key: "config:tsconfig.json", Source: "tsconfig.json", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0003", Kind: FactFile, Key: "source:src/App.tsx", Source: "src/App.tsx", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
	}

	result := Classify(facts)

	// Frameworks should include react and next.js
	fwMap := make(map[string]bool)
	for _, fw := range result.Frameworks {
		fwMap[fw] = true
	}
	if !fwMap["react"] || !fwMap["next.js"] {
		t.Errorf("expected react and next.js in frameworks, got %v", result.Frameworks)
	}

	// AppTypes should include web-application
	hasWebApp := false
	for _, at := range result.AppTypes {
		if at == "web-application" {
			hasWebApp = true
		}
	}
	if !hasWebApp {
		t.Errorf("expected web-application in app_types, got %v", result.AppTypes)
	}

	// Capabilities should include web-application and ui
	capMap := make(map[string]bool)
	for _, c := range result.Capabilities {
		capMap[c] = true
	}
	if !capMap["web-application"] || !capMap["ui"] {
		t.Errorf("expected web-application and ui in capabilities, got %v", result.Capabilities)
	}

	// Testing should include jest from scripts
	hasJest := false
	for _, tst := range result.Testing {
		if tst == "jest" {
			hasJest = true
		}
	}
	if !hasJest {
		t.Errorf("expected jest in testing, got %v", result.Testing)
	}
}

func TestClassifyPythonAPIService(t *testing.T) {
	facts := []ObservedFact{
		{
			ID:         "fact-0001",
			Kind:       FactManifest,
			Key:        "manifest:pyproject.toml",
			Source:     "pyproject.toml",
			Extraction: ExtractionManifestField,
			Confidence: ConfidenceFactual,
			Metadata: map[string]any{
				"dependencies": []string{"fastapi", "sqlalchemy", "psycopg2-binary"},
			},
		},
		{ID: "fact-0002", Kind: FactTest, Key: "test:tests/test_api.py", Source: "tests/test_api.py", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0003", Kind: FactDoc, Key: "doc:docs/architecture.md", Source: "docs/architecture.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0004", Kind: FactConfig, Key: "config:security.md", Source: "SECURITY.md", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
		{ID: "fact-0005", Kind: FactConfig, Key: "config:.env.example", Source: ".env.example", Extraction: ExtractionFilenamePattern, Confidence: ConfidenceFactual},
	}

	result := Classify(facts)

	// AppTypes should include api-service
	hasAPI := false
	for _, at := range result.AppTypes {
		if at == "api-service" {
			hasAPI = true
		}
	}
	if !hasAPI {
		t.Errorf("expected api-service in app_types, got %v", result.AppTypes)
	}

	// Persistence should include sqlalchemy and postgresql
	persMap := make(map[string]bool)
	for _, p := range result.Persistence {
		persMap[p] = true
	}
	if !persMap["sqlalchemy"] || !persMap["postgresql"] {
		t.Errorf("expected sqlalchemy and postgresql in persistence, got %v", result.Persistence)
	}

	// Testing should include pytest
	hasPytest := false
	for _, tst := range result.Testing {
		if tst == "pytest" {
			hasPytest = true
		}
	}
	if !hasPytest {
		t.Errorf("expected pytest in testing, got %v", result.Testing)
	}

	// DocRoles should include system-architecture
	hasArch := false
	for _, dr := range result.DocRoles {
		if dr == "system-architecture" {
			hasArch = true
		}
	}
	if !hasArch {
		t.Errorf("expected system-architecture in doc_roles, got %v", result.DocRoles)
	}

	// Security should include security-policy and env-example
	secMap := make(map[string]bool)
	for _, s := range result.Security {
		secMap[s] = true
	}
	if !secMap["security-policy"] || !secMap["env-example"] {
		t.Errorf("expected security-policy and env-example in security, got %v", result.Security)
	}
}

func TestClassifyEmptyFacts(t *testing.T) {
	result := Classify([]ObservedFact{})
	if result.Version != 1 {
		t.Fatalf("expected version 1, got %d", result.Version)
	}
	if len(result.Signals) != 0 {
		t.Errorf("expected 0 signals for empty facts, got %d", len(result.Signals))
	}
	if len(result.Languages) != 0 || len(result.AppTypes) != 0 {
		t.Errorf("expected empty slices for empty facts")
	}
}

func TestClassifyUnknownAppTypeFallback(t *testing.T) {
	facts := []ObservedFact{
		{ID: "fact-0001", Kind: FactFile, Key: "source:data.txt", Source: "data.txt", Extraction: ExtractionStaticPresence, Confidence: ConfidenceFactual},
	}
	result := Classify(facts)
	hasUnknown := false
	for _, at := range result.AppTypes {
		if at == "unknown" {
			hasUnknown = true
		}
	}
	if !hasUnknown {
		t.Errorf("expected unknown in app_types when no archetypes match, got %v", result.AppTypes)
	}
}
