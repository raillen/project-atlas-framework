package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type RunStatus string

const (
	RunCreated         RunStatus = "created"
	RunPlanning        RunStatus = "planning"
	RunReady           RunStatus = "ready"
	RunRunning         RunStatus = "running"
	RunWaitingApproval RunStatus = "waiting_approval"
	RunWaitingResource RunStatus = "waiting_resource"
	RunRetrying        RunStatus = "retrying"
	RunVerifying       RunStatus = "verifying"
	RunBlocked         RunStatus = "blocked"
	RunCompleted       RunStatus = "completed"
	RunFailed          RunStatus = "failed"
	RunCancelled       RunStatus = "cancelled"
)

type Run struct {
	ID                 string    `json:"id"`
	Version            int       `json:"version"`
	GoalID             string    `json:"goal_id,omitempty"`
	TaskID             string    `json:"task_id,omitempty"`
	Status             RunStatus `json:"status"`
	CreatedAt          string    `json:"created_at"`
	StartedAt          string    `json:"started_at,omitempty"`
	FinishedAt         string    `json:"finished_at,omitempty"`
	Sessions           []string  `json:"sessions,omitempty"`
	LatestCheckpoint   string    `json:"latest_checkpoint,omitempty"`
	Evidence           []string  `json:"evidence,omitempty"`
	Failure            string    `json:"failure,omitempty"`
	PendingSideEffects []string  `json:"pending_side_effects,omitempty"`
}

type ExecutorSession struct {
	ID          string `json:"id"`
	Version     int    `json:"version"`
	RunID       string `json:"run_id"`
	StartedAt   string `json:"started_at"`
	EndedAt     string `json:"ended_at,omitempty"`
	Harness     string `json:"harness,omitempty"`
	Model       string `json:"model,omitempty"`
	EndReason   string `json:"end_reason,omitempty"`
	Status      string `json:"status"`
	Predecessor string `json:"predecessor,omitempty"`
	Successor   string `json:"successor,omitempty"`
}

type Checkpoint struct {
	ID                 string         `json:"id"`
	Version            int            `json:"version"`
	RunID              string         `json:"run_id"`
	Phase              string         `json:"phase"`
	CreatedAt          string         `json:"created_at"`
	WorkspaceRef       string         `json:"workspace_ref,omitempty"`
	CompletedSteps     []string       `json:"completed_steps"`
	PendingSteps       []string       `json:"pending_steps"`
	Evidence           []string       `json:"evidence,omitempty"`
	PendingSideEffects []string       `json:"pending_side_effects,omitempty"`
	Budget             map[string]any `json:"budget,omitempty"`
	ContextSources     []string       `json:"context_sources,omitempty"`
	ResumeToken        string         `json:"resume_token,omitempty"`
}

type ContinuationRecord struct {
	Version            int             `json:"version"`
	Project            string          `json:"project,omitempty"`
	Goal               string          `json:"goal,omitempty"`
	Task               string          `json:"task,omitempty"`
	Run                string          `json:"run"`
	State              string          `json:"state"`
	Repository         RepositoryState `json:"repository"`
	Executor           map[string]any  `json:"executor,omitempty"`
	Completed          []string        `json:"completed"`
	Current            []string        `json:"current"`
	Pending            []string        `json:"pending"`
	Decisions          []string        `json:"decisions,omitempty"`
	Files              map[string]any  `json:"files,omitempty"`
	Evidence           []string        `json:"evidence,omitempty"`
	Blockers           []string        `json:"blockers,omitempty"`
	PendingSideEffects []string        `json:"pending_side_effects,omitempty"`
	NextSteps          []string        `json:"next_steps"`
	ContextSources     []string        `json:"context_sources,omitempty"`
}

type RepositoryState struct {
	Branch   string `json:"branch"`
	Revision string `json:"revision"`
	Dirty    bool   `json:"dirty"`
}

