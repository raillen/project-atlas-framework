package adoption

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/raillen/project-atlas-framework/internal/indexing"
	"github.com/raillen/project-atlas-framework/internal/runtime"
)

// ScanBudget controls how much work the scanner performs in a single pass.
type ScanBudget struct {
	MaxFiles int   `json:"max_files"`
	MaxBytes int64 `json:"max_bytes"`
}

// DefaultBudget provides sensible defaults for a repository scan.
func DefaultBudget() ScanBudget {
	return ScanBudget{MaxFiles: 10000, MaxBytes: 50 * 1024 * 1024} // 10k files, 50 MB
}

// BudgetStatus tracks scan resource consumption.
type BudgetStatus struct {
	MaxFiles     int    `json:"max_files"`
	MaxBytes     int64  `json:"max_bytes"`
	FilesScanned int    `json:"files_scanned"`
	BytesScanned int64  `json:"bytes_scanned"`
	Exhausted    bool   `json:"exhausted"`
	Continuation string `json:"continuation,omitempty"`
}

// ScanResult is the complete output of a discovery pass.
type ScanResult struct {
	Version          int            `json:"version"`
	Root             string         `json:"root"`
	Revision         string         `json:"revision"`
	Branch           string         `json:"branch"`
	Dirty            bool           `json:"dirty"`
	Facts            []ObservedFact `json:"facts"`
	Budget           BudgetStatus   `json:"budget"`
	ScannedAt        string         `json:"scanned_at"`
	PreviousRevision string         `json:"previous_revision,omitempty"`
	Incremental      bool           `json:"incremental"`
	ExcludedPaths    []string       `json:"excluded_paths,omitempty"`
	Warnings         []string       `json:"warnings,omitempty"`
}

// ScanOptions configures a scan pass.
type ScanOptions struct {
	Budget          ScanBudget
	PreviousIndex   *indexing.Index // non-nil enables incremental mode
	ExcludePatterns []string        // additional exclusion globs
}

// DefaultExclusions returns directory names that are always excluded from scanning.
func DefaultExclusions() []string {
	return []string{
		".git", "node_modules", "vendor", "__pycache__", ".venv", "venv",
		".mypy_cache", ".ruff_cache", ".pytest_cache", "dist", "build",
		".next", ".nuxt", "target", ".gradle", ".idea", ".vscode",
	}
}

// Scanner discovers observed facts in a repository root.
type Scanner struct {
	root       string
	options    ScanOptions
	exclusions map[string]bool
	budget     BudgetStatus
	facts      []ObservedFact
	warnings   []string
	repo       runtime.RepositoryState
	nextID     int
}

// NewScanner creates a scanner for the given root with options.
func NewScanner(root string, opts ScanOptions) *Scanner {
	excl := map[string]bool{}
	for _, p := range DefaultExclusions() {
		excl[p] = true
	}
	for _, p := range opts.ExcludePatterns {
		excl[p] = true
	}
	return &Scanner{
		root:       root,
		options:    opts,
		exclusions: excl,
		budget: BudgetStatus{
			MaxFiles: opts.Budget.MaxFiles,
			MaxBytes: opts.Budget.MaxBytes,
		},
	}
}

// Scan executes the discovery pass and returns the result.
func (s *Scanner) Scan() ScanResult {
	s.repo = runtime.InspectRepository(s.root)
	now := time.Now().UTC().Format(time.RFC3339)

	_ = filepath.WalkDir(s.root, s.visit)

	result := ScanResult{
		Version:   1,
		Root:      s.root,
		Revision:  s.repo.Revision,
		Branch:    s.repo.Branch,
		Dirty:     s.repo.Dirty,
		Facts:     s.facts,
		Budget:    s.budget,
		ScannedAt: now,
		Warnings:  s.warnings,
	}

	if s.options.PreviousIndex != nil {
		result.Incremental = true
		result.PreviousRevision = s.options.PreviousIndex.Revision
	}

	excluded := make([]string, 0, len(s.exclusions))
	for p := range s.exclusions {
		excluded = append(excluded, p)
	}
	result.ExcludedPaths = excluded

	if result.Facts == nil {
		result.Facts = []ObservedFact{}
	}
	if result.ExcludedPaths == nil {
		result.ExcludedPaths = []string{}
	}
	if result.Warnings == nil {
		result.Warnings = []string{}
	}

	return result
}

func (s *Scanner) visit(path string, d fs.DirEntry, err error) error {
	if err != nil {
		s.warnings = append(s.warnings, fmt.Sprintf("walk error: %s: %v", path, err))
		return nil // continue walking
	}

	rel, _ := filepath.Rel(s.root, path)
	if rel == "." {
		return nil
	}

	// Check exclusions against each path component.
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if s.exclusions[part] {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
	}

	// Budget check.
	if s.budget.FilesScanned >= s.budget.MaxFiles {
		s.budget.Exhausted = true
		return fs.SkipAll
	}

	if d.IsDir() {
		s.emitDirectoryFact(rel)
		return nil
	}

	info, infoErr := d.Info()
	if infoErr != nil {
		s.warnings = append(s.warnings, fmt.Sprintf("stat error: %s: %v", rel, infoErr))
		return nil
	}

	if s.budget.MaxBytes > 0 && s.budget.BytesScanned+info.Size() > s.budget.MaxBytes {
		s.budget.Exhausted = true
		s.budget.Continuation = rel
		return fs.SkipAll
	}

	s.budget.FilesScanned++
	s.budget.BytesScanned += info.Size()

	s.classifyAndEmit(rel)
	return nil
}

