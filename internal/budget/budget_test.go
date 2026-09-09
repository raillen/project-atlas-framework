package budget

import "testing"

func TestHardBudget(t *testing.T) {
	e := New("run", map[string]float64{"input_tokens": 100})
	e.Mode = "hard"
	var err error
	e, err = e.Consume(map[string]float64{"input_tokens": 90})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = e.Consume(map[string]float64{"input_tokens": 20}); err == nil {
		t.Fatal("expected exhaustion")
	}
}
func TestReservation(t *testing.T) {
	e := New("run", map[string]float64{"tool_calls": 2})
	var err error
	e, err = e.Reserve("r1", "step", map[string]float64{"tool_calls": 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Reservations) != 1 {
		t.Fatal("missing reservation")
	}
}