var transitions = map[RunStatus]map[RunStatus]bool{RunCreated: {RunPlanning: true, RunCancelled: true}, RunPlanning: {RunReady: true, RunBlocked: true, RunCancelled: true}, RunReady: {RunRunning: true, RunCancelled: true}, RunRunning: {RunWaitingApproval: true, RunWaitingResource: true, RunRetrying: true, RunVerifying: true, RunBlocked: true, RunCompleted: true, RunFailed: true, RunCancelled: true}, RunRetrying: {RunRunning: true, RunFailed: true, RunCancelled: true}, RunVerifying: {RunCompleted: true, RunFailed: true, RunBlocked: true, RunCancelled: true}, RunWaitingApproval: {RunRunning: true, RunBlocked: true, RunCancelled: true}, RunWaitingResource: {RunRunning: true, RunBlocked: true, RunCancelled: true}, RunBlocked: {RunPlanning: true, RunCancelled: true}}

func NewRun(id, goal, task string) Run {
	return Run{ID: id, Version: 1, GoalID: goal, TaskID: task, Status: RunCreated, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
}
func (r Run) Transition(next RunStatus) (Run, error) {
	if !transitions[r.Status][next] {
		return r, fmt.Errorf("invalid run transition: %s -> %s", r.Status, next)
	}
	r.Status = next
	if next == RunRunning && r.StartedAt == "" {
		r.StartedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	if next == RunCompleted || next == RunFailed || next == RunCancelled {
		r.FinishedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	return r, nil
}
func (r Run) AttachSession(session ExecutorSession) Run {
	r.Sessions = append(r.Sessions, session.ID)
	return r
}
func NewSession(id, runID, harness, model string) ExecutorSession {
	return ExecutorSession{ID: id, Version: 1, RunID: runID, StartedAt: time.Now().UTC().Format(time.RFC3339Nano), Harness: harness, Model: model, Status: "active"}
}
func (s ExecutorSession) End(reason string) ExecutorSession {
	s.Status = "ended"
	s.EndReason = reason
	s.EndedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return s
}
func NewCheckpoint(id, runID, phase string, completed, pending []string) Checkpoint {
	return Checkpoint{ID: id, Version: 1, RunID: runID, Phase: phase, CreatedAt: time.Now().UTC().Format(time.RFC3339Nano), CompletedSteps: completed, PendingSteps: pending, ResumeToken: id}
}
func ContinuationFromRun(run Run, repo RepositoryState, completed, current, pending, next []string) ContinuationRecord {
	return ContinuationRecord{Version: 1, Project: "atlas", Goal: run.GoalID, Task: run.TaskID, Run: run.ID, State: string(run.Status), Repository: repo, Completed: completed, Current: current, Pending: pending, NextSteps: next, PendingSideEffects: run.PendingSideEffects, Evidence: run.Evidence}
}

func SaveJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
func LoadJSON[T any](path string) (T, error) {
	var value T
	data, err := os.ReadFile(path)
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(data, &value)
	return value, err
}
func SortStrings(values []string) []string {
	out := append([]string{}, values...)
	sort.Strings(out)
	return out
}
func RenderPrompt(record ContinuationRecord) string {
	var b strings.Builder
	b.WriteString("Atlas Portable Continuation\n\n")
	fmt.Fprintf(&b, "Run: %s\nState: %s\n", record.Run, record.State)
	if record.Goal != "" {
		fmt.Fprintf(&b, "Goal: %s\n", record.Goal)
	}
	if record.Repository.Branch != "" {
		fmt.Fprintf(&b, "Branch: %s @ %s (dirty=%t)\n", record.Repository.Branch, record.Repository.Revision, record.Repository.Dirty)
	}
	b.WriteString("\nCompleted:\n")
	for _, v := range record.Completed {
		fmt.Fprintf(&b, "- %s\n", v)
	}
	b.WriteString("\nCurrent:\n")
	for _, v := range record.Current {
		fmt.Fprintf(&b, "- %s\n", v)
	}
	b.WriteString("\nPending:\n")
	for _, v := range record.Pending {
		fmt.Fprintf(&b, "- %s\n", v)
	}
	b.WriteString("\nNext steps:\n")
	for _, v := range record.NextSteps {
		fmt.Fprintf(&b, "- %s\n", v)
	}
	if len(record.PendingSideEffects) > 0 {
		b.WriteString("\nPending side effects require reconciliation before replay:\n")
		for _, v := range record.PendingSideEffects {
			fmt.Fprintf(&b, "- %s\n", v)
		}
	}
	return b.String()
}