func (s *Scanner) emitDirectoryFact(rel string) {
	name := filepath.Base(rel)
	kind := classifyDirectory(name)
	if kind == "" {
		return // not a structurally significant directory
	}
	s.emit(ObservedFact{
		Kind:       kind,
		Key:        "directory:" + name,
		Source:     rel,
		Extraction: ExtractionDirectoryStructure,
		Confidence: ConfidenceFactual,
	})
}

func (s *Scanner) classifyAndEmit(rel string) {
	name := filepath.Base(rel)
	ext := strings.ToLower(filepath.Ext(name))
	lower := strings.ToLower(name)

	// Compute hash for source_hash.
	hash := fileHash(filepath.Join(s.root, rel))

	// Manifests.
	if isManifest(lower) {
		s.emit(ObservedFact{
			Kind:       FactManifest,
			Key:        "manifest:" + lower,
			Source:     rel,
			SourceHash: hash,
			Extraction: ExtractionFilenamePattern,
			Confidence: ConfidenceFactual,
		})
		return
	}

	// CI / Workflows.
	if isCI(rel) {
		s.emit(ObservedFact{
			Kind:       FactCI,
			Key:        "ci:" + name,
			Source:     rel,
			SourceHash: hash,
			Extraction: ExtractionFilenamePattern,
			Confidence: ConfidenceFactual,
		})
		return
	}

	// Agent rules (checked before docs so AGENTS.md is not misclassified).
	if isAgentRules(lower) {
		s.emit(ObservedFact{
			Kind:       FactAgentRules,
			Key:        "agent-rules:" + name,
			Source:     rel,
			SourceHash: hash,
			Extraction: ExtractionFilenamePattern,
			Confidence: ConfidenceFactual,
		})
		return
	}

	// Documentation.
	if isDoc(rel, lower, ext) {
		s.emit(ObservedFact{
			Kind:       FactDoc,
			Key:        "doc:" + rel,
			Source:     rel,
			SourceHash: hash,
			Extraction: ExtractionFilenamePattern,
			Confidence: ConfidenceFactual,
		})
		return
	}

	// Schemas.
	if isSchema(lower) {
		s.emit(ObservedFact{
			Kind:       FactSchema,
			Key:        "schema:" + name,
			Source:     rel,
			SourceHash: hash,
			Extraction: ExtractionFilenamePattern,
			Confidence: ConfidenceFactual,
		})
		return
	}

	// ADRs.
	if isADR(rel) {
		s.emit(ObservedFact{
			Kind:       FactADR,
			Key:        "adr:" + name,
			Source:     rel,
			SourceHash: hash,
			Extraction: ExtractionFilenamePattern,
			Confidence: ConfidenceFactual,
		})
		return
	}

	// Configs.
	if isConfig(lower, ext) {
		s.emit(ObservedFact{
			Kind:       FactConfig,
			Key:        "config:" + name,
			Source:     rel,
			SourceHash: hash,
			Extraction: ExtractionFilenamePattern,
			Confidence: ConfidenceFactual,
		})
		return
	}

	// Test files.
	if isTest(rel, lower) {
		s.emit(ObservedFact{
			Kind:       FactTest,
			Key:        "test:" + rel,
			Source:     rel,
			SourceHash: hash,
			Extraction: ExtractionFilenamePattern,
			Confidence: ConfidenceFactual,
		})
		return
	}

	// Source files — emitted as generic file facts.
	if isSourceFile(ext) {
		s.emit(ObservedFact{
			Kind:       FactFile,
			Key:        "source:" + rel,
			Source:     rel,
			SourceHash: hash,
			Extraction: ExtractionStaticPresence,
			Confidence: ConfidenceFactual,
		})
	}
}

func (s *Scanner) emit(fact ObservedFact) {
	s.nextID++
	fact.ID = fmt.Sprintf("fact-%04d", s.nextID)
	fact.Revision = s.repo.Revision
	fact.ScannedAt = time.Now().UTC().Format(time.RFC3339)
	s.facts = append(s.facts, fact)
}

