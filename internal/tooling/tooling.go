package tooling

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// SecretFinding represents a detected secret or credential.
type SecretFinding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Rule    string `json:"rule"`
	Preview string `json:"preview"`
}

var secretPatterns = []struct {
	rule    string
	pattern *regexp.Regexp
}{
	{"AWS Access Key", regexp.MustCompile(`(?i)\bAKIA[0-9A-Z]{16}\b`)},
	{"GitHub Token", regexp.MustCompile(`(?i)\b(ghp_[a-zA-Z0-9]{36}|github_pat_[a-zA-Z0-9_]{82})\b`)},
	{"Private Key Header", regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----`)},
	{"Generic Password/Secret", regexp.MustCompile(`(?i)\b(password|secret|api_key|access_token|bearer)\s*[:=]\s*["'][a-zA-Z0-9_\-.~!@#$%^&*()+=]{8,}["']`)},
}

// ScanSecrets scans files under root for sensitive credentials.
func ScanSecrets(root string) ([]SecretFinding, error) {
	findings := []SecretFinding{}

	skipDirs := map[string]bool{
		".git":         true,
		"node_modules": true,
		".prumo/cache": true,
		"vendor":       true,
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > 2*1024*1024 {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			for _, sp := range secretPatterns {
				if sp.pattern.MatchString(line) {
					preview := strings.TrimSpace(line)
					if len(preview) > 80 {
						preview = preview[:77] + "..."
					}
					findings = append(findings, SecretFinding{
						File:    rel,
						Line:    lineNum,
						Rule:    sp.rule,
						Preview: preview,
					})
				}
			}
		}
		return nil
	})

	return findings, err
}

// HeaderFinding represents the analysis of a specific security header.
type HeaderFinding struct {
	Header string `json:"header"`
	Status string `json:"status"` // PASS, MISSING, WARNING
	Detail string `json:"detail"`
}

// ScanHeaders checks HTTP response headers against OWASP recommendations.
func ScanHeaders(targetURL string) ([]HeaderFinding, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("HEAD", targetURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Prumo-Tooling/0.5")

	resp, err := client.Do(req)
	if err != nil {
		req.Method = "GET"
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	checks := []struct {
		header      string
		warningDesc string
	}{
		{"Strict-Transport-Security", "Missing HSTS! Vulnerable to protocol downgrade attacks."},
		{"Content-Security-Policy", "Missing CSP! High risk of XSS and content injection."},
		{"X-Frame-Options", "Missing X-Frame-Options! Vulnerable to Clickjacking."},
		{"X-Content-Type-Options", "Missing X-Content-Type-Options! Vulnerable to MIME sniffing."},
		{"Referrer-Policy", "Missing Referrer-Policy! May leak internal URLs."},
		{"Permissions-Policy", "Missing Permissions-Policy! Browser device capabilities not restricted."},
	}

	findings := make([]HeaderFinding, 0, len(checks))
	for _, c := range checks {
		val := resp.Header.Get(c.header)
		if val == "" {
			findings = append(findings, HeaderFinding{
				Header: c.header,
				Status: "MISSING",
				Detail: c.warningDesc,
			})
		} else {
			findings = append(findings, HeaderFinding{
				Header: c.header,
				Status: "PASS",
				Detail: fmt.Sprintf("Configured: %s", val),
			})
		}
	}
	return findings, nil
}

// GenerateStrideTemplate creates a STRIDE threat modeling data structure.
func GenerateStrideTemplate(component string) map[string]any {
	return map[string]any{
		"component": component,
		"generated": time.Now().UTC().Format(time.RFC3339),
		"stride_assessment": map[string]any{
			"Spoofing": map[string]any{
				"applicable":  true,
				"threats":     []string{"Attacker pretends to be a valid user or service"},
				"mitigations": []string{"Mutual TLS, cryptographic signatures, strong authentication"},
			},
			"Tampering": map[string]any{
				"applicable":  true,
				"threats":     []string{"Data or executable code modified in transit or rest"},
				"mitigations": []string{"HMAC / SHA-256 integrity verification, immutable audit logs"},
			},
			"Repudiation": map[string]any{
				"applicable":  true,
				"threats":     []string{"User denies performing an action without verifiable evidence"},
				"mitigations": []string{"Tamper-evident append-only event trail and evidence records"},
			},
			"Information Disclosure": map[string]any{
				"applicable":  true,
				"threats":     []string{"Sensitive data exposed to unauthorized parties"},
				"mitigations": []string{"Secrets scrubbing, encryption at rest and in transit, strict RBAC"},
			},
			"Denial of Service": map[string]any{
				"applicable":  true,
				"threats":     []string{"Resource exhaustion causing outage or degraded service"},
				"mitigations": []string{"Rate limiting, bounded context expansion, strict timeouts"},
			},
			"Elevation of Privilege": map[string]any{
				"applicable":  true,
				"threats":     []string{"Unprivileged entity gains administrative capabilities"},
				"mitigations": []string{"Principle of least privilege, strict execution sandbox boundaries"},
			},
		},
	}
}

// GenerateSecurityChecklist creates a PR review checklist in Markdown format.
func GenerateSecurityChecklist(prNumber string) string {
	return fmt.Sprintf(`## Security Review Checklist for PR #%s

Please verify all security considerations prior to approval and merge:

### 1. Input & Boundary Validation
- [ ] All inputs are strictly validated and bounded against explicit schemas.
- [ ] Output encoding or contextual escaping is applied to prevent injection attacks (SQLi, XSS, Command Injection).

### 2. Authentication & Authorization
- [ ] Sensitive operations and endpoints require authenticated identities.
- [ ] Proper authorization checks are enforced before reading or altering resources (no IDOR).

### 3. Data Protection & Secrets
- [ ] Zero hardcoded credentials, tokens, or encryption keys in the codebase.
- [ ] Sensitive configuration is supplied via secure environment variables or vault.
- [ ] Sensitive fields are redacted in logs and error messages.

### 4. Concurrency & Resource Limits
- [ ] Operations subject to external input have deterministic timeouts.
- [ ] Concurrent routines avoid race conditions (verified with `+"`-race`"+`).
- [ ] Allocations and buffer sizes are bounded to prevent denial of service.

### 5. Dependencies & Third-Party Code
- [ ] New external dependencies are vetted for active maintenance and CVE advisories.
- [ ] Dependency licenses are compatible with project distribution requirements.
`, prNumber)
}

// PermissionFinding represents an insecure file permission detected.
type PermissionFinding struct {
	Path   string `json:"path"`
	Mode   string `json:"mode"`
	Reason string `json:"reason"`
}

// CheckPermissions scans root for insecure file modes (e.g. world-writable).
func CheckPermissions(root string) ([]PermissionFinding, error) {
	findings := []PermissionFinding{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		mode := info.Mode().Perm()
		if mode&0002 != 0 {
			rel, _ := filepath.Rel(root, path)
			findings = append(findings, PermissionFinding{
				Path:   rel,
				Mode:   fmt.Sprintf("%04o", mode),
				Reason: "File or directory is world-writable (o+w)",
			})
		}
		return nil
	})

	return findings, err
}

// ComplexityFinding represents a file or function violating complexity standards.
type ComplexityFinding struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Metric    string `json:"metric"`
	Value     int    `json:"value"`
	Threshold int    `json:"threshold"`
	Message   string `json:"message"`
}

// AnalyzeComplexity checks line lengths of files and functions.
func AnalyzeComplexity(root string, maxFuncLines, maxFileLines int) ([]ComplexityFinding, error) {
	if maxFuncLines <= 0 {
		maxFuncLines = 60
	}
	if maxFileLines <= 0 {
		maxFileLines = 600
	}

	findings := []ComplexityFinding{}
	sourceExts := map[string]bool{
		".go": true, ".py": true, ".js": true, ".ts": true,
		".rs": true, ".java": true, ".c": true, ".cpp": true,
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(path)
		if !sourceExts[ext] {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		rel, _ := filepath.Rel(root, path)
		scanner := bufio.NewScanner(f)
		lineCount := 0
		funcStartLine := 0
		funcLineCount := 0
		inFunc := false

		funcRegex := regexp.MustCompile(`^\s*(func |def |function |async function |pub fn |fn )`)

		for scanner.Scan() {
			lineCount++
			line := scanner.Text()

			if funcRegex.MatchString(line) {
				if inFunc && funcLineCount > maxFuncLines {
					findings = append(findings, ComplexityFinding{
						File:      rel,
						Line:      funcStartLine,
						Metric:    "function_length",
						Value:     funcLineCount,
						Threshold: maxFuncLines,
						Message:   fmt.Sprintf("Function starting at line %d has %d lines (exceeds limit of %d)", funcStartLine, funcLineCount, maxFuncLines),
					})
				}
				inFunc = true
				funcStartLine = lineCount
				funcLineCount = 0
			} else if inFunc {
				funcLineCount++
			}
		}

		if inFunc && funcLineCount > maxFuncLines {
			findings = append(findings, ComplexityFinding{
				File:      rel,
				Line:      funcStartLine,
				Metric:    "function_length",
				Value:     funcLineCount,
				Threshold: maxFuncLines,
				Message:   fmt.Sprintf("Function starting at line %d has %d lines (exceeds limit of %d)", funcStartLine, funcLineCount, maxFuncLines),
			})
		}

		if lineCount > maxFileLines {
			findings = append(findings, ComplexityFinding{
				File:      rel,
				Line:      1,
				Metric:    "file_length",
				Value:     lineCount,
				Threshold: maxFileLines,
				Message:   fmt.Sprintf("File has %d lines (exceeds limit of %d)", lineCount, maxFileLines),
			})
		}

		return nil
	})

	return findings, err
}

// ErrorPatternFinding represents code suppressing or mishandling errors.
type ErrorPatternFinding struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Pattern string `json:"pattern"`
	Preview string `json:"preview"`
}

