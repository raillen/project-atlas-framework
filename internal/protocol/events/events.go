package events

import "fmt"

var KnownTypes = []string{
	"goal.created", "goal.transitioned", "goal.amended",
	"plan.created", "task.started", "task.completed", "task.failed",
	"attempt.started", "attempt.failed", "attempt.completed",
	"evidence.recorded", "gate.evaluated", "gate.waived",
	"run.started", "run.completed", "run.failed",
}

func KnownType(eventType string) bool {
	for _, known := range KnownTypes {
		if known == eventType {
			return true
		}
	}
	return false
}

func ValidateEvent(event map[string]any) error {
	eventType, _ := event["type"].(string)
	if eventType == "" {
		return fmt.Errorf("event missing type")
	}
	if !KnownType(eventType) {
		return fmt.Errorf("unknown event type: %s", eventType)
	}
	return nil
}
