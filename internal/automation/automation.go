package automation

type Execution struct {
	ID             string `json:"id"`
	EventID        string `json:"event_id"`
	IdempotencyKey string `json:"idempotency_key"`
	Status         string `json:"status"`
	Attempts       int    `json:"attempts"`
	Failure        string `json:"failure,omitempty"`
}
type DLQEntry struct {
	EventID  string `json:"event_id"`
	Failure  string `json:"failure"`
	Attempts int    `json:"attempts"`
	Recovery string `json:"recovery"`
}

func ShouldRun(existing []Execution, key string) bool {
	for _, e := range existing {
		if e.IdempotencyKey == key && e.Status == "completed" {
			return false
		}
	}
	return true
}
