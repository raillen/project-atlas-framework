package tooling

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// EscapeHatchRegistry represents the registered escape hatches in a project.
type EscapeHatchRegistry struct {
	SchemaVersion int                `json:"schema_version"`
	Project       string             `json:"project,omitempty"`
	DefaultPolicy string             `json:"default_policy,omitempty"`
	EscapeHatches []EscapeHatchEntry `json:"escape_hatches"`
}

// EscapeHatchEntry is a single formally registered and reviewed escape hatch.
type EscapeHatchEntry struct {
	ID                 string `json:"id"`
	Language           string `json:"language"`
	File               string `json:"file"`
	LineRange          [2]int `json:"line_range,omitempty"`
	HatchType          string `json:"hatch_type"`
	Justification      string `json:"justification"`
	SafetyInvariants   string `json:"safety_invariants"`
	QuarantineBoundary string `json:"quarantine_boundary,omitempty"`
	Reviewer           string `json:"reviewer,omitempty"`
	ApprovedAt         string `json:"approved_at,omitempty"`
	ExpiresAt          string `json:"expires_at,omitempty"`
	Status             string `json:"status,omitempty"`
}

// EscapeHatchFinding represents an escape hatch detected in the codebase.
type EscapeHatchFinding struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	Language   string `json:"language"`
	HatchType  string `json:"hatch_type"`
	Construct  string `json:"construct"`
	Snippet    string `json:"snippet"`
	Registered bool   `json:"registered"`
	RegistryID string `json:"registry_id,omitempty"`
	Severity   string `json:"severity"` // "error" or "warning"
	Message    string `json:"message"`
}

// EscapeHatchReport is the aggregated output of the escape hatch scanner.
type EscapeHatchReport struct {
	Root                 string               `json:"root"`
	TotalScannedFiles    int                  `json:"total_scanned_files"`
	TotalFindings        int                  `json:"total_findings"`
	UnregisteredFindings int                  `json:"unregistered_findings"`
	RegisteredFindings   int                  `json:"registered_findings"`
	Clean                bool                 `json:"clean"`
	Findings             []EscapeHatchFinding `json:"findings"`
}

type patternRule struct {
	language  string
	hatchType string
	name      string
	re        *regexp.Regexp
	message   string
}

