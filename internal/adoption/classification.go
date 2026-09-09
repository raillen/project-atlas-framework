package adoption

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// SignalDimension identifies the category of a classification signal.
type SignalDimension string

const (
	DimLanguage    SignalDimension = "language"
	DimToolchain   SignalDimension = "toolchain"
	DimAppType     SignalDimension = "app-type"
	DimFramework   SignalDimension = "framework"
	DimPersistence SignalDimension = "persistence"
	DimCapability  SignalDimension = "capability"
	DimTesting     SignalDimension = "testing"
	DimDeployment  SignalDimension = "deployment"
	DimDocRole     SignalDimension = "doc-role"
	DimAIHarness   SignalDimension = "ai-harness"
	DimSecurity    SignalDimension = "security"
)

// Valid reports whether the dimension is a recognised classification dimension.
func (d SignalDimension) Valid() bool {
	switch d {
	case DimLanguage, DimToolchain, DimAppType, DimFramework, DimPersistence,
		DimCapability, DimTesting, DimDeployment, DimDocRole, DimAIHarness, DimSecurity:
		return true
	}
	return false
}

// ClassificationSignal represents one classified aspect of the repository,
// backed by observed facts as evidence.
type ClassificationSignal struct {
	Dimension  SignalDimension `json:"dimension"`
	Value      string          `json:"value"`
	Confidence FactConfidence  `json:"confidence"`
	Evidence   []string        `json:"evidence"` // ObservedFact IDs
}

// Validate checks that the signal meets invariant requirements.
func (s ClassificationSignal) Validate() error {
	if !s.Dimension.Valid() {
		return fmt.Errorf("invalid signal dimension: %q", s.Dimension)
	}
	if strings.TrimSpace(s.Value) == "" {
		return fmt.Errorf("signal value is required")
	}
	if !s.Confidence.Valid() {
		return fmt.Errorf("invalid signal confidence: %q", s.Confidence)
	}
	if len(s.Evidence) == 0 {
		return fmt.Errorf("signal requires at least one evidence fact ID")
	}
	return nil
}

// RepositoryClassification is the complete evidence-driven classification output.
type RepositoryClassification struct {
	Version      int                    `json:"version"`
	Signals      []ClassificationSignal `json:"signals"`
	Languages    []string               `json:"languages"`
	Toolchains   []string               `json:"toolchains"`
	AppTypes     []string               `json:"app_types"`
	Frameworks   []string               `json:"frameworks"`
	Persistence  []string               `json:"persistence"`
	Capabilities []string               `json:"capabilities"`
	Testing      []string               `json:"testing"`
	Deployment   []string               `json:"deployment"`
	DocRoles     []string               `json:"doc_roles"`
	AIHarness    []string               `json:"ai_harness"`
	Security     []string               `json:"security"`
}

// Classify analyses observed facts and produces an evidence-driven repository classification.
// It is a pure function that does not perform filesystem operations.
func Classify(facts []ObservedFact) RepositoryClassification {
	c := &classifier{
		facts:   facts,
		signals: make([]ClassificationSignal, 0),
	}

	c.classifyLanguagesAndToolchains()
	c.classifyFrameworksAndPersistence()
	c.classifyTesting()
	c.classifyDeployment()
	c.classifyDocRoles()
	c.classifyAIHarness()
	c.classifySecurity()
	c.classifyAppTypesAndCapabilities()

	return c.buildResult()
}

type classifier struct {
	facts   []ObservedFact
	signals []ClassificationSignal
}

func (c *classifier) addSignal(dim SignalDimension, val string, conf FactConfidence, evidenceFactID string) {
	if val == "" || evidenceFactID == "" {
		return
	}
	// Check if already exists; if so, add evidence
	for i := range c.signals {
		if c.signals[i].Dimension == dim && c.signals[i].Value == val {
			for _, e := range c.signals[i].Evidence {
				if e == evidenceFactID {
					return
				}
			}
			c.signals[i].Evidence = append(c.signals[i].Evidence, evidenceFactID)
			return
		}
	}
	c.signals = append(c.signals, ClassificationSignal{
		Dimension:  dim,
		Value:      val,
		Confidence: conf,
		Evidence:   []string{evidenceFactID},
	})
}

