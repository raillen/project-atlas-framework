package automation

import "testing"

func TestIdempotency(t *testing.T) {
	if ShouldRun([]Execution{{IdempotencyKey: "e1", Status: "completed"}}, "e1") {
		t.Fatal("completed execution replayed")
	}
	if !ShouldRun(nil, "e2") {
		t.Fatal("new execution blocked")
	}
}
