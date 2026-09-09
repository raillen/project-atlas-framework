package gates

import "fmt"

type Result struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

func Validate(gate map[string]any) error {
	if fmt.Sprint(gate["id"]) == "" || fmt.Sprint(gate["id"]) == "<nil>" {
		return fmt.Errorf("gate missing id")
	}
	if fmt.Sprint(gate["status"]) == "" || fmt.Sprint(gate["status"]) == "<nil>" {
		return fmt.Errorf("gate missing status")
	}
	return nil
}