func (c *classifier) classifyLanguagesAndToolchains() {
	for _, f := range c.facts {
		lowerSource := strings.ToLower(f.Source)
		base := filepath.Base(lowerSource)
		ext := strings.ToLower(filepath.Ext(lowerSource))

		switch {
		case base == "go.mod":
			c.addSignal(DimLanguage, "go", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "go-modules", ConfidenceFactual, f.ID)
		case base == "package.json":
			c.addSignal(DimToolchain, "npm", ConfidenceFactual, f.ID)
			c.addSignal(DimLanguage, "javascript", ConfidenceHigh, f.ID)
		case base == "cargo.toml":
			c.addSignal(DimLanguage, "rust", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "cargo", ConfidenceFactual, f.ID)
		case base == "pyproject.toml":
			c.addSignal(DimLanguage, "python", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "poetry/pip", ConfidenceFactual, f.ID)
		case base == "requirements.txt" || base == "setup.py" || base == "setup.cfg":
			c.addSignal(DimLanguage, "python", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "pip", ConfidenceFactual, f.ID)
		case base == "pom.xml":
			c.addSignal(DimLanguage, "java", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "maven", ConfidenceFactual, f.ID)
		case base == "build.gradle" || base == "build.gradle.kts":
			c.addSignal(DimLanguage, "java", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "gradle", ConfidenceFactual, f.ID)
		case base == "gemfile":
			c.addSignal(DimLanguage, "ruby", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "bundler", ConfidenceFactual, f.ID)
		case base == "composer.json":
			c.addSignal(DimLanguage, "php", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "composer", ConfidenceFactual, f.ID)
		case base == "mix.exs":
			c.addSignal(DimLanguage, "elixir", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "mix", ConfidenceFactual, f.ID)
		case base == "pubspec.yaml":
			c.addSignal(DimLanguage, "dart", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "flutter/pub", ConfidenceFactual, f.ID)
		}

		// Toolchain configs
		switch {
		case base == "makefile":
			c.addSignal(DimToolchain, "make", ConfidenceFactual, f.ID)
		case base == "dockerfile" || strings.HasPrefix(base, "dockerfile.") || strings.HasPrefix(base, "docker-compose"):
			c.addSignal(DimToolchain, "docker", ConfidenceFactual, f.ID)
		case base == "tsconfig.json":
			c.addSignal(DimLanguage, "typescript", ConfidenceFactual, f.ID)
			c.addSignal(DimToolchain, "typescript", ConfidenceFactual, f.ID)
		case strings.HasPrefix(base, "vite.config"):
			c.addSignal(DimToolchain, "vite", ConfidenceFactual, f.ID)
		case strings.HasPrefix(base, "webpack.config"):
			c.addSignal(DimToolchain, "webpack", ConfidenceFactual, f.ID)
		case base == ".goreleaser.yml" || base == ".goreleaser.yaml":
			c.addSignal(DimToolchain, "goreleaser", ConfidenceFactual, f.ID)
		case base == "buf.yaml" || base == "buf.gen.yaml":
			c.addSignal(DimToolchain, "buf", ConfidenceFactual, f.ID)
		}

		// File extension signals
		switch ext {
		case ".go":
			c.addSignal(DimLanguage, "go", ConfidenceFactual, f.ID)
		case ".ts", ".tsx":
			c.addSignal(DimLanguage, "typescript", ConfidenceFactual, f.ID)
		case ".js", ".jsx", ".mjs", ".cjs":
			c.addSignal(DimLanguage, "javascript", ConfidenceFactual, f.ID)
		case ".py":
			c.addSignal(DimLanguage, "python", ConfidenceFactual, f.ID)
		case ".rs":
			c.addSignal(DimLanguage, "rust", ConfidenceFactual, f.ID)
		case ".java":
			c.addSignal(DimLanguage, "java", ConfidenceFactual, f.ID)
		case ".kt", ".kts":
			c.addSignal(DimLanguage, "kotlin", ConfidenceFactual, f.ID)
		case ".rb":
			c.addSignal(DimLanguage, "ruby", ConfidenceFactual, f.ID)
		case ".php":
			c.addSignal(DimLanguage, "php", ConfidenceFactual, f.ID)
		case ".cs":
			c.addSignal(DimLanguage, "csharp", ConfidenceFactual, f.ID)
		case ".c", ".h":
			c.addSignal(DimLanguage, "c", ConfidenceFactual, f.ID)
		case ".cpp", ".cc", ".hpp":
			c.addSignal(DimLanguage, "cpp", ConfidenceFactual, f.ID)
		case ".swift":
			c.addSignal(DimLanguage, "swift", ConfidenceFactual, f.ID)
		case ".dart":
			c.addSignal(DimLanguage, "dart", ConfidenceFactual, f.ID)
		case ".ex", ".exs":
			c.addSignal(DimLanguage, "elixir", ConfidenceFactual, f.ID)
		case ".sh", ".bash":
			c.addSignal(DimLanguage, "shell", ConfidenceFactual, f.ID)
		}
	}
}