var languagePatterns = []patternRule{
	// C++ Modern rules
	{"cpp", "raw_cast", "reinterpret_cast", regexp.MustCompile(`\breinterpret_cast\s*<`), "Forbidden raw pointer cast: reinterpret_cast violates type safety."},
	{"cpp", "raw_cast", "c_style_void_cast", regexp.MustCompile(`\(\s*void\s*\*\s*\)`), "Forbidden C-style void* cast: use std::byte, std::span, or type-safe variants."},
	{"cpp", "naked_allocation", "raw_allocation", regexp.MustCompile(`\b(malloc|free|calloc|realloc)\s*\(`), "Forbidden naked C heap allocation: use RAII and standard containers/smart pointers."},
	{"cpp", "naked_allocation", "naked_delete", regexp.MustCompile(`\bdelete(\s*\[\])?\s+[a-zA-Z0-9_]+`), "Forbidden naked delete: use std::unique_ptr or std::shared_ptr."},
	{"cpp", "goto_statement", "goto", regexp.MustCompile(`\bgoto\s+[a-zA-Z_]`), "Forbidden goto statement: use structured control flow."},
	{"cpp", "unsafe_block", "inline_asm", regexp.MustCompile(`\b(__asm__|asm)\s*(\{|volatile)`), "Inline assembly detected: must be quarantined in verified translation units."},
	{"cpp", "suppression_comment", "nolint", regexp.MustCompile(`(?i)//\s*nolint\b|/\*\s*nolint\s*\*/`), "Unregistered linter suppression comment."},

	// C Rules
	{"c", "banned_api", "banned_string_api", regexp.MustCompile(`\b(gets|strcpy|strcat|sprintf)\s*\(`), "Banned unsafe unbounded string API: use bounded or safe alternatives."},
	{"c", "goto_statement", "goto", regexp.MustCompile(`\bgoto\s+[a-zA-Z_]`), "Goto statement must be reviewed and quarantined."},
	{"c", "suppression_comment", "nolint", regexp.MustCompile(`(?i)//\s*nolint\b|/\*\s*nolint\s*\*/`), "Unregistered linter suppression comment."},

	// Zig Rules
	{"zig", "raw_cast", "ptrCast", regexp.MustCompile(`@ptrCast\b`), "Explicit @ptrCast requires safety invariant justification."},
	{"zig", "raw_cast", "intFromPtr", regexp.MustCompile(`@intFromPtr\b`), "Explicit pointer-to-integer conversion requires registration."},
	{"zig", "raw_cast", "ptrFromInt", regexp.MustCompile(`@ptrFromInt\b`), "Explicit integer-to-pointer conversion requires registration."},

	// Dlang Rules
	{"d", "unsafe_block", "trusted", regexp.MustCompile(`@trusted\b`), "Unverified @trusted block: must register safety boundary."},
	{"d", "concurrency_override", "gshared", regexp.MustCompile(`__gshared\b`), "__gshared bypasses thread-local storage: must register safety proof."},

	// C3 Rules
	{"c3", "raw_cast", "void_ptr_cast", regexp.MustCompile(`\(\s*void\s*\*\s*\)`), "Unsafe void* conversion requires registration."},

	// Odin Rules
	{"odin", "raw_cast", "rawptr_cast", regexp.MustCompile(`\bcast\s*\(\s*rawptr\s*\)`), "Raw pointer cast requires registration."},

	// Rust Rules
	{"rust", "unsafe_block", "unsafe_code", regexp.MustCompile(`\bunsafe\s*(\{|fn\b)`), "Unsafe Rust requires formal registration and invariant proof."},
	{"rust", "suppression_comment", "allow_attr", regexp.MustCompile(`#!?\[allow\(`), "Compiler/Clippy lint suppression requires registration."},

	// Go Rules
	{"go", "unsafe_block", "unsafe_pointer", regexp.MustCompile(`\bunsafe` + `\.Pointer\b`), "unsafe pointer bypasses Go type system: requires registration."},
	{"go", "suppression_comment", "nolint", regexp.MustCompile(`//\s*nolint\b`), "Linter suppression comment requires registration."},

	// TypeScript / JavaScript
	{"typescript", "suppression_comment", "ts_ignore", regexp.MustCompile(`@ts-(ignore|nocheck)`), "TypeScript type checking suppression requires registration."},
	{"typescript", "raw_cast", "as_any", regexp.MustCompile(`\bas\s+any\b`), "'as any' escapes static type check: use unknown or schema validation."},
	{"javascript", "banned_api", "eval", regexp.MustCompile(`\beval\s*\(`), "Dynamic eval execution is prohibited."},

	// Python
	{"python", "suppression_comment", "type_ignore", regexp.MustCompile(`#\s*type:\s*ignore`), "Type check suppression requires registration."},
	{"python", "suppression_comment", "noqa", regexp.MustCompile(`#\s*noqa\b`), "Linter suppression comment requires registration."},
	{"python", "banned_api", "eval_exec", regexp.MustCompile(`\b(eval|exec)\s*\(`), "Dynamic code execution is prohibited."},

	// C#
	{"csharp", "unsafe_block", "unsafe_block", regexp.MustCompile(`\bunsafe\s*\{`), "Unsafe C# block requires registration."},
	{"csharp", "suppression_comment", "pragma_warning", regexp.MustCompile(`#pragma\s+warning\s+disable`), "Compiler warning suppression requires registration."},

	// Java
	{"java", "suppression_comment", "suppress_warnings", regexp.MustCompile(`@SuppressWarnings\(`), "Java warning suppression requires registration."},
	{"java", "unsafe_block", "sun_misc_unsafe", regexp.MustCompile(`sun\.misc\.Unsafe`), "Direct sun.misc.Unsafe usage requires registration."},

	// Kotlin
	{"kotlin", "raw_cast", "not_null_assertion", regexp.MustCompile(`[a-zA-Z0-9_]\s*!!`), "Unsafe not-null assertion (!!) bypasses type safety."},
	{"kotlin", "suppression_comment", "suppress_annotation", regexp.MustCompile(`@Suppress\(`), "Kotlin warning suppression requires registration."},

	// Swift
	{"swift", "unsafe_block", "unsafe_pointer", regexp.MustCompile(`\bUnsafe(Mutable)?(Raw)?Pointer\b`), "Swift UnsafePointer usage requires registration."},

	// Dart
	{"dart", "raw_cast", "as_dynamic", regexp.MustCompile(`\bas\s+dynamic\b`), "Unsound dynamic cast in Dart requires registration."},
	{"dart", "suppression_comment", "ignore_comment", regexp.MustCompile(`//\s*ignore(_for_file)?:\s*`), "Dart analyzer suppression requires registration."},

	// Elixir
	{"elixir", "banned_api", "os_cmd", regexp.MustCompile(`:os\.cmd\s*\(`), "Unsafe shell execution via :os.cmd is prohibited."},

	// Ruby
	{"ruby", "banned_api", "eval", regexp.MustCompile(`\beval\s*\(`), "Dynamic eval in Ruby is prohibited."},
	{"ruby", "suppression_comment", "rubocop_disable", regexp.MustCompile(`#\s*rubocop:disable`), "RuboCop suppression requires registration."},

	// PHP
	{"php", "suppression_comment", "error_suppression", regexp.MustCompile(`@[a-zA-Z_\\$]`), "Error suppression operator (@) is prohibited."},
	{"php", "banned_api", "eval", regexp.MustCompile(`\beval\s*\(`), "Dynamic eval execution in PHP is prohibited."},
	{"php", "suppression_comment", "phpstan_ignore", regexp.MustCompile(`//\s*@phpstan-ignore`), "PHPStan suppression requires registration."},

	// Lua
	{"lua", "banned_api", "loadstring", regexp.MustCompile(`\b(loadstring|loadfile)\s*\(`), "Dynamic code loading in Lua requires registration and sandboxing."},

	// Bash
	{"bash", "banned_api", "eval", regexp.MustCompile(`\beval\s+`), "Dynamic eval execution in shell script is prohibited."},
	{"bash", "suppression_comment", "shellcheck_disable", regexp.MustCompile(`#\s*shellcheck\s+disable`), "ShellCheck suppression requires registration."},
}

