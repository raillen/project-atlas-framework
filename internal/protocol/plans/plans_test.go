package plans

import "testing"

func TestCheckDAGCycles(t *testing.T) {
	if errors := CheckDAGCycles([]map[string]any{{"id": "A", "dependencies": []any{}}, {"id": "B", "dependencies": []any{"A"}}}); len(errors) != 0 {
		t.Fatalf("expected acyclic graph, got %v", errors)
	}
	if errors := CheckDAGCycles([]map[string]any{{"id": "A", "dependencies": []any{"B"}}, {"id": "B", "dependencies": []any{"A"}}}); len(errors) == 0 {
		t.Fatalf("expected cycle")
	}
}

func TestCheckDAGCyclesMissing(t *testing.T) {
	errors := CheckDAGCycles([]map[string]any{{"id": "A", "dependencies": []any{"missing"}}})
	if len(errors) != 1 {
		t.Fatalf("expected one missing dependency error, got %v", errors)
	}
}
