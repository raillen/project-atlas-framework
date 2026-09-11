// Package runtime implements the NativeAgent reentrant state machine (HA2).
// No monolithic for{model()} loop: Step advances one Phase at a time and
// every safe point can persist/cancel/pause/resume/handoff/replay.
package runtime

import (
	"context"
	"fmt"

	"github.com/raillen/prumo/internal/harness/agent"
	"github.com/raillen/prumo/internal/harness/checkpoint"
	"github.com/raillen/prumo/internal/harness/model"
	"github.com/raillen/prumo/internal/harness/perm"
)

// Services wires the ports the loop depends on (no vendor types).
type Services struct {
	Models      model.Provider
	Tools       ToolExecutor
	Perms       *perm.Engine
	Checkpoints *checkpoint.Store
	Events      func(agent.AgentEvent)
	// ContextManifest builds the context pointer for a turn.
	ContextManifest func(ctx context.Context, state agent.NativeAgentState) (string, error)
	// Budgets enforcement hook; nil disables.
	ConsumeBudget func(usage agent.Usage) error
}

// ToolExecutor executes one normalized ToolCall.
type ToolExecutor interface {
	Execute(ctx context.Context, call agent.ToolCall) (agent.ToolResult, error)
	KindOf(toolName string) string
}

// Runner holds mutable conversation buffers; canonical state stays in
// agent.NativeAgentState and is checkpointed at safe points.
type Runner struct {
	Svc      Services
	State    agent.NativeAgentState
	Messages []agent.Message
	Turn     agent.Turn
	Events   []agent.ModelEvent
	ToolQ    []agent.ToolCall
	Obs      []agent.Observation
	// AfterSideEffects flips true once an observable effect applies; the
	// gateway must then refuse transparent fallback.
	AfterSideEffects bool
	MaxTurns         int
	TurnsDone        int
	// QualityGate, when set, vets completion inside EvaluateStop: a failing
	// gate turns completion into PhaseFailed instead of Checkpoint.
	QualityGate func() error
}

// NewRunner initializes a Run session.
func NewRunner(svc Services, runID, sessionID string) *Runner {
	return &Runner{
		Svc:      svc,
		State:    agent.NativeAgentState{RunID: runID, SessionID: sessionID, TurnID: "turn-1", Phase: agent.PhasePrepare, Revision: 1, UpdatedAt: agent.Now()},
		Turn:     agent.Turn{ID: "turn-1", RunID: runID, SessionID: sessionID, Index: 1, Status: "open", StartedAt: agent.Now()},
		MaxTurns: 10,
	}
}

func (r *Runner) emit(kind string, payload map[string]any) {
	if r.Svc.Events == nil {
		return
	}
	r.Svc.Events(agent.AgentEvent{ID: fmt.Sprintf("ev-%d", len(payload)+1), RunID: r.State.RunID, TurnID: r.State.TurnID, Kind: kind, Payload: payload, CreatedAt: agent.Now()})
}