func (c *classifier) classifyFrameworksAndPersistence() {
	for _, f := range c.facts {
		lowerSource := strings.ToLower(f.Source)
		base := filepath.Base(lowerSource)

		// Check parsed dependencies in manifest metadata
		if deps, ok := f.Metadata["dependencies"].([]string); ok {
			for _, dep := range deps {
				lowerDep := strings.ToLower(dep)
				c.classifyDep(lowerDep, f.ID)
			}
		} else if rawDeps, ok := f.Metadata["dependencies"].([]any); ok {
			for _, dep := range rawDeps {
				lowerDep := strings.ToLower(fmt.Sprint(dep))
				c.classifyDep(lowerDep, f.ID)
			}
		}

		// SQL and schema persistence signals
		if strings.HasSuffix(lowerSource, ".sql") || strings.Contains(lowerSource, "migration") {
			c.addSignal(DimPersistence, "sql", ConfidenceFactual, f.ID)
		}
		if base == "prisma" || strings.HasSuffix(lowerSource, "schema.prisma") {
			c.addSignal(DimPersistence, "prisma", ConfidenceFactual, f.ID)
			c.addSignal(DimFramework, "prisma", ConfidenceFactual, f.ID)
		}
	}
}

func (c *classifier) classifyDep(dep string, factID string) {
	// Web/API frameworks
	switch {
	case strings.Contains(dep, "react"):
		c.addSignal(DimFramework, "react", ConfidenceFactual, factID)
	case strings.Contains(dep, "next"):
		c.addSignal(DimFramework, "next.js", ConfidenceFactual, factID)
	case strings.Contains(dep, "vue"):
		c.addSignal(DimFramework, "vue", ConfidenceFactual, factID)
	case strings.Contains(dep, "svelte"):
		c.addSignal(DimFramework, "svelte", ConfidenceFactual, factID)
	case strings.Contains(dep, "angular"):
		c.addSignal(DimFramework, "angular", ConfidenceFactual, factID)
	case strings.Contains(dep, "express"):
		c.addSignal(DimFramework, "express", ConfidenceFactual, factID)
	case strings.Contains(dep, "fastify"):
		c.addSignal(DimFramework, "fastify", ConfidenceFactual, factID)
	case strings.Contains(dep, "gin-gonic/gin"):
		c.addSignal(DimFramework, "gin", ConfidenceFactual, factID)
	case strings.Contains(dep, "labstack/echo"):
		c.addSignal(DimFramework, "echo", ConfidenceFactual, factID)
	case strings.Contains(dep, "spf13/cobra"):
		c.addSignal(DimFramework, "cobra", ConfidenceFactual, factID)
	case strings.Contains(dep, "django"):
		c.addSignal(DimFramework, "django", ConfidenceFactual, factID)
	case strings.Contains(dep, "flask"):
		c.addSignal(DimFramework, "flask", ConfidenceFactual, factID)
	case strings.Contains(dep, "fastapi"):
		c.addSignal(DimFramework, "fastapi", ConfidenceFactual, factID)
	case strings.Contains(dep, "actix"):
		c.addSignal(DimFramework, "actix-web", ConfidenceFactual, factID)
	case strings.Contains(dep, "axum"):
		c.addSignal(DimFramework, "axum", ConfidenceFactual, factID)
	case strings.Contains(dep, "tokio"):
		c.addSignal(DimFramework, "tokio", ConfidenceFactual, factID)
	case strings.Contains(dep, "clap"):
		c.addSignal(DimFramework, "clap", ConfidenceFactual, factID)
	}

	// Persistence / ORM
	switch {
	case strings.Contains(dep, "gorm.io/gorm"):
		c.addSignal(DimPersistence, "gorm", ConfidenceFactual, factID)
	case strings.Contains(dep, "sqlalchemy"):
		c.addSignal(DimPersistence, "sqlalchemy", ConfidenceFactual, factID)
	case dep == "pg" || strings.Contains(dep, "lib/pq") || strings.Contains(dep, "pgx") || strings.Contains(dep, "psycopg") || strings.Contains(dep, "postgres"):
		c.addSignal(DimPersistence, "postgresql", ConfidenceFactual, factID)
	case strings.Contains(dep, "sqlite"):
		c.addSignal(DimPersistence, "sqlite", ConfidenceFactual, factID)
	case strings.Contains(dep, "mysql") || strings.Contains(dep, "go-sql-driver/mysql"):
		c.addSignal(DimPersistence, "mysql", ConfidenceFactual, factID)
	case strings.Contains(dep, "mongoose") || strings.Contains(dep, "mongodb"):
		c.addSignal(DimPersistence, "mongodb", ConfidenceFactual, factID)
	case strings.Contains(dep, "redis") || strings.Contains(dep, "go-redis"):
		c.addSignal(DimPersistence, "redis", ConfidenceFactual, factID)
	}
}

