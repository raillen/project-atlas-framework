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
	case "run":
		run := runtime.NewRun(id, "", "")
		if err := runtime.SaveJSON(filepath.Join(root, ".atlas", "runtime", "runs", id+".json"), run); err != nil {
			return serviceError(asJSON, err)
		}
		record := runtime.ContinuationFromRun(run, runtime.RepositoryState{}, nil, []string{"start work"}, nil, []string{"inspect Goal and repository state"})
		if err := runtime.SaveJSON(path, record); err != nil {
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