// Step advances exactly one Phase. Callers loop until Complete/Failed,
// checkpointing between steps to survive process restarts.
func (r *Runner) Step(ctx context.Context) error {
	switch r.State.Phase {
	case agent.PhasePrepare:
		r.State.Phase = agent.PhaseCompileContext
	case agent.PhaseCompileContext:
		id := "ctx-manifest"
		if r.Svc.ContextManifest != nil {
			var err error
			id, err = r.Svc.ContextManifest(ctx, r.State)
			if err != nil {
				return err
			}
		}
		r.State.ContextManifestID = id
		r.State.Phase = agent.PhaseRequestModel
	case agent.PhaseRequestModel:
		req := agent.ModelRequest{
			RequestID: fmt.Sprintf("%s:%s:req", r.State.RunID, r.State.TurnID),
			RunID:     r.State.RunID, TurnID: r.State.TurnID,
			Messages: r.Messages,
		}
		ch, err := r.Svc.Models.Stream(ctx, req)
		if err != nil {
			r.State.Phase = agent.PhaseFailed
			r.State.StopReason = err.Error()
			return err
		}
		r.Events = nil
		for ev := range ch {
			select {
			case <-ctx.Done():
				r.State.Phase = agent.PhaseYield
				r.State.StopReason = "cancelled"
				return ctx.Err()
			default:
			}
			r.Events = append(r.Events, ev)
			if ev.Kind == agent.EventUsageUpdated && ev.Usage != nil && r.Svc.ConsumeBudget != nil {
				if err := r.Svc.ConsumeBudget(*ev.Usage); err != nil {
					r.State.Phase = agent.PhaseFailed
					r.State.StopReason = "budget exhausted: " + err.Error()
					return err
				}
			}
			if ev.Kind == agent.EventError && !ev.Retryable {
				r.State.Phase = agent.PhaseFailed
				r.State.StopReason = ev.Error
				return fmt.Errorf("model error: %s", ev.Error)
			}
		}
		r.State.Phase = agent.PhaseConsumeModelEvent
	case agent.PhaseConsumeModelEvent:
		r.State.Phase = agent.PhasePlanToolCalls
	case agent.PhasePlanToolCalls:
		r.ToolQ = nil
		for _, ev := range r.Events {
			if ev.Kind == agent.EventToolCallReady && ev.ToolCall != nil {
				r.ToolQ = append(r.ToolQ, *ev.ToolCall)
			}
		}
		if len(r.ToolQ) == 0 {
			r.State.Phase = agent.PhaseEvaluateStop
		} else {
			r.State.Phase = agent.PhasePermissionCheck
		}
	case agent.PhasePermissionCheck:
		if r.Svc.Perms == nil {
			r.State.Phase = agent.PhaseExecuteTool
			break
		}
		for _, tc := range r.ToolQ {
			kind := ""
			if r.Svc.Tools != nil {
				kind = r.Svc.Tools.KindOf(tc.Name)
			}
			res := r.Svc.Perms.Evaluate(agent.PermissionRequest{
				ID: fmt.Sprintf("perm-%s", tc.ID), RunID: r.State.RunID, TurnID: r.State.TurnID,
				Action: tc.Name, Resource: fmt.Sprint(tc.Arguments["path"]),
			}, kind, "policy")
			if res.Decision == agent.PermissionDeny {
				r.State.Phase = agent.PhaseFailed
				r.State.StopReason = "permission denied: " + res.Reason
				r.emit("permission_denied", map[string]any{"tool": tc.Name, "reason": res.Reason})
				return fmt.Errorf("permission denied for %s: %s", tc.Name, res.Reason)
			}
			if res.Decision == agent.PermissionAsk {
				r.State.Phase = agent.PhaseYield
				r.State.StopReason = "permission wait: " + tc.ID
				r.State.PendingPerms = []string{res.RequestID}
				r.emit("permission_wait", map[string]any{"tool": tc.Name})
				return nil
			}
		}
		r.State.Phase = agent.PhaseExecuteTool
	case agent.PhaseExecuteTool:
		if r.Svc.Tools == nil {
			return fmt.Errorf("no tool executor")
		}
		for _, tc := range r.ToolQ {
			res, err := r.Svc.Tools.Execute(ctx, tc)
			if err != nil {
				return err
			}
			r.Obs = append(r.Obs, agent.Observation{ID: "obs-" + tc.ID, TurnID: r.State.TurnID, ToolCallID: tc.ID, Content: res.Output, CreatedAt: agent.Now()})
			if res.ExitCode == 0 {
				r.AfterSideEffects = true
			}
		}
		r.State.Phase = agent.PhaseRecordObservation
	case agent.PhaseRecordObservation:
		r.State.Phase = agent.PhaseEvaluateStop
	case agent.PhaseEvaluateStop:
		r.TurnsDone++
		if r.TurnsDone >= r.MaxTurns {
			r.State.Phase = agent.PhaseCheckpoint
			r.State.StopReason = "max turns reached"
			break
		}
		// Completion when last model events contained Completed and no tools.
		completed := false
		for _, ev := range r.Events {
			if ev.Kind == agent.EventCompleted {
				completed = true
			}
		}
		if completed && len(r.ToolQ) == 0 {
			if r.QualityGate != nil {
				if err := r.QualityGate(); err != nil {
					r.State.Phase = agent.PhaseFailed
					r.State.StopReason = err.Error()
					return err
				}
			}
			r.State.Phase = agent.PhaseCheckpoint
			r.State.StopReason = "completed"
		} else if len(r.ToolQ) > 0 {
			// Feed observations back as messages and continue.
			for _, o := range r.Obs {
				r.Messages = append(r.Messages, agent.Message{ID: o.ID, TurnID: o.TurnID, Role: agent.RoleTool, Content: o.Content, CreatedAt: agent.Now()})
			}
			r.Obs = nil
			r.State.Phase = agent.PhaseRequestModel
		} else {
			r.State.Phase = agent.PhaseCheckpoint
			r.State.StopReason = "no tool calls and no completion"
		}
	case agent.PhaseCheckpoint:
		if r.Svc.Checkpoints != nil {
			r.State.Revision++
			r.State.UpdatedAt = agent.Now()
			cp := agent.Checkpoint{ID: fmt.Sprintf("%s-r%d", r.State.RunID, r.State.Revision), RunID: r.State.RunID, State: r.State, CreatedAt: agent.Now()}
			if err := r.Svc.Checkpoints.Save(cp); err != nil {
				return err
			}
		}
		if r.State.StopReason == "completed" {
			r.State.Phase = agent.PhaseComplete
		} else if r.TurnsDone >= r.MaxTurns {
			r.State.Phase = agent.PhaseComplete
		} else {
			r.State.Phase = agent.PhaseYield
		}
	case agent.PhaseYield, agent.PhaseComplete, agent.PhaseFailed:
		return nil
	default:
		return fmt.Errorf("unknown phase %s", r.State.Phase)
	}
	r.State.UpdatedAt = agent.Now()
	return nil
}

// RunUntilDone steps until Complete/Failed/Yield or ctx cancel.
func (r *Runner) RunUntilDone(ctx context.Context) error {
	for i := 0; i < 1000; i++ {
		if r.State.Phase == agent.PhaseComplete || r.State.Phase == agent.PhaseFailed || r.State.Phase == agent.PhaseYield {
			return nil
		}
		if err := r.Step(ctx); err != nil {
			// Permission wait yields without error propagation beyond state.
			if r.State.Phase == agent.PhaseYield {
				return nil
			}
			return err
		}
	}
	return fmt.Errorf("runaway loop guard")
}