func fileHash(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

// --- Classification helpers ---

func isManifest(name string) bool {
	manifests := map[string]bool{
		"package.json": true, "go.mod": true, "cargo.toml": true,
		"pyproject.toml": true, "setup.py": true, "setup.cfg": true,
		"pom.xml": true, "build.gradle": true, "build.gradle.kts": true,
		"gemfile": true, "composer.json": true, "mix.exs": true,
		"pubspec.yaml": true, "deno.json": true, "deno.jsonc": true,
		"atlas.json": true, "requirements.txt": true,
	}
	return manifests[name]
}

func isCI(rel string) bool {
	lower := strings.ToLower(rel)
	return strings.HasPrefix(lower, ".github"+string(filepath.Separator)+"workflows"+string(filepath.Separator)) ||
		strings.HasPrefix(lower, ".circleci"+string(filepath.Separator)) ||
		strings.HasPrefix(lower, ".gitlab-ci") ||
		strings.HasPrefix(lower, "jenkinsfile") ||
		strings.HasPrefix(lower, ".travis")
}

func isDoc(rel, name, ext string) bool {
	if ext != ".md" && ext != ".rst" && ext != ".txt" && ext != ".adoc" {
		return false
	}
	lower := strings.ToLower(rel)
	if strings.Contains(lower, "doc") {
		return true
	}
	knownDocs := map[string]bool{
		"readme.md": true, "changelog.md": true, "contributing.md": true,
		"license.md": true, "security.md": true, "code_of_conduct.md": true,
		"history.md": true,
	}
	if knownDocs[name] {
		return true
	}
	// Top-level markdown files are docs.
	if !strings.Contains(rel, string(filepath.Separator)) {
		return true
	}
	return false
}

func isConfig(name, ext string) bool {
	configs := map[string]bool{
		".editorconfig": true, ".eslintrc": true, ".eslintrc.json": true,
		".prettierrc": true, ".prettierrc.json": true, "tsconfig.json": true,
		"webpack.config.js": true, "vite.config.ts": true, "vite.config.js": true,
		".env.example": true, "docker-compose.yml": true, "docker-compose.yaml": true,
		"dockerfile": true, "makefile": true, ".gitignore": true, ".gitattributes": true,
		"renovate.json": true, ".releaserc": true, ".goreleaser.yml": true,
		".goreleaser.yaml": true, "buf.yaml": true, "buf.gen.yaml": true,
		"rustfmt.toml": true, "clippy.toml": true,
	}
	if configs[name] {
		return true
	}
	if ext == ".ini" || ext == ".cfg" {
		return true
	}
	return false
}

func isSchema(name string) bool {
	return strings.HasSuffix(name, ".schema.json") ||
		strings.HasSuffix(name, ".schema.yaml") ||
		strings.HasSuffix(name, ".graphql") ||
		strings.HasSuffix(name, ".proto")
}

func isADR(rel string) bool {
	lower := strings.ToLower(rel)
	parts := strings.Split(lower, string(filepath.Separator))
	for _, p := range parts {
		if p == "adr" || p == "adrs" || p == "decisions" {
			return true
		}
	}
	return false
}

func isAgentRules(name string) bool {
	rules := map[string]bool{
		"agents.md": true, ".cursorrules": true, ".clinerules": true,
		"copilot-instructions.md": true,
	}
	return rules[name]
}

func isTest(rel, name string) bool {
	if strings.HasPrefix(name, "test_") || strings.HasSuffix(name, "_test.go") ||
		strings.HasSuffix(name, ".test.ts") || strings.HasSuffix(name, ".test.js") ||
		strings.HasSuffix(name, ".spec.ts") || strings.HasSuffix(name, ".spec.js") ||
		strings.HasSuffix(name, ".test.tsx") || strings.HasSuffix(name, ".test.jsx") {
		return true
	}
	// Files inside test directories.
	lower := strings.ToLower(rel)
	parts := strings.Split(lower, string(filepath.Separator))
	for _, p := range parts {
		if p == "test" || p == "tests" || p == "__tests__" || p == "spec" || p == "specs" || p == "e2e" {
			return true
		}
	}
	return false
}

func isSourceFile(ext string) bool {
	exts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true, ".tsx": true,
		".jsx": true, ".rs": true, ".java": true, ".kt": true, ".rb": true,
		".c": true, ".cpp": true, ".h": true, ".hpp": true, ".cs": true,
		".swift": true, ".ex": true, ".exs": true, ".dart": true, ".php": true,
		".lua": true, ".zig": true, ".nim": true, ".v": true, ".sql": true,
		".sh": true, ".bash": true, ".zsh": true, ".fish": true,
	}
	return exts[ext]
}

func classifyDirectory(name string) FactKind {
	dirs := map[string]FactKind{
		"docs": FactDoc, "doc": FactDoc, "documentation": FactDoc,
		"test": FactTest, "tests": FactTest, "spec": FactTest, "specs": FactTest,
		"__tests__": FactTest, "e2e": FactTest,
		"schemas": FactSchema, "schema": FactSchema,
		".github": FactCI, ".circleci": FactCI, ".gitlab": FactCI,
		"adr": FactADR, "adrs": FactADR, "decisions": FactADR,
		"scripts": FactConfig,
	}
	if kind, ok := dirs[name]; ok {
		return kind
	}
	return ""
}