func (c *classifier) classifyTesting() {
	for _, f := range c.facts {
		lowerSource := strings.ToLower(f.Source)
		base := filepath.Base(lowerSource)

		switch {
		case strings.HasSuffix(lowerSource, "_test.go"):
			c.addSignal(DimTesting, "go-test", ConfidenceFactual, f.ID)
		case strings.HasPrefix(base, "test_") && strings.HasSuffix(base, ".py"):
			c.addSignal(DimTesting, "pytest", ConfidenceFactual, f.ID)
		case strings.HasSuffix(lowerSource, ".test.ts") || strings.HasSuffix(lowerSource, ".spec.ts") ||
			strings.HasSuffix(lowerSource, ".test.js") || strings.HasSuffix(lowerSource, ".spec.js"):
			c.addSignal(DimTesting, "javascript-test", ConfidenceFactual, f.ID)
		case strings.HasPrefix(lowerSource, "playwright") || strings.Contains(lowerSource, "e2e"):
			c.addSignal(DimTesting, "e2e-test", ConfidenceHigh, f.ID)
		}

		if scripts, ok := f.Metadata["scripts"].(map[string]string); ok {
			for _, val := range scripts {
				lowerVal := strings.ToLower(val)
				if strings.Contains(lowerVal, "jest") {
					c.addSignal(DimTesting, "jest", ConfidenceFactual, f.ID)
				}
				if strings.Contains(lowerVal, "vitest") {
					c.addSignal(DimTesting, "vitest", ConfidenceFactual, f.ID)
				}
			}
		}
	}
}

