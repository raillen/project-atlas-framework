package events

import "testing"

func TestValidateEvent(t *testing.T) {
	if err := ValidateEvent(map[string]any{"type": "task.started"}); err != nil {
		t.Fatalf("expected valid event: %v", err)
	}
	if err := ValidateEvent(map[string]any{"type": "unknown"}); err == nil {
		t.Fatalf("expected unknown type error")
	}
	if err := ValidateEvent(map[string]any{}); err == nil {
		t.Fatalf("expected missing type error")
	}
}
