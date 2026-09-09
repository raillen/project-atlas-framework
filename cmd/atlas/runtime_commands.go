package main

import (
	"fmt"
	"path/filepath"

	"github.com/raillen/project-atlas-framework/internal/protocol"
	"github.com/raillen/project-atlas-framework/internal/runtime"
)

func runRuntime(asJSON bool, args []string) int {
	if len(args) == 0 {
		return exitUsage
	}
	root := "."
	id := "R-bootstrap"
	for i := 1; i < len(args); i++ {
		if args[i] == "--path" && i+1 < len(args) {
			root = args[i+1]
			i++
		}
		if args[i] == "--run" && i+1 < len(args) {
			id = args[i+1]
			i++
		}
	}
	path := filepath.Join(root, ".atlas", "runtime", "continuation.json")
	switch args[0] {
	case "budget":
		data, err := runtime.LoadJSON[map[string]any](filepath.Join(root, ".atlas", "runtime", "budget.json"))
		if err != nil {
			return serviceError(asJSON, err)
		}
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(data))
		}
		fmt.Printf("%v\n", data)
		return exitOK
	case "run":
		if len(args) > 1 && (args[1] == "cancel" || args[1] == "resume") {
			run, err := runtime.LoadJSON[runtime.Run](filepath.Join(root, ".atlas", "runtime", "runs", id+".json"))
			if err != nil {
				return serviceError(asJSON, err)
			}
			if args[1] == "cancel" {
				run, err = run.Transition(runtime.RunCancelled)
			} else {
				run, err = run.Transition(runtime.RunRunning)
			}
			if err != nil {
				return serviceError(asJSON, err)
			}
			if err := runtime.SaveJSON(filepath.Join(root, ".atlas", "runtime", "runs", id+".json"), run); err != nil {
				return serviceError(asJSON, err)
			}
			if asJSON {
				return printEnvelope(protocol.OkEnvelope(run))
			}
			fmt.Printf("Run %s: %s\n", run.ID, run.Status)
			return exitOK
		}
		if len(args) > 1 && args[1] == "show" {
			record, err := runtime.LoadJSON[runtime.Run](filepath.Join(root, ".atlas", "runtime", "runs", id+".json"))
			if err != nil {
				return serviceError(asJSON, err)
			}
			if asJSON {
				return printEnvelope(protocol.OkEnvelope(record))
			}
			fmt.Printf("Run %s: %s\n", record.ID, record.Status)
			return exitOK
		}
		run := runtime.NewRun(id, "", "")
		if err := runtime.SaveJSON(filepath.Join(root, ".atlas", "runtime", "runs", id+".json"), run); err != nil {
			return serviceError(asJSON, err)
		}
		record := runtime.ContinuationFromRun(run, runtime.InspectRepository(root), nil, []string{"start work"}, nil, []string{"inspect Goal and repository state"})
		if err := runtime.SaveJSON(path, record); err != nil {
			return serviceError(asJSON, err)
		}
		if err := runtime.SaveJSON(filepath.Join(root, ".atlas", "runtime", "budget.json"), map[string]any{"version": 1, "scope": "run", "limits": map[string]any{"input_tokens": 8000, "output_tokens": 3000, "tool_calls": 20}, "usage": map[string]any{}, "reservations": []any{}, "mode": "soft"}); err != nil {
			return serviceError(asJSON, err)
		}
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(run))
		}
		fmt.Printf("Created Run %s\n", id)
		return exitOK
	case "continue":
		record, err := runtime.LoadJSON[runtime.ContinuationRecord](path)
		if err != nil {
			return serviceError(asJSON, err)
		}
		if hasFlag(args, "--prompt") {
			fmt.Print(runtime.RenderPrompt(record))
			return exitOK
		}
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(record))
		}
		fmt.Print(runtime.RenderPrompt(record))
		return exitOK
	default:
		return exitUsage
	}
}
