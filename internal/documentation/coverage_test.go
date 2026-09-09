package docengine

import (
	"path/filepath"
	"testing"
)

func TestAtlasAuditAndReadiness(t *testing.T) {
	root, _ := filepath.Abs("../..")
	audit, err := Audit(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(audit.ApplicableContracts) == 0 {
		t.Fatal("expected applicable contracts")
	}
	readiness, err := Readiness(root, "M5")
	if err != nil {
		t.Fatal(err)
	}
	if len(readiness.Coverage) == 0 {
		t.Fatal("expected coverage")
	}
}
