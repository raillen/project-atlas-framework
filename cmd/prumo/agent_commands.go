// Command: prumo agent — headless harness entrypoint (HA2..HA11).
//
// Subcommands:
//   run      --goal <text> --path <dir> [--provider fake|openai-compat] [--model ...] [--max-turns N]
//   resume   --run <id> --path <dir>
//   handoff  --run <id> --from native --to <agent> [--path <dir>]
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/raillen/prumo/internal/harness/aci"
	"github.com/raillen/prumo/internal/harness/agent"
	"github.com/raillen/prumo/internal/harness/checkpoint"
	"github.com/raillen/prumo/internal/harness/handoff"
	"github.com/raillen/prumo/internal/harness/model"
	"github.com/raillen/prumo/internal/harness/perm"
	harnessruntime "github.com/raillen/prumo/internal/harness/runtime"
	"github.com/raillen/prumo/internal/protocol"
)

func runAgent(asJSON bool, args []string) int {
	if len(args) == 0 {
		return exitUsage
	}
	switch args[0] {
	case "run":
		return runAgentRun(asJSON, args[1:])
	case "resume":
		return runAgentResume(asJSON, args[1:])
	case "handoff":
		return runAgentHandoff(asJSON, args[1:])
	default:
		return exitUsage
	}
}

func agentFlags(args []string) map[string]string {
	out := map[string]string{}
	for i := 0; i < len(args); i++ {
		if len(args[i]) > 2 && args[i][:2] == "--" && i+1 < len(args) {
			out[args[i][2:]] = args[i+1]
			i++
		}
	}
	return out
}

func runAgentRun(asJSON bool, args []string) int {
	f := agentFlags(args)
	goal := f["goal"]
	if goal == "" {
		goal = "headless run"
	}
	root := f["path"]
	if root == "" {
		root = "."
	}
	providerName := f["provider"]
	if providerName == "" {
		providerName = "fake"
	}
	modelName := f["model"]
	baseURL := f["base-url"]
	apiKey := os.Getenv("PRUMO_MODEL_API_KEY")
	maxTurns := 5
	if v, ok := f["max-turns"]; ok {
		fmt.Sscanf(v, "%d", &maxTurns)
	}
	runID := f["run"]
	if runID == "" {
		runID = "R-agent-1"
	}

	var provider model.Provider
	switch providerName {
	case "fake":
		provider = model.NewFake(map[string][]model.ScriptStep{
			"*": {
				{Kind: "text", Text: "goal: " + goal},
				{Kind: "tool_call", Tool: &agent.ToolCall{ID: "c1", TurnID: "turn-1", Name: "git.status", IdempotencyKey: runID + ":c1"}},
				{Kind: "complete"},
			},
		})
	case "openai-compat":
		if baseURL == "" {
			baseURL = os.Getenv("PRUMO_MODEL_BASE_URL")
		}
		if baseURL == "" {
			return serviceError(asJSON, fmt.Errorf("openai-compat requires --base-url or PRUMO_MODEL_BASE_URL"))
		}
		provider = model.NewOpenAICompat(baseURL, apiKey, modelName)
	default:
		return serviceError(asJSON, fmt.Errorf("unknown provider %s (fake|openai-compat)", providerName))
	}

	dir := filepath.Join(root, ".prumo", "runtime", "harness")
	runner := harnessruntime.NewRunner(harnessruntime.Services{
		Models:      provider,
		Tools:       aci.New(root),
		Perms:       perm.New(perm.Policy{DefaultAction: agent.PermissionAllow, DenyPrefixes: []string{"/etc", ".."}, AskKinds: []string{"destructive"}}),
		Checkpoints: checkpoint.New(dir),
		Events: func(ev agent.AgentEvent) {
			if !asJSON {
				fmt.Fprintf(os.Stderr, "[%s] %s\n", ev.Kind, ev.TurnID)
			}
		},
		ContextManifest: func(_ context.Context, _ agent.NativeAgentState) (string, error) {
			return "ctx-" + runID, nil
		},
	}, runID, "S-1")
	runner.MaxTurns = maxTurns
	runner.Messages = []agent.Message{{ID: "m1", Role: agent.RoleUser, Content: goal, CreatedAt: agent.Now()}}
	if err := runner.RunUntilDone(context.Background()); err != nil {
		return serviceError(asJSON, err)
	}
	result := map[string]any{"run_id": runID, "phase": string(runner.State.Phase), "stop_reason": runner.State.StopReason, "revision": runner.State.Revision, "turns": runner.TurnsDone}
	if asJSON {
		return printEnvelope(protocol.OkEnvelope(result))
	}
	fmt.Printf("Run %s: %s (%s)\n", runID, runner.State.Phase, runner.State.StopReason)
	return exitOK
}

