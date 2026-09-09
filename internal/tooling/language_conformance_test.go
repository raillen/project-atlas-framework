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
}
