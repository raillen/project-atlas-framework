package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/raillen/project-atlas-framework/internal/planning"
)

func staticSession(id, scope, goal, updatedAt string) planning.PlanningSession {
	return planning.PlanningSession{
		ID:        id,
		RunID:     "RUN-" + id,
		Scope:     scope,
		Goal:      goal,
		UpdatedAt: updatedAt,
	}
}

func TestSessionRoundTrip(t *testing.T) {
	root := t.TempDir()
	session := staticSession("S1", "goal:G001", "onboarding", "2026-09-09T10:00:00.000000000Z")
	if err := SaveSession(root, session); err != nil {
		t.Fatalf("SaveSession: %v", err)
	}
	loaded, err := LoadSession(root, "S1")
	if err != nil {
		t.Fatalf("LoadSession: %v", err)
	}
	if !reflect.DeepEqual(loaded, session) {
		t.Fatalf("round trip mismatch:\n got %#v\nwant %#v", loaded, session)
	}
}

func TestSessionMissing(t *testing.T) {
	root := t.TempDir()
	if _, err := LoadSession(root, "NOPE"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestSessionVersionRejected(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".ai", "plan", "sessions", "S1.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(map[string]any{
		"version": 99,
		"id":      "S1",
		"run_id":  "RUN-S1",
		"scope":   "goal:G001",
	})
	if err := os.WriteFile(path, append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSession(root, "S1"); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported-version error, got %v", err)
	}
}

func TestSaveSessionRejectsInvalidID(t *testing.T) {
	root := t.TempDir()
	session := staticSession("../escape", "goal:G001", "x", "2026-09-09T10:00:00.000000000Z")
	if err := SaveSession(root, session); err == nil {
		t.Fatal("expected invalid session id to be rejected")
	}
}

func TestListSessionsDeterministic(t *testing.T) {
	root := t.TempDir()
	if err := SaveSession(root, staticSession("S2", "goal:G002", "b", "2026-09-09T10:00:00.000000000Z")); err != nil {
		t.Fatal(err)
	}
	if err := SaveSession(root, staticSession("S1", "goal:G001", "a", "2026-09-09T09:00:00.000000000Z")); err != nil {
		t.Fatal(err)
	}
	sessions, err := ListSessions(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 || sessions[0].ID != "S1" || sessions[1].ID != "S2" {
		t.Fatalf("expected S1,S2 in id order, got %#v", sessions)
	}
	empty, err := ListSessions(t.TempDir())
	if err != nil || len(empty) != 0 {
		t.Fatalf("expected empty list for missing directory, got %#v, %v", empty, err)
	}
}

func TestLatestSession(t *testing.T) {
	root := t.TempDir()
	older := staticSession("S1", "goal:G001", "a", "2026-09-09T09:00:00.000000000Z")
	latest := staticSession("S2", "goal:G002", "b", "2026-09-09T12:00:00.000000000Z")
	if err := SaveSession(root, older); err != nil {
		t.Fatal(err)
	}
	if err := SaveSession(root, latest); err != nil {
		t.Fatal(err)
	}
	got, err := LatestSession(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != "S2" {
		t.Fatalf("expected S2 as latest, got %s", got.ID)
	}
	if _, err := LatestSession(t.TempDir()); err == nil || !strings.Contains(err.Error(), "no planning session") {
		t.Fatalf("expected no-session error, got %v", err)
	}
}

func TestFocusSessionCreateOrResume(t *testing.T) {
	root := t.TempDir()
	created, err := FocusSession(root, "G042", "")
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "PLAN-G042" || created.Scope != "goal:G042" || created.RunID != "PLAN-G042" {
		t.Fatalf("unexpected created session: %#v", created)
	}
	found, err := FocusSession(root, "G042", "")
	if err != nil {
		t.Fatal(err)
	}
	if found.ID != created.ID {
		t.Fatalf("expected create-or-resume to return same session, got %s", found.ID)
	}
	explicit, err := FocusSession(root, "G043", "S-CUSTOM")
	if err != nil {
		t.Fatal(err)
	}
	if explicit.ID != "S-CUSTOM" || explicit.Scope != "goal:G043" {
		t.Fatalf("unexpected explicit session: %#v", explicit)
	}
	if _, err := FocusSession(root, "G042", "S-CUSTOM"); err == nil || !strings.Contains(err.Error(), "not scoped") {
		t.Fatalf("expected scope mismatch error for explicitly requested other goal, got %v", err)
	}
}

func TestDefaultSessionID(t *testing.T) {
	cases := map[string]string{
		"G042":            "PLAN-G042",
		"goal with space": "PLAN-goal_with_space",
		"":                "PLAN",
	}
	for in, want := range cases {
		if got := DefaultSessionID(in); got != want {
			t.Fatalf("DefaultSessionID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLoadSessionBindingsMissingIsEmpty(t *testing.T) {
	root := t.TempDir()
	bindings, err := LoadSessionBindings(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(bindings) != 0 {
		t.Fatalf("expected empty bindings, got %#v", bindings)
	}
	withBindings := t.TempDir()
	path := filepath.Join(withBindings, "docs", "contracts", "bindings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`[{"contract_id":"c1","sources":["docs/a.md"],"ownership":"human","authority":"canonical"}]`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	bindings, err = LoadSessionBindings(withBindings)
	if err != nil || len(bindings) != 1 {
		t.Fatalf("expected one binding, got %#v, %v", bindings, err)
	}
}

func TestStatusOf(t *testing.T) {
	session := staticSession("S1", "goal:G042", "G042", "2026-09-09T10:00:00.000000000Z")
	session.Decisions = []planning.DecisionProposal{{ID: "DP-1"}}
	session.Open = []planning.OpenQuestion{
		{ID: "Q1", Scope: "goal:G042", Priority: "blocker", Status: "open"},
	}
	session.LastPreview.BlockersResolved = 1
	session.LastPreview.OpenRemaining = 1
	status := StatusOf(session)
	if status.SessionID != "S1" || status.Decisions != 1 || status.OpenQuestions != 1 ||
		status.BlockersResolved != 1 || status.OpenRemaining != 1 || !status.Resumable {
		t.Fatalf("unexpected status: %#v", status)
	}
	done := StatusOf(staticSession("S2", "goal:G042", "G042", "2026-09-09T10:00:00.000000000Z"))
	if done.Resumable {
		t.Fatal("expected empty session to be not resumable")
	}
}
