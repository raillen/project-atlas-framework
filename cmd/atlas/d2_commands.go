package main

import (
	"fmt"
	"strconv"

	"github.com/raillen/project-atlas-framework/internal/environment"
	"github.com/raillen/project-atlas-framework/internal/modelregistry"
	"github.com/raillen/project-atlas-framework/internal/protocol"
	"github.com/raillen/project-atlas-framework/internal/toolgateway"
)

func defaultToolRegistry() *toolgateway.Registry {
	r := toolgateway.NewRegistry()
	r.Register(toolgateway.Descriptor{
		ID:              "read_file",
		Version:         1,
		Description:     "Read file content within project root",
		Kind:            toolgateway.ReadOnly,
		Trust:           "core",
		Capabilities:    []string{"fs", "read"},
		FilesystemScope: []string{"project-root"},
		TimeoutMS:       5000,
		OutputLimit:     102400,
	})
	r.Register(toolgateway.Descriptor{
		ID:              "write_file",
		Version:         1,
		Description:     "Write or modify file within project root",
		Kind:            toolgateway.SideEffecting,
		Trust:           "core",
		Capabilities:    []string{"fs", "write"},
		FilesystemScope: []string{"project-root"},
		TimeoutMS:       10000,
		OutputLimit:     102400,
	})
	r.Register(toolgateway.Descriptor{
		ID:              "exec_command",
		Version:         1,
		Description:     "Execute shell command in project directory",
		Kind:            toolgateway.SideEffecting,
		Trust:           "trusted",
		Capabilities:    []string{"exec", "shell"},
		FilesystemScope: []string{"project-root"},
		TimeoutMS:       60000,
		OutputLimit:     204800,
	})
	r.Register(toolgateway.Descriptor{
		ID:              "git_commit",
		Version:         1,
		Description:     "Create a Git commit according to repository policy",
		Kind:            toolgateway.SideEffecting,
		Trust:           "core",
		Capabilities:    []string{"scm", "git"},
		FilesystemScope: []string{"project-root"},
		TimeoutMS:       30000,
		OutputLimit:     51200,
	})
	return r
}

func defaultModelRegistry() *modelregistry.Registry {
	r := modelregistry.NewRegistry()
	r.Register(modelregistry.Descriptor{
		ID:               "local-default",
		Version:          1,
		Provider:         "local",
		Privacy:          modelregistry.Local,
		ContextTokens:    8192,
		StructuredOutput: true,
		ToolUse:          true,
		LatencyClass:     "fast",
	})
	r.Register(modelregistry.Descriptor{
		ID:               "cloud-standard",
		Version:          1,
		Provider:         "trusted-cloud",
		Privacy:          modelregistry.Trusted,
		ContextTokens:    32768,
		StructuredOutput: true,
		ToolUse:          true,
		LatencyClass:     "standard",
	})
	r.Register(modelregistry.Descriptor{
		ID:               "cloud-deep",
		Version:          1,
		Provider:         "external-cloud",
		Privacy:          modelregistry.External,
		ContextTokens:    65536,
		StructuredOutput: true,
		ToolUse:          true,
		LatencyClass:     "deep",
	})
	return r
}