var bareErrorPatterns = []struct {
	patternName string
	regex       *regexp.Regexp
}{
	{"Bare Python Except", regexp.MustCompile(`^\s*except\s*:\s*(#.*)?$`)},
	{"Suppressed Catch Block", regexp.MustCompile(`catch\s*\([^)]*\)\s*\{\s*\}`)},
	{"Ignored Error Assignment", regexp.MustCompile(`\b_\s*=\s*err\b`)},
}

// CheckBareErrors checks code files for bare exception swallowing or unhandled errors.
func CheckBareErrors(root string) ([]ErrorPatternFinding, error) {
	findings := []ErrorPatternFinding{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > 2*1024*1024 {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		rel, _ := filepath.Rel(root, path)
		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			for _, ep := range bareErrorPatterns {
				if ep.regex.MatchString(line) {
					findings = append(findings, ErrorPatternFinding{
						File:    rel,
						Line:    lineNum,
						Pattern: ep.patternName,
						Preview: strings.TrimSpace(line),
					})
				}
			}
		}
		return nil
	})

	return findings, err
}

var subprocessRiskPatterns = []struct {
	riskName string
	regex    *regexp.Regexp
}{
	{"Python Shell=True Execution", regexp.MustCompile(`subprocess\.(Popen|call|check_call|check_output|run)\(.*shell\s*=\s*True.*\)`)},
	{"Raw OS System Execution", regexp.MustCompile(`\bos\.system\(`)},
	{"Unescaped Exec Command", regexp.MustCompile(`\bexec\.Command\("sh",\s*"-c"`)},
}

// CheckSubprocesses checks code files for hazardous shell execution patterns.
func CheckSubprocesses(root string) ([]ErrorPatternFinding, error) {
	findings := []ErrorPatternFinding{}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if info.Size() > 2*1024*1024 {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		rel, _ := filepath.Rel(root, path)
		scanner := bufio.NewScanner(f)
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			for _, sr := range subprocessRiskPatterns {
				if sr.regex.MatchString(line) {
					findings = append(findings, ErrorPatternFinding{
						File:    rel,
						Line:    lineNum,
						Pattern: sr.riskName,
						Preview: strings.TrimSpace(line),
					})
				}
			}
		}
		return nil
	})

	return findings, err
}