func (c *classifier) classifyDeployment() {
	for _, f := range c.facts {
		lowerSource := strings.ToLower(f.Source)
		base := filepath.Base(lowerSource)

		switch {
		case strings.HasPrefix(lowerSource, ".github/workflows/"):
			c.addSignal(DimDeployment, "github-actions", ConfidenceFactual, f.ID)
			c.addSignal(DimCapability, "ci-cd", ConfidenceFactual, f.ID)
		case strings.HasPrefix(lowerSource, ".circleci/"):
			c.addSignal(DimDeployment, "circleci", ConfidenceFactual, f.ID)
			c.addSignal(DimCapability, "ci-cd", ConfidenceFactual, f.ID)
		case strings.HasPrefix(lowerSource, ".gitlab-ci"):
			c.addSignal(DimDeployment, "gitlab-ci", ConfidenceFactual, f.ID)
			c.addSignal(DimCapability, "ci-cd", ConfidenceFactual, f.ID)
		case base == "dockerfile" || strings.HasPrefix(base, "dockerfile."):
			c.addSignal(DimDeployment, "docker", ConfidenceFactual, f.ID)
		case strings.HasPrefix(base, "docker-compose"):
			c.addSignal(DimDeployment, "docker-compose", ConfidenceFactual, f.ID)
		case strings.Contains(lowerSource, "k8s") || strings.Contains(lowerSource, "kubernetes") || strings.Contains(lowerSource, "helm"):
			c.addSignal(DimDeployment, "kubernetes", ConfidenceHigh, f.ID)
		case base == ".goreleaser.yml" || base == ".goreleaser.yaml":
			c.addSignal(DimDeployment, "goreleaser", ConfidenceFactual, f.ID)
		}
	}
}

func (c *classifier) classifyDocRoles() {
	for _, f := range c.facts {
		lowerSource := strings.ToLower(f.Source)
		base := filepath.Base(lowerSource)

		switch {
		case base == "readme.md":
			c.addSignal(DimDocRole, "product-vision", ConfidenceHigh, f.ID)
		case strings.Contains(lowerSource, "architecture") || strings.Contains(lowerSource, "system") || strings.Contains(lowerSource, "design"):
			c.addSignal(DimDocRole, "system-architecture", ConfidenceHigh, f.ID)
		case strings.Contains(lowerSource, "testing") || strings.Contains(lowerSource, "test-strategy"):
			c.addSignal(DimDocRole, "testing-strategy", ConfidenceHigh, f.ID)
		case base == "security.md" || strings.Contains(lowerSource, "trust-model"):
			c.addSignal(DimDocRole, "security-trust", ConfidenceHigh, f.ID)
		case strings.Contains(lowerSource, "cli") && f.Kind == FactDoc:
			c.addSignal(DimDocRole, "cli-reference", ConfidenceHigh, f.ID)
		case strings.Contains(lowerSource, "install") || strings.Contains(lowerSource, "getting-started"):
			c.addSignal(DimDocRole, "installation", ConfidenceHigh, f.ID)
		case base == "changelog.md" || base == "history.md":
			c.addSignal(DimDocRole, "changelog", ConfidenceFactual, f.ID)
		case f.Kind == FactADR:
			c.addSignal(DimDocRole, "adr", ConfidenceFactual, f.ID)
		}
	}
}

func (c *classifier) classifyAIHarness() {
	for _, f := range c.facts {
		base := strings.ToLower(filepath.Base(f.Source))
		switch base {
		case "agents.md":
			c.addSignal(DimAIHarness, "agents-markdown", ConfidenceFactual, f.ID)
		case ".cursorrules":
			c.addSignal(DimAIHarness, "cursor-rules", ConfidenceFactual, f.ID)
		case ".clinerules":
			c.addSignal(DimAIHarness, "cline-rules", ConfidenceFactual, f.ID)
		case "copilot-instructions.md":
			c.addSignal(DimAIHarness, "copilot-instructions", ConfidenceFactual, f.ID)
		}
	}
}

func (c *classifier) classifySecurity() {
	for _, f := range c.facts {
		lowerSource := strings.ToLower(f.Source)
		base := filepath.Base(lowerSource)

		switch {
		case base == "security.md":
			c.addSignal(DimSecurity, "security-policy", ConfidenceFactual, f.ID)
		case base == ".env.example" || base == ".env.sample":
			c.addSignal(DimSecurity, "env-example", ConfidenceFactual, f.ID)
		case strings.HasPrefix(f.Key, "env-presence:") || base == ".env" || strings.HasPrefix(base, ".env."):
			c.addSignal(DimSecurity, "env-file-present", ConfidenceFactual, f.ID)
		}
	}
}

