package observability

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Event struct {
	ID    string         `json:"id"`
	Type  string         `json:"type"`
	RunID string         `json:"run_id"`
	At    string         `json:"at"`
	Data  map[string]any `json:"data,omitempty"`
}

func NewEvent(id, kind, run string, data map[string]any) Event {
	return Event{ID: id, Type: kind, RunID: run, At: time.Now().UTC().Format(time.RFC3339Nano), Data: data}
}
func Append(path string, event Event) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}
