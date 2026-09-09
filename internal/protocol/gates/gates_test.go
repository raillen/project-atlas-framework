package gates

import "testing"

func TestValidate(t *testing.T) {
	if err := Validate(map[string]any{"id": "tests", "status": "passed"}); err != nil {
		t.Fatalf("expected valid gate: %v", err)
	}
	if err := Validate(map[string]any{"id": "tests"}); err == nil {
		t.Fatalf("expected missing status error")
	}
}
