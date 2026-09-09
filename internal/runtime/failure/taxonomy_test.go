package failure

import "testing"

func TestRetryable(t *testing.T) {
	if !RateLimited.Retryable() {
		t.Fatal("rate limit should retry")
	}
	if BudgetExhausted.Retryable() {
		t.Fatal("budget should stop")
	}
}
