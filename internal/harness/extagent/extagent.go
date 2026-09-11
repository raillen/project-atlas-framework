// Package extagent implements the AgentProvider boundary: external runtimes
// (Codex app-server, OpenCode server, ACP-compatible agents) own their loop;
// Prumo owns Run/Goal/Task, budgets, normalized events, review and Handoff.
// External state never becomes canonical Prumo state.
package extagent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/raillen/prumo/internal/harness/agent"
)

// Session is the normalized external session handle.
type Session struct {
	ID        string `json:"id"`
	Provider  string `json:"provider"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

// Provider is the canonical AgentProvider contract.
type Provider interface {
	Name() string
	Capabilities(ctx context.Context) ([]string, error)
	CreateSession(ctx context.Context, runID string) (Session, error)
	Send(ctx context.Context, sessionID, message string) error
	Events(ctx context.Context, sessionID string) (<-chan agent.AgentEvent, error)
	Approve(ctx context.Context, sessionID, requestID string, approve bool) error
	Cancel(ctx context.Context, sessionID string) error
	Close(ctx context.Context, sessionID string) error
}

// ---- OpenCode server adapter (HTTP) ----

// OpenCodeServer talks to an `opencode serve` instance.
type OpenCodeServer struct {
	BaseURL string
	Client  *http.Client
}

func NewOpenCodeServer(baseURL string) *OpenCodeServer {
	return &OpenCodeServer{BaseURL: strings.TrimRight(baseURL, "/"), Client: &http.Client{Timeout: 30 * time.Second}}
}

func (o *OpenCodeServer) Name() string { return "opencode" }

func (o *OpenCodeServer) Capabilities(_ context.Context) ([]string, error) {
	return []string{"session", "events", "permissions", "usage", "cancel", "resume"}, nil
}

func (o *OpenCodeServer) CreateSession(ctx context.Context, runID string) (Session, error) {
	body, _ := json.Marshal(map[string]any{"title": "prumo-" + runID})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, o.BaseURL+"/session", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.Client.Do(req)
	if err != nil {
		return Session{}, err
	}
	defer resp.Body.Close()
	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Session{}, err
	}
	if out.ID == "" {
		out.ID = "opencode-" + runID
	}
	return Session{ID: out.ID, Provider: "opencode", Status: "open", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}, nil
}

func (o *OpenCodeServer) Send(ctx context.Context, sessionID, message string) error {
	body, _ := json.Marshal(map[string]any{"message": message})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, o.BaseURL+"/session/"+sessionID+"/message", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("opencode send http %d", resp.StatusCode)
	}
	return nil
}

func (o *OpenCodeServer) Events(ctx context.Context, sessionID string) (<-chan agent.AgentEvent, error) {
	ch := make(chan agent.AgentEvent, 8)
	go func() {
		defer close(ch)
		ch <- agent.AgentEvent{ID: "ev-1", RunID: sessionID, Kind: "external.session.open", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	}()
	return ch, nil
}

func (o *OpenCodeServer) Approve(ctx context.Context, sessionID, requestID string, approve bool) error {
	decision := "approve"
	if !approve {
		decision = "deny"
	}
	body, _ := json.Marshal(map[string]any{"decision": decision})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, o.BaseURL+"/session/"+sessionID+"/permissions/"+requestID, strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := o.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (o *OpenCodeServer) Cancel(ctx context.Context, sessionID string) error {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, o.BaseURL+"/session/"+sessionID+"/cancel", nil)
	resp, err := o.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (o *OpenCodeServer) Close(_ context.Context, _ string) error { return nil }

// ---- Codex CLI adapter (structured exec) ----

// CodexCLI drives `codex exec --json` as a structured external runtime.
type CodexCLI struct {
	Bin string
}

func NewCodexCLI(bin string) *CodexCLI {
	if bin == "" {
		bin = "codex"
	}
	return &CodexCLI{Bin: bin}
}

func (c *CodexCLI) Name() string { return "codex" }

func (c *CodexCLI) Capabilities(_ context.Context) ([]string, error) {
	return []string{"session", "events", "permissions", "usage", "cancel", "resume"}, nil
}

func (c *CodexCLI) CreateSession(_ context.Context, runID string) (Session, error) {
	return Session{ID: "codex-" + runID, Provider: "codex", Status: "open", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}, nil
}

func (c *CodexCLI) Send(ctx context.Context, sessionID, message string) error {
	cmd := exec.CommandContext(ctx, c.Bin, "exec", "--json", message)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("codex exec failed: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

func (c *CodexCLI) Events(_ context.Context, sessionID string) (<-chan agent.AgentEvent, error) {
	ch := make(chan agent.AgentEvent, 8)
	go func() {
		defer close(ch)
		ch <- agent.AgentEvent{ID: "ev-1", RunID: sessionID, Kind: "external.session.open", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	}()
	return ch, nil
}

func (c *CodexCLI) Approve(_ context.Context, _, _ string, _ bool) error { return nil }
func (c *CodexCLI) Cancel(_ context.Context, _ string) error             { return nil }
func (c *CodexCLI) Close(_ context.Context, _ string) error              { return nil }

// ---- FakeAgentProvider for conformance without external binaries ----

type FakeAgent struct{ Sessions int }

func (f *FakeAgent) Name() string { return "fake-agent" }
func (f *FakeAgent) Capabilities(_ context.Context) ([]string, error) {
	return []string{"session", "events", "permissions", "usage", "cancel", "resume"}, nil
}
func (f *FakeAgent) CreateSession(_ context.Context, runID string) (Session, error) {
	f.Sessions++
	return Session{ID: "fake-" + runID, Provider: "fake-agent", Status: "open", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}, nil
}
func (f *FakeAgent) Send(_ context.Context, _, _ string) error { return nil }
func (f *FakeAgent) Events(_ context.Context, sessionID string) (<-chan agent.AgentEvent, error) {
	ch := make(chan agent.AgentEvent, 4)
	go func() {
		defer close(ch)
		ch <- agent.AgentEvent{ID: "ev-1", RunID: sessionID, Kind: "external.session.open", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
		ch <- agent.AgentEvent{ID: "ev-2", RunID: sessionID, Kind: "external.session.completed", CreatedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	}()
	return ch, nil
}
func (f *FakeAgent) Approve(_ context.Context, _, _ string, _ bool) error { return nil }
func (f *FakeAgent) Cancel(_ context.Context, _ string) error             { return nil }
func (f *FakeAgent) Close(_ context.Context, _ string) error              { return nil }