var inlineAnnotationRegex = regexp.MustCompile(`ATLAS:ESCAPE_HATCH\[([A-Za-z0-9._-]+)\]`)

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".cpp", ".cxx", ".cc", ".hpp", ".hxx":
		return "cpp"
	case ".h":
		// Ambiguous C or C++: check directory or treat as cpp/c
		return "cpp"
	case ".c":
		return "c"
	case ".s", ".asm":
		return "asm"
	case ".zig":
		return "zig"
	case ".d":
		return "d"
	case ".c3":
		return "c3"
	case ".odin":
		return "odin"
	case ".rs":
		return "rust"
	case ".go":
		return "go"
	case ".ts", ".tsx":
		return "typescript"
	case ".js", ".jsx", ".mjs":
		return "javascript"
	case ".py":
		return "python"
	case ".cs":
		return "csharp"
	case ".java":
		return "java"
	case ".kt", ".kts":
		return "kotlin"
	case ".swift":
		return "swift"
	case ".dart":
		return "dart"
	case ".ex", ".exs":
		return "elixir"
	case ".rb":
		return "ruby"
	case ".php":
		return "php"
	case ".lua":
		return "lua"
	case ".sh", ".bash":
		return "bash"
	default:
		return ""
	}
}

// LoadEscapeHatchRegistry loads the registry file from .atlas/escape-hatches.json.
func LoadEscapeHatchRegistry(root string) (*EscapeHatchRegistry, error) {
	candidates := []string{
		filepath.Join(root, ".atlas", "escape-hatches.json"),
		filepath.Join(root, "escape-hatches.json"),
	}
	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err == nil {
			var reg EscapeHatchRegistry
			if err := json.Unmarshal(data, &reg); err != nil {
				return nil, fmt.Errorf("malformed escape hatch registry at %s: %w", p, err)
			}
			return &reg, nil
		}
	}
	return &EscapeHatchRegistry{SchemaVersion: 1, EscapeHatches: []EscapeHatchEntry{}}, nil
}

