package experience

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ExperienceProvider defines the contract for episodic and experiential state storage.
type ExperienceProvider interface {
	RecordEvent(event SessionEvent) error
	GetEvents(sessionID string) ([]SessionEvent, error)
	SaveSummary(summary Summary) error
	GetSummary(sessionID string) (Summary, error)
	SaveGoalSummary(summary GoalSummary) error
	GetGoalSummary(goalID string) (GoalSummary, error)
	CreateHandoff(handoff Handoff) error
	GetHandoff(id string) (Handoff, error)
	RecordProposal(proposal ExperienceProposal) error
	GetProposals() ([]ExperienceProposal, error)
	Prune(policy RetentionPolicy) (int, error)
}

// FileProvider implements ExperienceProvider using JSON files on disk.
type FileProvider struct {
	rootDir string
	mu      sync.RWMutex
}

// NewFileProvider creates a FileProvider at the specified base directory.
func NewFileProvider(rootDir string) (*FileProvider, error) {
	if err := os.MkdirAll(filepath.Join(rootDir, "events"), 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "summaries"), 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "handoffs"), 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(rootDir, "proposals"), 0755); err != nil {
		return nil, err
	}
	return &FileProvider{rootDir: rootDir}, nil
}

func (p *FileProvider) RecordEvent(event SessionEvent) error {
	if err := event.Validate(); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	events, _ := p.getEventsUnlocked(event.SessionID)
	events = append(events, event)

	data, err := json.MarshalIndent(events, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(p.rootDir, "events", event.SessionID+".json"), data, 0644)
}

func (p *FileProvider) GetEvents(sessionID string) ([]SessionEvent, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.getEventsUnlocked(sessionID)
}

func (p *FileProvider) getEventsUnlocked(sessionID string) ([]SessionEvent, error) {
	file := filepath.Join(p.rootDir, "events", sessionID+".json")
	data, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return []SessionEvent{}, nil
	}
	if err != nil {
		return nil, err
	}
	var events []SessionEvent
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (p *FileProvider) SaveSummary(summary Summary) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(p.rootDir, "summaries", "session-"+summary.SessionID+".json"), data, 0644)
}

func (p *FileProvider) GetSummary(sessionID string) (Summary, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	file := filepath.Join(p.rootDir, "summaries", "session-"+sessionID+".json")
	data, err := os.ReadFile(file)
	if err != nil {
		return Summary{}, err
	}
	var s Summary
	if err := json.Unmarshal(data, &s); err != nil {
		return Summary{}, err
	}
	return s, nil
}

func (p *FileProvider) SaveGoalSummary(summary GoalSummary) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(p.rootDir, "summaries", "goal-"+summary.GoalID+".json"), data, 0644)
}

func (p *FileProvider) GetGoalSummary(goalID string) (GoalSummary, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	file := filepath.Join(p.rootDir, "summaries", "goal-"+goalID+".json")
	data, err := os.ReadFile(file)
	if err != nil {
		return GoalSummary{}, err
	}
	var gs GoalSummary
	if err := json.Unmarshal(data, &gs); err != nil {
		return GoalSummary{}, err
	}
	return gs, nil
}

func (p *FileProvider) CreateHandoff(handoff Handoff) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	data, err := json.MarshalIndent(handoff, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(p.rootDir, "handoffs", handoff.ID+".json"), data, 0644)
}

func (p *FileProvider) GetHandoff(id string) (Handoff, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	file := filepath.Join(p.rootDir, "handoffs", id+".json")
	data, err := os.ReadFile(file)
	if err != nil {
		return Handoff{}, fmt.Errorf("handoff %q not found", id)
	}
	var h Handoff
	if err := json.Unmarshal(data, &h); err != nil {
		return Handoff{}, err
	}
	return h, nil
}

func (p *FileProvider) RecordProposal(proposal ExperienceProposal) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	data, err := json.MarshalIndent(proposal, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(p.rootDir, "proposals", proposal.ID+".json"), data, 0644)
}

func (p *FileProvider) GetProposals() ([]ExperienceProposal, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	entries, err := os.ReadDir(filepath.Join(p.rootDir, "proposals"))
	if err != nil {
		return nil, err
	}
	var props []ExperienceProposal
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(p.rootDir, "proposals", entry.Name()))
		if err == nil {
			var prop ExperienceProposal
			if err := json.Unmarshal(data, &prop); err == nil {
				props = append(props, prop)
			}
		}
	}
	return props, nil
}

func (p *FileProvider) Prune(policy RetentionPolicy) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now().UTC()
	prunedCount := 0

	eventsDir := filepath.Join(p.rootDir, "events")
	files, err := os.ReadDir(eventsDir)
	if err != nil {
		return 0, err
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		filePath := filepath.Join(eventsDir, f.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		var events []SessionEvent
		if err := json.Unmarshal(data, &events); err != nil {
			continue
		}

		var retained []SessionEvent
		for _, ev := range events {
			if policy.ShouldRetainEvent(ev, now) {
				retained = append(retained, ev)
			} else {
				prunedCount++
			}
		}

		// Enforce MaxEventsPerSession
		if policy.MaxEventsPerSession > 0 && len(retained) > policy.MaxEventsPerSession {
			overflow := len(retained) - policy.MaxEventsPerSession
			prunedCount += overflow
			retained = retained[overflow:]
		}

		if len(retained) == 0 {
			_ = os.Remove(filePath)
		} else if len(retained) < len(events) {
			newData, _ := json.MarshalIndent(retained, "", "  ")
			_ = os.WriteFile(filePath, newData, 0644)
		}
	}

	return prunedCount, nil
}
