package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	docengine "github.com/raillen/prumo/internal/documentation"
	"github.com/raillen/prumo/internal/protocol"
)

func runDocumentation(asJSON bool, args []string) int {
	if len(args) < 2 || args[0] != "docs" {
		fmt.Fprintln(os.Stderr, "error: expected docs <contracts|profiles|audit|readiness>")
		return exitUsage
	}
	root := "."
	goal := ""
	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "--path":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --path requires a value")
				return exitUsage
			}
			root = args[i+1]
			i++
		case "--goal":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "error: --goal requires a value")
				return exitUsage
			}
			goal = args[i+1]
			i++
		}
	}
	registry, err := docengine.LoadRegistry(repoRoot())
	if err != nil {
		return serviceError(asJSON, err)
	}
	var result any
	switch args[1] {
	case "contracts":
		if len(args) > 2 && args[2] == "show" {
			if len(args) < 4 {
				fmt.Fprintln(os.Stderr, "error: docs contracts show requires id")
				return exitUsage
			}
			contract, ok := registry.Contracts[args[3]]
			if !ok {
				return serviceError(asJSON, fmt.Errorf("unknown documentation contract: %s", args[3]))
			}
			result = contract
		} else {
			ids := []string{}
			for id := range registry.Contracts {
				ids = append(ids, id)
			}
			sortStrings(ids)
			result = ids
		}
	case "profiles":
		if len(args) > 2 && args[2] == "show" {
			if len(args) < 4 {
				fmt.Fprintln(os.Stderr, "error: docs profiles show requires id")
				return exitUsage
			}
			profile, ok := registry.Profiles[args[3]]
			if !ok {
				return serviceError(asJSON, fmt.Errorf("unknown documentation profile: %s", args[3]))
			}
			result = profile
		} else {
			ids := []string{}
			for id := range registry.Profiles {
				ids = append(ids, id)
			}
			sortStrings(ids)
			result = ids
		}
	case "audit":
		report, err := docengine.Audit(root)
		if err != nil {
			return serviceError(asJSON, err)
		}
		result = report
	case "readiness":
		report, err := docengine.Readiness(root, goal)
		if err != nil {
			return serviceError(asJSON, err)
		}
		result = report
	case "impact":
		changed := []string{}
		if len(args) > 2 && args[2] != "--path" {
			changed = append(changed, args[2:]...)
		}
		impact, err := docengine.AnalyzeImpact(root, changed)
		if err != nil {
			return serviceError(asJSON, err)
		}
		result = impact
	case "delta":
		return runDocsDelta(asJSON, args, root, goal)
	case "contradictions":
		findings, err := docengine.DetectContradictions(root)
		if err != nil {
			return serviceError(asJSON, err)
		}
		result = findings
	default:
		fmt.Fprintf(os.Stderr, "error: unsupported docs command: %s\n", args[1])
		return exitUsage
	}
	if asJSON {
		return printEnvelope(protocol.OkEnvelope(result))
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	return exitOK
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

func documentationRoot(path string) string {
	if path == "" {
		return "."
	}
	return filepath.Clean(path)
}
