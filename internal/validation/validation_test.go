package validation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckJSON(t *testing.T) {
	if err := CheckJSON([]byte(`{"valid": true}`)); err != nil {
		t.Fatalf("expected valid json, got error: %v", err)
	}
	if err := CheckJSON([]byte(`invalid json`)); err == nil {
		t.Fatalf("expected invalid json to return error")
	}
	if err := CheckJSON([]byte(`null`)); err == nil {
		t.Fatalf("expected null json to return error")
	}
	if err := CheckJSONObject([]byte(`{"valid": true}`)); err != nil {
		t.Fatalf("expected valid object, got error: %v", err)
	}
	if err := CheckJSONObject([]byte(`[1, 2, 3]`)); err == nil {
		t.Fatalf("expected array to fail object check")
	}
}

func schemaDir(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	return filepath.Join(root, "schemas")
}

func TestLoadRegistry(t *testing.T) {
	registry, err := LoadRegistry(schemaDir(t))
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	if _, ok := registry.Schema("goal.schema.json"); !ok {
		t.Fatalf("goal schema missing")
	}
}

func TestValidateGoalFixture(t *testing.T) {
	dir := schemaDir(t)
	registry, err := LoadRegistry(dir)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "..", "conformance", "goals", "valid_locked_goal.json"))
	if err != nil {
		t.Skipf("fixture missing: %v", err)
	}
	var instance any
	if err := json.Unmarshal(data, &instance); err != nil {
		t.Fatalf("fixture parse: %v", err)
	}
	schema, _ := registry.Schema("goal.schema.json")
	if errors := ValidateSchema(instance, schema, registry); len(errors) != 0 {
		t.Fatalf("expected conformance goal valid, got %v", errors)
	}
}

func TestValidateGoalRejectsInvalid(t *testing.T) {
	dir := schemaDir(t)
	registry, err := LoadRegistry(dir)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	schema, _ := registry.Schema("goal.schema.json")
	errors := ValidateSchema(map[string]any{"id": "BAD ID!", "state": "UNKNOWN"}, schema, registry)
	if len(errors) == 0 {
		t.Fatalf("expected invalid goal to fail")
	}
}

func TestValidateWithNestedRef(t *testing.T) {
	dir := schemaDir(t)
	registry, err := LoadRegistry(dir)
	if err != nil {
		t.Fatalf("registry: %v", err)
	}
	schema, _ := registry.Schema("plan.schema.json")
	errors := ValidateSchema(map[string]any{}, schema, registry)
	if len(errors) == 0 {
		t.Fatalf("expected empty plan to fail")
	}
}
