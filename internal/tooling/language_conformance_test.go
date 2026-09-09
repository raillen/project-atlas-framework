package tooling

import (
	"path/filepath"
	"testing"
)

func TestLanguageConformanceFixtures(t *testing.T) {
	fixtureBase, err := filepath.Abs("../../tests/fixtures/language_conformance")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("cpp_clean", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "cpp_clean"))
		if err != nil {
			t.Fatal(err)
		}
		if !report.Clean {
			t.Errorf("expected clean=true for cpp_clean, got %d findings", report.TotalFindings)
		}
	})

	t.Run("cpp_violating", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "cpp_violating"))
		if err != nil {
			t.Fatal(err)
		}
		if report.Clean {
			t.Errorf("expected clean=false for cpp_violating")
		}
		if report.UnregisteredFindings < 3 {
			t.Errorf("expected at least 3 unregistered findings, got %d", report.UnregisteredFindings)
		}
	})

	t.Run("cpp_registered", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "cpp_registered"))
		if err != nil {
			t.Fatal(err)
		}
		if !report.Clean {
			t.Errorf("expected clean=true for cpp_registered, got clean=false")
		}
		if report.RegisteredFindings != 1 {
			t.Errorf("expected 1 registered finding, got %d", report.RegisteredFindings)
		}
		if report.UnregisteredFindings != 0 {
			t.Errorf("expected 0 unregistered findings, got %d", report.UnregisteredFindings)
		}
	})

	t.Run("zig_clean", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "zig_clean"))
		if err != nil {
			t.Fatal(err)
		}
		if !report.Clean {
			t.Errorf("expected clean=true for zig_clean, got %d findings", report.TotalFindings)
		}
	})

	t.Run("rust_clean", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "rust_clean"))
		if err != nil {
			t.Fatal(err)
		}
		if !report.Clean {
			t.Errorf("expected clean=true for rust_clean, got %d findings", report.TotalFindings)
		}
	})

	t.Run("go_clean", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "go_clean"))
		if err != nil {
			t.Fatal(err)
		}
		if !report.Clean {
			t.Errorf("expected clean=true for go_clean, got %d findings", report.TotalFindings)
		}
	})

	t.Run("ts_clean", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "ts_clean"))
		if err != nil {
			t.Fatal(err)
		}
		if !report.Clean {
			t.Errorf("expected clean=true for ts_clean, got %d findings", report.TotalFindings)
		}
	})

	t.Run("python_violating", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "python_violating"))
		if err != nil {
			t.Fatal(err)
		}
		if report.Clean {
			t.Errorf("expected clean=false for python_violating")
		}
		if report.UnregisteredFindings < 1 {
			t.Errorf("expected at least 1 unregistered finding, got %d", report.UnregisteredFindings)
		}
	})

	t.Run("kotlin_violating", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "kotlin_violating"))
		if err != nil {
			t.Fatal(err)
		}
		if report.Clean {
			t.Errorf("expected clean=false for kotlin_violating")
		}
		if report.UnregisteredFindings < 1 {
			t.Errorf("expected at least 1 unregistered finding, got %d", report.UnregisteredFindings)
		}
	})

	t.Run("html_clean", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "html_clean"))
		if err != nil {
			t.Fatal(err)
		}
		if !report.Clean {
			t.Errorf("expected clean=true for html_clean, got %d findings", report.TotalFindings)
		}
	})

	t.Run("sql_clean", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "sql_clean"))
		if err != nil {
			t.Fatal(err)
		}
		if !report.Clean {
			t.Errorf("expected clean=true for sql_clean, got %d findings", report.TotalFindings)
		}
	})

	t.Run("dockerfile_clean", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "dockerfile_clean"))
		if err != nil {
			t.Fatal(err)
		}
		if !report.Clean {
			t.Errorf("expected clean=true for dockerfile_clean, got %d findings", report.TotalFindings)
		}
	})

	t.Run("dockerfile_violating", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "dockerfile_violating"))
		if err != nil {
			t.Fatal(err)
		}
		if report.Clean {
			t.Errorf("expected clean=false for dockerfile_violating")
		}
		if report.UnregisteredFindings < 2 {
			t.Errorf("expected at least 2 unregistered findings, got %d", report.UnregisteredFindings)
		}
	})

	t.Run("html_violating", func(t *testing.T) {
		report, err := ScanEscapeHatches(filepath.Join(fixtureBase, "html_violating"))
		if err != nil {
			t.Fatal(err)
		}
		if report.Clean {
			t.Errorf("expected clean=false for html_violating")
		}
		if report.UnregisteredFindings < 2 {
			t.Errorf("expected at least 2 unregistered findings, got %d", report.UnregisteredFindings)
		}
	})
}