func (c *classifier) classifyAppTypesAndCapabilities() {
	hasCmd := false
	hasWebFramework := false
	hasAPIFramework := false
	hasSchemas := false
	var cmdFactID string
	var webFactID string
	var apiFactID string
	var schemaFactID string

	for _, s := range c.signals {
		if s.Dimension == DimFramework {
			switch s.Value {
			case "react", "next.js", "vue", "svelte", "angular":
				hasWebFramework = true
				if len(s.Evidence) > 0 {
					webFactID = s.Evidence[0]
				}
			case "gin", "echo", "fastapi", "express", "fastify", "actix-web", "axum":
				hasAPIFramework = true
				if len(s.Evidence) > 0 {
					apiFactID = s.Evidence[0]
				}
			case "cobra", "clap":
				hasCmd = true
				if len(s.Evidence) > 0 {
					cmdFactID = s.Evidence[0]
				}
			}
		}
	}

	for _, f := range c.facts {
		lowerSource := strings.ToLower(f.Source)
		if strings.HasPrefix(lowerSource, "cmd/") || strings.HasPrefix(lowerSource, "cmd\\") {
			hasCmd = true
			if cmdFactID == "" {
				cmdFactID = f.ID
			}
		}
		if f.Kind == FactSchema {
			hasSchemas = true
			if schemaFactID == "" {
				schemaFactID = f.ID
			}
		}
	}

	// Emit AppTypes and Capabilities
	if hasCmd {
		c.addSignal(DimAppType, "cli", ConfidenceHigh, cmdFactID)
		c.addSignal(DimCapability, "cli", ConfidenceHigh, cmdFactID)
	}
	if hasWebFramework {
		c.addSignal(DimAppType, "web-application", ConfidenceHigh, webFactID)
		c.addSignal(DimCapability, "web-application", ConfidenceHigh, webFactID)
		c.addSignal(DimCapability, "ui", ConfidenceHigh, webFactID)
	}
	if hasAPIFramework {
		c.addSignal(DimAppType, "api-service", ConfidenceHigh, apiFactID)
		c.addSignal(DimCapability, "api-service", ConfidenceHigh, apiFactID)
	}
	if hasSchemas {
		c.addSignal(DimCapability, "schema", ConfidenceFactual, schemaFactID)
	}

	// Mirror languages as capabilities (as done in coverage.go)
	for _, s := range c.signals {
		if s.Dimension == DimLanguage {
			c.addSignal(DimCapability, s.Value, s.Confidence, s.Evidence[0])
		}
	}

	// If no app type identified and we have facts, emit "unknown" app-type
	hasAppType := false
	for _, s := range c.signals {
		if s.Dimension == DimAppType {
			hasAppType = true
			break
		}
	}
	if !hasAppType && len(c.facts) > 0 {
		c.addSignal(DimAppType, "unknown", ConfidenceLow, c.facts[0].ID)
	}
}

func (c *classifier) buildResult() RepositoryClassification {
	result := RepositoryClassification{
		Version:      1,
		Signals:      c.signals,
		Languages:    c.extractSorted(DimLanguage),
		Toolchains:   c.extractSorted(DimToolchain),
		AppTypes:     c.extractSorted(DimAppType),
		Frameworks:   c.extractSorted(DimFramework),
		Persistence:  c.extractSorted(DimPersistence),
		Capabilities: c.extractSorted(DimCapability),
		Testing:      c.extractSorted(DimTesting),
		Deployment:   c.extractSorted(DimDeployment),
		DocRoles:     c.extractSorted(DimDocRole),
		AIHarness:    c.extractSorted(DimAIHarness),
		Security:     c.extractSorted(DimSecurity),
	}
	return result
}

func (c *classifier) extractSorted(dim SignalDimension) []string {
	seen := make(map[string]bool)
	var list []string
	for _, s := range c.signals {
		if s.Dimension == dim && !seen[s.Value] {
			seen[s.Value] = true
			list = append(list, s.Value)
		}
	}
	if list == nil {
		return []string{}
	}
	sort.Strings(list)
	return list
}
