package tooling

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanSecrets(t *testing.T) {
	dir := t.TempDir()
	cleanFile := filepath.Join(dir, "clean.go")
	if err := os.WriteFile(cleanFile, []byte("package main\n\nvar host = \"localhost:8080\"\n"), 0644); err != nil {
		t.Fatalf("write clean: %v", err)
	}

	secretFile := filepath.Join(dir, "leaks.txt")
	leakContent := "AWS_KEY = AKIAIOSFODNN7EXAMPLE\nghp_token = ghp_111122223333444455556666777788889999\n"
	if err := os.WriteFile(secretFile, []byte(leakContent), 0644); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	findings, err := ScanSecrets(dir)
	if err != nil {
		t.Fatalf("scan secrets: %v", err)
	}
	if len(findings) < 2 {
		t.Fatalf("expected at least 2 findings, got %d: %v", len(findings), findings)
	}
}

func TestGenerateStrideTemplate(t *testing.T) {
	tmpl := GenerateStrideTemplate("auth-service")
	if tmpl["component"] != "auth-service" {
		t.Fatalf("expected component auth-service, got %v", tmpl["component"])
	}
	assessment, ok := tmpl["stride_assessment"].(map[string]any)
	if !ok {
		t.Fatalf("expected stride_assessment map")
	}
	for _, category := range []string{"Spoofing", "Tampering", "Repudiation", "Information Disclosure", "Denial of Service", "Elevation of Privilege"} {
		if _, exists := assessment[category]; !exists {
			t.Errorf("missing category %s in assessment", category)
		}
	}
}

func TestGenerateSecurityChecklist(t *testing.T) {
	cl := GenerateSecurityChecklist("42")
	if cl == "" {
		t.Fatalf("expected non-empty checklist")
	}
	if !testing.Short() && len(cl) < 100 {
		t.Fatalf("expected detailed checklist")
	}
}

func TestCheckPermissions(t *testing.T) {
	dir := t.TempDir()
	safeFile := filepath.Join(dir, "safe.txt")
	if err := os.WriteFile(safeFile, []byte("safe"), 0644); err != nil {
		t.Fatalf("write safe: %v", err)
	}
	insecureFile := filepath.Join(dir, "insecure.txt")
	if err := os.WriteFile(insecureFile, []byte("insecure"), 0777); err != nil {
		t.Fatalf("write insecure: %v", err)
	}
	if err := os.Chmod(insecureFile, 0777); err != nil {
		t.Fatalf("chmod insecure: %v", err)
	}

	findings, err := CheckPermissions(dir)
	if err != nil {
		t.Fatalf("check permissions: %v", err)
	}
	if len(findings) == 0 {
		t.Fatalf("expected world-writable finding")
	}
}

func TestAnalyzeComplexity(t *testing.T) {
	dir := t.TempDir()
	longFunc := "package main\n\nfunc veryLong() {\n"
	for i := 0; i < 70; i++ {
		longFunc += "\tprintln(1)\n"
	}
	longFunc += "}\n"

	filePath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(filePath, []byte(longFunc), 0644); err != nil {
		t.Fatalf("write code: %v", err)
	}

	findings, err := AnalyzeComplexity(dir, 50, 500)
	if err != nil {
		t.Fatalf("analyze complexity: %v", err)
	}
	if len(findings) == 0 {
		t.Fatalf("expected complexity finding for long function")
	}
}

func TestCheckBareErrors(t *testing.T) {
	dir := t.TempDir()
	pyCode := "try:\n    do_something()\nexcept:\n    pass\n"
	if err := os.WriteFile(filepath.Join(dir, "err.py"), []byte(pyCode), 0644); err != nil {
		t.Fatalf("write err.py: %v", err)
	}

	findings, err := CheckBareErrors(dir)
	if err != nil {
		t.Fatalf("check bare errors: %v", err)
	}
	if len(findings) == 0 {
		t.Fatalf("expected bare except finding")
	}
}

func TestCheckSubprocesses(t *testing.T) {
	dir := t.TempDir()
	pyCode := "import subprocess\nsubprocess.run('ls -la', shell=True)\n"
	if err := os.WriteFile(filepath.Join(dir, "sub.py"), []byte(pyCode), 0644); err != nil {
		t.Fatalf("write sub.py: %v", err)
	}

	findings, err := CheckSubprocesses(dir)
	if err != nil {
		t.Fatalf("check subprocesses: %v", err)
	}
	if len(findings) == 0 {
		t.Fatalf("expected subprocess finding")
	}
}