// ScanEscapeHatches scans a repository or directory for unregistered escape hatches.
func ScanEscapeHatches(root string) (*EscapeHatchReport, error) {
	registry, err := LoadEscapeHatchRegistry(root)
	if err != nil {
		return nil, err
	}

	report := &EscapeHatchReport{
		Root:     root,
		Findings: []EscapeHatchFinding{},
		Clean:    true,
	}

	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		".atlas/cache": true,
		"vendor":       true,
		"target":       true,
		"build":        true,
		"dist":         true,
		"fixtures":     true,
	}

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if path != root && skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > 2*1024*1024 {
			return nil
		}

		lang := detectLanguage(path)
		if lang == "" {
			return nil
		}

		report.TotalScannedFiles++

		relPath, _ := filepath.Rel(root, path)
		relPath = filepath.ToSlash(relPath)

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineNum := 0
		var prevLine string

		for scanner.Scan() {
			lineNum++
			lineText := scanner.Text()

			// Check inline annotation on current or previous line
			var inlineID string
			if match := inlineAnnotationRegex.FindStringSubmatch(lineText); len(match) > 1 {
				inlineID = match[1]
			} else if match := inlineAnnotationRegex.FindStringSubmatch(prevLine); len(match) > 1 {
				inlineID = match[1]
			}

			// Scan against patterns matching this language (or C rules if header)
			for _, rule := range languagePatterns {
				if rule.language != lang {
					// Also check C rules on cpp headers if applicable
					if !(lang == "cpp" && rule.language == "c") {
						continue
					}
				}

				if rule.re.MatchString(lineText) {
					// Check if registered
					isRegistered := false
					regID := inlineID

					if inlineID != "" {
						isRegistered = true
					} else {
						// Look in registry
						for _, entry := range registry.EscapeHatches {
							entryFile := filepath.ToSlash(entry.File)
							if entryFile == relPath || strings.HasSuffix(relPath, entryFile) {
								if entry.HatchType == rule.hatchType || entry.HatchType == "all" {
									if entry.LineRange[0] == 0 || (lineNum >= entry.LineRange[0] && lineNum <= entry.LineRange[1]) {
										if entry.Status != "revoked" {
											isRegistered = true
											regID = entry.ID
											break
										}
									}
								}
							}
						}
					}

					severity := "error"
					if isRegistered {
						severity = "warning"
						report.RegisteredFindings++
					} else {
						report.UnregisteredFindings++
						report.Clean = false
					}
					report.TotalFindings++

					trimmed := strings.TrimSpace(lineText)
					if len(trimmed) > 120 {
						trimmed = trimmed[:120] + "..."
					}

					report.Findings = append(report.Findings, EscapeHatchFinding{
						File:       relPath,
						Line:       lineNum,
						Language:   lang,
						HatchType:  rule.hatchType,
						Construct:  rule.name,
						Snippet:    trimmed,
						Registered: isRegistered,
						RegistryID: regID,
						Severity:   severity,
						Message:    rule.message,
					})
				}
			}

			prevLine = lineText
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return report, nil
}
