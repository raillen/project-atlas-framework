package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/raillen/project-atlas-framework/internal/automation"
	"github.com/raillen/project-atlas-framework/internal/packages"
	"github.com/raillen/project-atlas-framework/internal/protocol"
	"github.com/raillen/project-atlas-framework/internal/runtime"
)

func runPlatform(asJSON bool, args []string) int {
	if len(args) < 2 {
		return exitUsage
	}
	switch args[0] {
	case "package":
		lock, err := runtime.LoadJSON[packages.Lock](filepath.Join(".atlas", "lock.json"))
		if err != nil {
			return serviceError(asJSON, err)
		}
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(lock))
		}
		for _, p := range lock.Packages {
			fmt.Printf("%s %s %s\n", p.ID, p.Version, p.Type)
		}
		return exitOK
	case "runtime":
		entries, _ := os.ReadDir(filepath.Join(".atlas", "runtime"))
		names := []string{}
		for _, e := range entries {
			names = append(names, e.Name())
		}
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(map[string]any{"runtimes": names}))
		}
		fmt.Println(names)
		return exitOK
	case "automation":
		if len(args) > 1 && args[1] == "failures" {
			return printEnvelope(protocol.OkEnvelope([]automation.DLQEntry{}))
		}
		return printEnvelope(protocol.OkEnvelope([]any{}))
	default:
		return exitUsage
	}
}
