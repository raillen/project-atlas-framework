package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	docengine "github.com/raillen/project-atlas-framework/internal/documentation"
	"github.com/raillen/project-atlas-framework/internal/planning"
)

// sessionIDPattern constrains planning session ids to stable identifiers that
// are also safe as file names for the persisted checkpoint.
var sessionIDPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// DefaultSessionID returns the conventional id for a goal-focused session. It
// is deterministic so the interaction protocol can create-or-resume without
// extra state.
func DefaultSessionID(goal string) string {
	id := strings.TrimSpace(goal)
	if id == "" {
		return "PLAN"
	}
	if !sessionIDPattern.MatchString(id) {
		id = sanitizeSessionID(id)
	}
	return "PLAN-" + id
}

func sanitizeSessionID(value string) string {
	var b strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "PLAN"
	}
	return out
}

// sessionPath resolves the persisted checkpoint location for a session. The
// checkpoint is derived state (a Working Context artifact), so it lives under
// .ai/plan and is never a canonical project file.
func sessionPath(root, sessionID string) (string, error) {
	if !sessionIDPattern.MatchString(sessionID) {
		return "", fmt.Errorf("invalid planning session id: %s", sessionID)
	}
	return filepath.Join(root, ".ai", "plan", "sessions", sessionID+".json"), nil
}

// SaveSession persists the versioned checkpoint of a planning session.
func SaveSession(root string, session planning.PlanningSession) error {
	if err := session.Validate(); err != nil {
		return err
	}
	path, err := sessionPath(root, session.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(session.Checkpoint(), "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0644)
}

// LoadSession restores a planning session from its persisted checkpoint,
// recompiling it from the checkpoint's canonical structured state. It never
// depends on a raw transcript.
func LoadSession(root, sessionID string) (planning.PlanningSession, error) {
	path, err := sessionPath(root, sessionID)
	if err != nil {
		return planning.PlanningSession{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return planning.PlanningSession{}, fmt.Errorf("planning session not found: %s", sessionID)
		}
		return planning.PlanningSession{}, err
	}
	var cp planning.SessionCheckpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return planning.PlanningSession{}, err
	}
	return planning.Resume(cp, cp.Decisions, cp.Open)
}

// ListSessions returns every persisted planning session in deterministic id
// order so machine output is stable and golden-testable.
func ListSessions(root string) ([]planning.PlanningSession, error) {
	dir := filepath.Join(root, ".ai", "plan", "sessions")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []planning.PlanningSession{}, nil
		}
		return nil, err
	}
	sessions := []planning.PlanningSession{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		sessionID := strings.TrimSuffix(entry.Name(), ".json")
		if !sessionIDPattern.MatchString(sessionID) {
			continue
		}
		session, err := LoadSession(root, sessionID)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}
	sort.Slice(sessions, func(i, j int) bool { return sessions[i].ID < sessions[j].ID })
	return sessions, nil
}

// LatestSession returns the most recently updated session, breaking ties by id
// so the interaction protocol always has one deterministic default target.
func LatestSession(root string) (planning.PlanningSession, error) {
	sessions, err := ListSessions(root)
	if err != nil {
		return planning.PlanningSession{}, err
	}
	if len(sessions) == 0 {
		return planning.PlanningSession{}, fmt.Errorf("no planning session found")
	}
	latest := sessions[0]
	for _, s := range sessions[1:] {
		if s.UpdatedAt > latest.UpdatedAt || (s.UpdatedAt == latest.UpdatedAt && s.ID > latest.ID) {
			latest = s
		}
	}
	return latest, nil
}

// FocusSession returns the session for a goal scope, creating a blank one and
// persisting it when none exists yet. An explicit missing sessionID creates a
// new session under that id. This gives `atlas plan --goal <id>` deterministic
// create-or-resume semantics.
func FocusSession(root, goal, sessionID string) (planning.PlanningSession, error) {
	scope := "goal:" + strings.TrimSpace(goal)
	sessions, err := ListSessions(root)
	if err != nil {
		return planning.PlanningSession{}, err
	}
	if sessionID == "" {
		for _, s := range sessions {
			if s.Scope == scope {
				return s, nil
			}
		}
		sessionID = DefaultSessionID(goal)
	}
	for _, s := range sessions {
		if s.ID == sessionID {
			if s.Scope != scope {
				return planning.PlanningSession{}, fmt.Errorf("session %s is not scoped to goal %s", sessionID, goal)
			}
			return s, nil
		}
	}
	runID := sessionID
	session := planning.NewPlanningSession(sessionID, runID, scope, goal)
	if err := SaveSession(root, session); err != nil {
		return planning.PlanningSession{}, err
	}
	return session, nil
}

// LoadSessionBindings returns the canonical docs bindings for a root, treating
// an absent bindings file as an empty binding set so the interaction protocol
// can still compile context from structured planning state alone.
func LoadSessionBindings(root string) ([]docengine.Binding, error) {
	bindings, err := docengine.LoadBindings(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []docengine.Binding{}, nil
		}
		return nil, err
	}
	return bindings, nil
}

// SessionStatus is the machine-readable summary of a planning session used by
// `atlas plan status`. Every field is derived from structured state only.
type SessionStatus struct {
	SessionID        string `json:"session_id"`
	RunID            string `json:"run_id"`
	Scope            string `json:"scope"`
	Goal             string `json:"goal"`
	Decisions        int    `json:"decisions"`
	OpenQuestions    int    `json:"open_questions"`
	BlockersResolved int    `json:"blockers_resolved"`
	OpenRemaining    int    `json:"open_remaining"`
	UpdatedAt        string `json:"updated_at"`
	Resumable        bool   `json:"resumable"`
}

// StatusOf derives a SessionStatus from a planning session without compiling
// context, so status stays cheap and side-effect-free. Resumable reports only
// whether questions remain; the compiled manifest pressure is reported by
// resume.
func StatusOf(session planning.PlanningSession) SessionStatus {
	open := 0
	for _, q := range session.Open {
		if q.IsOpen() {
			open++
		}
	}
	return SessionStatus{
		SessionID:        session.ID,
		RunID:            session.RunID,
		Scope:            session.Scope,
		Goal:             session.Goal,
		Decisions:        len(session.Decisions),
		OpenQuestions:    open,
		BlockersResolved: session.LastPreview.BlockersResolved,
		OpenRemaining:    session.LastPreview.OpenRemaining,
		UpdatedAt:        session.UpdatedAt,
		Resumable:        open > 0,
	}
}