func runTool(asJSON bool, root string, args []string) int {
	if len(args) == 0 {
		return exitUsage
	}
	reg := defaultToolRegistry()
	switch args[0] {
	case "list":
		tools := reg.List()
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(tools))
		}
		fmt.Printf("%-15s %-15s %-8s %s\n", "ID", "KIND", "TRUST", "DESCRIPTION")
		fmt.Println("---------------------------------------------------------------")
		for _, t := range tools {
			fmt.Printf("%-15s %-15s %-8s %s\n", t.ID, t.Kind, t.Trust, t.Description)
		}
		return exitOK
	case "inspect":
		if len(args) < 2 {
			fmt.Println("error: tool inspect requires tool ID")
			return exitUsage
		}
		tool, ok := reg.Get(args[1])
		if !ok {
			fmt.Printf("error: tool not found: %s\n", args[1])
			return exitValidation
		}
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(tool))
		}
		fmt.Printf("Tool: %s (version: %d)\nKind: %s\nTrust: %s\nCapabilities: %v\n",
			tool.ID, tool.Version, tool.Kind, tool.Trust, tool.Capabilities)
		return exitOK
	case "evaluate":
		if len(args) < 2 {
			fmt.Println("error: tool evaluate requires tool ID")
			return exitUsage
		}
		tool, ok := reg.Get(args[1])
		if !ok {
			fmt.Printf("error: tool not found: %s\n", args[1])
			return exitValidation
		}
		target := ""
		if len(args) > 2 {
			target = args[2]
		}
		safe := hasFlag(args, "--safe")
		dec := toolgateway.Evaluate(tool, root, target, safe)
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(dec))
		}
		if dec.Allowed {
			fmt.Printf("Allowed: %s\n", dec.Reason)
			return exitOK
		}
		fmt.Printf("Denied: %s\n", dec.Reason)
		return exitValidation
	default:
		return exitUsage
	}
}

func runModel(asJSON bool, root string, args []string) int {
	if len(args) == 0 {
		return exitUsage
	}
	reg := defaultModelRegistry()
	switch args[0] {
	case "list":
		models := reg.List()
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(models))
		}
		fmt.Printf("%-16s %-12s %-10s %-8s %s\n", "ID", "PROVIDER", "PRIVACY", "CONTEXT", "LATENCY")
		fmt.Println("-----------------------------------------------------------------")
		for _, m := range models {
			fmt.Printf("%-16s %-12s %-10s %-8d %s\n", m.ID, m.Provider, m.Privacy, m.ContextTokens, m.LatencyClass)
		}
		return exitOK
	case "route":
		req := modelregistry.RouteRequest{
			MinimumContext: 1000,
		}
		for i := 1; i < len(args); i++ {
			if args[i] == "--data-class" && i+1 < len(args) {
				req.DataClass = args[i+1]
				i++
			}
			if args[i] == "--tokens" && i+1 < len(args) {
				if n, err := strconv.Atoi(args[i+1]); err == nil {
					req.MinimumContext = n
				}
				i++
			}
			if args[i] == "--tools" {
				req.NeedTools = true
			}
			if args[i] == "--latency" && i+1 < len(args) {
				req.LatencyPreference = args[i+1]
				i++
			}
		}
		resp, err := reg.Route(req)
		if err != nil {
			if asJSON {
				return envelopeError(protocol.CodeValidationFailed, err.Error())
			}
			fmt.Printf("error: %v\n", err)
			return exitValidation
		}
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(resp))
		}
		fmt.Printf("Selected Route: %s (provider: %s)\n", resp.Primary.ID, resp.Primary.Provider)
		fmt.Printf("Explanation: %s\n", resp.Explanation)
		if len(resp.Fallbacks) > 0 {
			fmt.Printf("Fallbacks (%d):\n", len(resp.Fallbacks))
			for _, fb := range resp.Fallbacks {
				fmt.Printf("  - %s (provider: %s)\n", fb.ID, fb.Provider)
			}
		}
		return exitOK
	default:
		return exitUsage
	}
}

func runEnv(asJSON bool, root string, args []string) int {
	if len(args) == 0 {
		return exitUsage
	}
	switch args[0] {
	case "list":
		local := environment.NewLocalEnvironment(root, true)
		envs := []environment.Descriptor{
			local.Descriptor(),
			{
				ID:              "env-worktree-template",
				Version:         1,
				Kind:            "worktree",
				Isolation:       environment.IsolationWorktree,
				Network:         "deny",
				FilesystemRoots: []string{root},
			},
		}
		if asJSON {
			return printEnvelope(protocol.OkEnvelope(envs))
		}
		fmt.Printf("%-24s %-10s %-15s %s\n", "ID", "KIND", "ISOLATION", "NETWORK")
		fmt.Println("-----------------------------------------------------------------")
		for _, e := range envs {
			fmt.Printf("%-24s %-10s %-15s %s\n", e.ID, e.Kind, e.Isolation, e.Network)
		}
		return exitOK
	default:
		return exitUsage
	}
}
