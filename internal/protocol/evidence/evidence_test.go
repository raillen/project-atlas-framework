package evidence

import "testing"

func TestValidate(t *testing.T) {
	if err := Validate(map[string]any{"id": "EV-001", "type": "test"}); err != nil {
		t.Fatalf("expected valid evidence: %v", err)
	}
	if err := Validate(map[string]any{"type": "test"}); err == nil {
		t.Fatalf("expected missing id error")
	}
}
