package evidence

import "fmt"

type Record struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

func Validate(record map[string]any) error {
	if fmt.Sprint(record["id"]) == "" || fmt.Sprint(record["id"]) == "<nil>" {
		return fmt.Errorf("evidence missing id")
	}
	if fmt.Sprint(record["type"]) == "" || fmt.Sprint(record["type"]) == "<nil>" {
		return fmt.Errorf("evidence missing type")
	}
	return nil
}