func runAgentResume(asJSON bool, args []string) int {
	f := agentFlags(args)
	root := f["path"]
	if root == "" {
		root = "."
	}
	runID := f["run"]
	if runID == "" {
		return serviceError(asJSON, fmt.Errorf("resume requires --run <id>"))
	}
	dir := filepath.Join(root, ".prumo", "runtime", "harness")
	store := checkpoint.New(dir)
	cp, err := store.Latest(runID)
	if err != nil {
		return serviceError(asJSON, err)
	}
	provider := model.NewFake(map[string][]model.ScriptStep{"*": {{Kind: "text", Text: "resumed"}, {Kind: "complete"}}})
	runner := harnessruntime.NewRunner(harnessruntime.Services{
		Models: provider, Tools: aci.New(root),
		Perms: perm.New(perm.Policy{DefaultAction: agent.PermissionAllow}),
		Checkpoints: store,
	}, runID, cp.State.SessionID)
	runner.State = cp.State
	// Step once past the saved safe point toward completion.
	if runner.State.Phase == agent.PhaseCheckpoint || runner.State.Phase == agent.PhaseYield {
		runner.State.Phase = agent.PhaseComplete
		if runner.State.StopReason == "" {
			runner.State.StopReason = "resumed"
		}
	}
	result := map[string]any{"run_id": runID, "resumed_from": cp.ID, "phase": string(runner.State.Phase)}
	if asJSON {
		return printEnvelope(protocol.OkEnvelope(result))
	}
	fmt.Printf("Resumed %s from %s: %s\n", runID, cp.ID, runner.State.Phase)
	return exitOK
}

func runAgentHandoff(asJSON bool, args []string) int {
	f := agentFlags(args)
	root := f["path"]
	if root == "" {
		root = "."
	}
	runID := f["run"]
	if runID == "" {
		return serviceError(asJSON, fmt.Errorf("handoff requires --run <id>"))
	}
	from := f["from"]
	if from == "" {
		from = "native"
	}
	to := f["to"]
	if to == "" {
		to = "external"
	}
	dir := filepath.Join(root, ".prumo", "runtime", "harness")
	store := checkpoint.New(dir)
	cp, err := store.Latest(runID)
	if err != nil {
		// Allow handoff from live state when no checkpoint exists yet.
		cp = agent.Checkpoint{ID: runID + "-live", RunID: runID, State: agent.NativeAgentState{RunID: runID, ContextManifestID: "ctx-" + runID}}
	}
	b, err := handoff.Build(from, to, cp.State, "rev-live", "handoff for "+runID, nil)
	if err != nil {
		return serviceError(asJSON, err)
	}
	if err := b.Validate(); err != nil {
		return serviceError(asJSON, err)
	}
	if asJSON {
		return printEnvelope(protocol.OkEnvelope(b))
	}
	fmt.Printf("Handoff %s: %s -> %s (checkpoint %s)\n", b.Handoff.ID, from, to, b.Refs["checkpoint"])
	return exitOK
}
