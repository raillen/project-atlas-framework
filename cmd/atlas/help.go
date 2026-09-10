package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/raillen/project-atlas-framework/internal/protocol"
)

// CommandInfo holds documentation metadata for a CLI command.
type CommandInfo struct {
	Name        string
	Category    string
	Summary     string
	Usage       string
	Description string
	Flags       []string
	Subcommands []string
	Examples    []string
}

var commandRegistry = map[string]CommandInfo{
	"init": {
		Name:     "init",
		Category: "Project Lifecycle",
		Summary:  "Initialize a new Project Atlas project workspace",
		Usage:    "atlas init [path] --profile <profile.json> [--non-interactive]",
		Description: "Scaffolds a new Project Atlas project at the given path (default: current directory).\n" +
			"Resolves the workforce from the specified project profile and generates atlas.json,\n" +
			".ai/ directory, docs/ATLAS.md, PROJECT_STATE.md, and initial derived state.",
		Flags: []string{
			"--profile <path>       Path to project-profile.json declaring preferred models and stack (required)",
			"--non-interactive      Execute without interactive prompts (fails if profile missing)",
			"--home <path>          Custom Atlas home directory",
			"--json                 Output structured JSON response",
		},
		Examples: []string{
			"atlas init ./my-project --profile examples/brasa/project-profile.json",
			"atlas init . --profile ./profile.json --non-interactive",
		},
	},
	"status": {
		Name:     "status",
		Category: "Project Lifecycle",
		Summary:  "Display current Project Atlas workspace root path",
		Usage:    "atlas status [--path <dir>]",
		Description: "Finds and prints the canonical root directory of the Project Atlas project.\n" +
			"Returns exit code 0 if found, exit code 6 (ATLAS_PROJECT_NOT_FOUND) if outside a project.",
		Flags: []string{
			"--path <dir>           Target directory to inspect (default: current directory)",
			"--json                 Output structured JSON envelope with root path",
		},
		Examples: []string{
			"atlas status",
			"atlas --json status --path ./subfolder",
		},
	},
	"validate": {
		Name:     "validate",
		Category: "Project Lifecycle",
		Summary:  "Validate project structure and schema conformance",
		Usage:    "atlas validate [path]",
		Description: "Performs strict schema validation against canonical JSON schemas.\n" +
			"Verifies atlas.json, goals, policies, and directory structure.",
		Flags: []string{
			"--json                 Output structured validation results",
		},
		Examples: []string{
			"atlas validate",
			"atlas validate ./my-project",
		},
	},
	"doctor": {
		Name:     "doctor",
		Category: "Project Lifecycle",
		Summary:  "Run comprehensive diagnostic health checks",
		Usage:    "atlas doctor [path]",
		Description: "Performs deep health checks across all framework subsystems:\n" +
			"- Versioning and lockfile consistency\n" +
			"- Dependencies and DAG cycles\n" +
			"- Workforce resolution and catalog integrity\n" +
			"- Repository governance policies and gates\n" +
			"- Evidence models and test coverage",
		Flags: []string{
			"--json                 Output diagnostics as structured JSON envelope",
		},
		Examples: []string{
			"atlas doctor",
			"atlas --json doctor ./my-project",
		},
	},
	"framework-check": {
		Name:     "framework-check",
		Category: "Project Lifecycle",
		Summary:  "Validate framework internal integrity and catalog schemas",
		Usage:    "atlas framework-check",
		Description: "Validates the framework's own catalog, workforce skills, recipes,\n" +
			"JSON schemas, and embedded resources.",
		Flags: []string{
			"--json                 Output structured JSON response",
		},
		Examples: []string{
			"atlas framework-check",
			"atlas --json framework-check",
		},
	},
	"version": {
		Name:        "version",
		Category:    "General",
		Summary:     "Display Atlas CLI and protocol version",
		Usage:       "atlas version",
		Description: "Prints the semantic version of the Atlas CLI and the protocol contract version.",
		Flags: []string{
			"--json                 Output version information as JSON envelope",
		},
		Examples: []string{
			"atlas version",
			"atlas --json version",
		},
	},
	"goal": {
		Name:     "goal",
		Category: "Goals & Execution",
		Summary:  "Manage goal lifecycle, states, and amendments",
		Usage:    "atlas goal <subcommand> [args] [--path <dir>]",
		Description: "Manages verifiable Goals through the protocol state machine:\n" +
			"DRAFT -> PLANNED -> LOCKED -> EXECUTING -> VERIFYING -> REVIEWING -> DONE.",
		Subcommands: []string{
			"new <id> <title> --phase <phase> --objective <text>   Create a new goal",
			"list                                                  List all project goals and states",
			"state <id> <state>                                    Transition goal state (requires evidence for DONE)",
			"amend <id> --file <amendment.json>                    Apply formal amendment to locked goal",
		},
		Flags: []string{
			"--path <dir>           Project directory (default: .)",
			"--phase <phase>        Phase identifier (e.g., P00, P01)",
			"--objective <text>     Objective description",
			"--file <path>          Path to amendment file",
			"--json                 Output structured JSON envelope",
		},
		Examples: []string{
			"atlas goal new P00-G01 \"Foundation\" --phase P00 --objective \"Set up base CI and tests\" --path ./my-project",
			"atlas goal list --path ./my-project",
			"atlas goal state P00-G01 LOCKED --path ./my-project",
		},
	},
	"plan": {
		Name:     "plan",
		Category: "Goals & Execution",
		Summary:  "Inspect and update the Living Plan",
		Usage:    "atlas plan <subcommand> [args] [--path <dir>]",
		Description: "Maintains the continuous Living Plan across development goals,\n" +
			"supporting delta feedback, question resolution, and decision previews.",
		Subcommands: []string{
			"init                                                  Initialize living plan in project",
			"show                                                  Display current living plan summary",
			"update --delta <delta.json>                           Apply delta feedback to plan",
		},
		Flags: []string{
			"--path <dir>           Project directory (default: .)",
			"--delta <path>         Path to delta feedback JSON",
			"--json                 Output structured JSON envelope",
		},
		Examples: []string{
			"atlas plan show --path ./my-project",
			"atlas plan update --delta feedback.json --path ./my-project",
		},
	},
	"context": {
		Name:     "context",
		Category: "Goals & Execution",
		Summary:  "Plan context strategy and budget allocation (LPC)",
		Usage:    "atlas context plan <description> [--path <dir>]",
		Description: "Applies Lean Progressive Context (LPC) to determine the minimal sufficient\n" +
			"context capsule, token budget, and retrieval strategy for an AI coding goal.",
		Flags: []string{
			"--path <dir>           Project directory (default: .)",
			"--json                 Output context plan envelope",
		},
		Examples: []string{
			"atlas context plan \"debug database connection leak\" --path ./my-project --json",
		},
	},
	"resolve": {
		Name:     "resolve",
		Category: "Workforce & Compilation",
		Summary:  "Resolve workforce skills, agents, and recipes for a profile",
		Usage:    "atlas resolve <profile.json>",
		Description: "Deterministically matches stack tokens, features, and risk profiles from\n" +
			"project-profile.json against the canonical workforce registry (151 skills, agents, recipes).",
		Flags: []string{
			"--json                 Output resolved workforce as JSON envelope",
		},
		Examples: []string{
			"atlas resolve examples/brasa/project-profile.json",
			"atlas --json resolve ./profile.json",
		},
	},
	"explain": {
		Name:     "explain",
		Category: "Workforce & Compilation",
		Summary:  "Inspect detailed contracts of workforce entities",
		Usage:    "atlas explain <type> <id>",
		Description: "Displays inputs, outputs, capabilities, required evidence, and instructions for:\n" +
			"- workforce <profile.json>: explain resolution rationale\n" +
			"- agent <agent-id>: explain agent persona and rules\n" +
			"- skill <skill-id>: explain implementation contract and checks\n" +
			"- recipe <recipe-id>: explain multi-step coordination flow",
		Flags: []string{
			"--json                 Output explanation as JSON envelope",
		},
		Examples: []string{
			"atlas explain skill lang-cpp",
			"atlas explain skill lang-rust --json",
			"atlas explain agent architect",
			"atlas explain recipe web-feature",
		},
	},
	"compile": {
		Name:     "compile",
		Category: "Workforce & Compilation",
		Summary:  "Compile and emit target harness adapters and skills",
		Usage:    "atlas compile --target <target> [--path <dir>]",
		Description: "Compiles canonical project state, workforce skills, instructions, and\n" +
			"rules into target-specific AI coding harness configurations.",
		Flags: []string{
			"--target <name>        Harness target: antigravity, opencode, codex, claudecode, claude, gemini, generic, traycer, chatgpt, kimi",
			"--path <dir>           Project directory (default: .)",
			"--json                 Output compilation summary as JSON",
		},
		Examples: []string{
			"atlas compile --target antigravity --path ./my-project",
			"atlas compile --target opencode --path ./my-project",
			"atlas compile --target codex --path ./my-project",
		},
	},
	"tool": {
		Name:     "tool",
		Category: "Tooling & Verification",
		Summary:  "Execute embedded developer and safety verification tools",
		Usage:    "atlas tool <tool-name> [args]",
		Description: "Executes pure Go native tools integrated with the Atlas Tool Gateway:\n" +
			"- check-escape-hatches: Scan source code for unregistered escape hatches\n" +
			"- verify-language-contract: Validate language safety profiles\n" +
			"- scan-sanitizers: Audit compiler sanitizer matrix\n" +
			"- scan-secrets: Scan for exposed API keys and credentials\n" +
			"- check-permissions: Audit file and directory permissions\n" +
			"- analyze-complexity: Static cognitive and cyclomatic complexity analysis\n" +
			"- check-bare-errors: Detect swallowed/unhandled error returns\n" +
			"- check-subprocesses: Audit external process spawning for command injection",
		Flags: []string{
			"--json                 Output tool results as JSON envelope",
		},
		Examples: []string{
			"atlas tool check-escape-hatches .",
			"atlas tool check-escape-hatches ./src/my-service",
			"atlas tool scan-secrets .",
			"atlas tool analyze-complexity ./internal",
		},
	},
	"adopt": {
		Name:     "adopt",
		Category: "Adoption & Migration",
		Summary:  "Brownfield project scanner and adoption engine",
		Usage:    "atlas adopt <subcommand> [path]",
		Description: "Scans non-Atlas codebases, discovers technical facts, classifies tech stacks,\n" +
			"and generates a non-destructive migration ledger and candidate Atlas configuration.",
		Subcommands: []string{
			"scan [path]        Scan directory tree and index project artifacts",
			"facts [path]       Extract observed facts (languages, frameworks, build systems)",
			"classify [path]    Classify project architecture and capability profile",
			"scaffold [path]    Generate candidate atlas.json without overwriting user files",
		},
		Flags: []string{
			"--json             Output scan results as JSON envelope",
		},
		Examples: []string{
			"atlas adopt scan ./legacy-app",
			"atlas adopt facts ./legacy-app --json",
			"atlas adopt classify ./legacy-app",
		},
	},
	"report": {
		Name:     "report",
		Category: "Governance & Traceability",
		Summary:  "Ingest execution reports and query project intelligence",
		Usage:    "atlas report <subcommand> [args] [--path <dir>]",
		Description: "Collects verifiable task execution evidence and updates derived\n" +
			"project intelligence metrics (.atlas/history/project-intelligence.json).",
		Subcommands: []string{
			"add <report.json>  Ingest a task execution evidence report",
			"summary            Print summary of project metrics, goals, and test history",
		},
		Flags: []string{
			"--path <dir>       Project directory (default: .)",
			"--json             Output summary as JSON envelope",
		},
		Examples: []string{
			"atlas report add ./task-report.json --path ./my-project",
			"atlas report summary --path ./my-project",
		},
	},
	"snapshot": {
		Name:     "snapshot",
		Category: "Project Lifecycle",
		Summary:  "Create an archive backup snapshot of project state",
		Usage:    "atlas snapshot [path] --output <archive.zip>",
		Description: "Generates a complete, reproducible zip archive of project canonical state,\n" +
			"governance policies, and goals for backup or safe migration rollback.",
		Flags: []string{
			"--output <path>    Target zip archive path (required)",
			"--json             Output snapshot metadata as JSON",
		},
		Examples: []string{
			"atlas snapshot ./my-project --output ./backup-v1.zip",
		},
	},
	"migrate": {
		Name:     "migrate",
		Category: "Project Lifecycle",
		Summary:  "Migrate project metadata and schemas across versions",
		Usage:    "atlas migrate [path] [--dry-run]",
		Description: "Upgrades project schemas and configurations to the current Atlas version.\n" +
			"Always run with --dry-run first to preview changes.",
		Flags: []string{
			"--dry-run          Preview migration changes without modifying files",
			"--json             Output migration plan as JSON",
		},
		Examples: []string{
			"atlas migrate ./my-project --dry-run",
			"atlas migrate ./my-project",
		},
	},
	"repo": {
		Name:     "repo",
		Category: "Governance & Traceability",
		Summary:  "Repository governance policy and risk gate enforcement",
		Usage:    "atlas repo <subcommand> [args]",
		Description: "Enforces repository policy (.atlas/repository/policy.json) for branches,\n" +
			"commit message conventions, PR risk evaluation, and emergency bypass audits.",
		Subcommands: []string{
			"policy             Display effective repository governance policy",
			"branch-check <name> Validate branch naming pattern",
			"commit-check <msg>  Validate conventional commit message format",
			"pr-check [args]    Evaluate PR risk level and required verification gates",
		},
		Flags: []string{
			"--json             Output policy evaluation as JSON",
		},
		Examples: []string{
			"atlas repo policy",
			"atlas repo branch-check feat/my-feature",
			"atlas repo commit-check \"feat(core): implement feature\"",
		},
	},
	"docs": {
		Name:     "docs",
		Category: "Governance & Traceability",
		Summary:  "Documentation architecture and semantic delta checks",
		Usage:    "atlas docs <subcommand> [args]",
		Description: "Validates local canonical engineering documentation against schema\n" +
			"contracts and tracks semantic documentation deltas.",
		Subcommands: []string{
			"check [path]       Validate documentation files against schemas",
			"export [path]      Export documentation index and manifest",
			"validate [path]    Validate canonical local documentation integrity",
		},
		Flags: []string{
			"--json             Output documentation check results as JSON",
		},
		Examples: []string{
			"atlas docs check",
			"atlas docs validate",
		},
	},
	"trace": {
		Name:     "trace",
		Category: "Governance & Traceability",
		Summary:  "Query traceability graph, events, and evidence",
		Usage:    "atlas trace <subcommand> [args]",
		Description: "Inspects provenance, causality DAGs, execution events, and\n" +
			"cryptographic evidence links across the lifecycle.",
		Subcommands: []string{
			"graph              Display traceability graph connections",
			"events             List recorded execution events",
			"evidence           Inspect cryptographic evidence entries",
		},
		Flags: []string{
			"--json             Output traceability data as JSON",
		},
		Examples: []string{
			"atlas trace graph",
			"atlas trace events --json",
		},
	},
	"journal": {
		Name:     "journal",
		Category: "Governance & Traceability",
		Summary:  "Inspect operational event journal",
		Usage:    "atlas journal [subcommand] [args]",
		Description: "Provides append-only operational audit trail inspection for agent\n" +
			"actions, tool executions, and state modifications.",
		Flags: []string{
			"--json             Output journal entries as JSON",
		},
		Examples: []string{
			"atlas journal",
		},
	},
	"experience": {
		Name:     "experience",
		Category: "Governance & Traceability",
		Summary:  "Developer experience heuristics and next-action guidance",
		Usage:    "atlas experience [subcommand] [args]",
		Description: "Evaluates developer workflow ergonomics, active bottlenecks, and\n" +
			"provides proactive next-action recommendations.",
		Flags: []string{
			"--json             Output experience metrics as JSON",
		},
		Examples: []string{
			"atlas experience",
		},
	},
	"setup": {
		Name:        "setup",
		Category:    "Environment & Connectors",
		Summary:     "Initialize global Atlas directories and machine state",
		Usage:       "atlas setup",
		Description: "Sets up ~/.atlas directories, catalogs, and local machine configuration.",
		Flags: []string{
			"--home <path>      Custom Atlas home directory (default: ~/.atlas)",
			"--json             Output setup status as JSON",
		},
		Examples: []string{
			"atlas setup",
		},
	},
	"install": {
		Name:     "install",
		Category: "Environment & Connectors",
		Summary:  "Install IDE and agent connectors or CLI binaries",
		Usage:    "atlas install <target> [--dir <path>]",
		Description: "Installs target harness connectors (e.g., opencode, antigravity, codex)\n" +
			"into host environments.",
		Flags: []string{
			"--dir <path>       Target installation directory",
			"--home <path>      Custom Atlas home directory",
			"--json             Output installation report as JSON",
		},
		Examples: []string{
			"atlas install opencode",
			"atlas install antigravity",
		},
	},
	"uninstall": {
		Name:        "uninstall",
		Category:    "Environment & Connectors",
		Summary:     "Uninstall specified connector",
		Usage:       "atlas uninstall <target>",
		Description: "Removes an installed harness connector from the host environment.",
		Flags: []string{
			"--home <path>      Custom Atlas home directory",
			"--json             Output uninstallation status as JSON",
		},
		Examples: []string{
			"atlas uninstall opencode",
		},
	},
	"connector": {
		Name:        "connector",
		Category:    "Environment & Connectors",
		Summary:     "Manage and inspect harness connectors",
		Usage:       "atlas connector <subcommand> [args]",
		Description: "Inspects status, capabilities, and health of installed AI harness connectors.",
		Subcommands: []string{
			"list               List all supported and installed connectors",
			"status <name>      Check status of a specific connector",
		},
		Flags: []string{
			"--json             Output connector data as JSON",
		},
		Examples: []string{
			"atlas connector list",
			"atlas connector status opencode",
		},
	},
	"run": {
		Name:     "run",
		Category: "Control Plane Runtime",
		Summary:  "Execute the autonomous goal implementation loop",
		Usage:    "atlas run [args]",
		Description: "Starts the deterministic execution loop for the active project goal,\n" +
			"enforcing Red-Green-Refactor cycles and quality gates.",
		Flags: []string{
			"--json             Output execution events as JSON",
		},
		Examples: []string{
			"atlas run",
		},
	},
	"continue": {
		Name:        "continue",
		Category:    "Control Plane Runtime",
		Summary:     "Resume execution from a checkpointed state",
		Usage:       "atlas continue [args]",
		Description: "Resumes execution from the most recent safe checkpoint in .atlas/runtime/.",
		Flags: []string{
			"--json             Output resume status as JSON",
		},
		Examples: []string{
			"atlas continue",
		},
	},
	"budget": {
		Name:        "budget",
		Category:    "Control Plane Runtime",
		Summary:     "Inspect token, cost, and execution budgets",
		Usage:       "atlas budget [args]",
		Description: "Displays consumed tokens, model call costs, and remaining budget limits.",
		Flags: []string{
			"--json             Output budget metrics as JSON",
		},
		Examples: []string{
			"atlas budget",
		},
	},
	"debug": {
		Name:        "debug",
		Category:    "Control Plane Runtime",
		Summary:     "Debug runtime state and Working Context Capsules",
		Usage:       "atlas debug [args]",
		Description: "Inspects memory capsules, active goal context, and runtime journal entries.",
		Flags: []string{
			"--json             Output debug state as JSON",
		},
		Examples: []string{
			"atlas debug",
		},
	},
	"model": {
		Name:     "model",
		Category: "Control Plane Runtime",
		Summary:  "Model registry, capabilities, and provider routing",
		Usage:    "atlas model [subcommand] [args]",
		Description: "Inspects configured AI models, capabilities (context window, tool calling),\n" +
			"and cost parameters across providers.",
		Flags: []string{
			"--json             Output model registry as JSON",
		},
		Examples: []string{
			"atlas model list",
		},
	},
	"env": {
		Name:        "env",
		Category:    "Environment & Connectors",
		Summary:     "Environment diagnostics and PATH configuration",
		Usage:       "atlas env [args]",
		Description: "Displays active environment variables, toolchain availability, and PATH state.",
		Flags: []string{
			"--json             Output environment status as JSON",
		},
		Examples: []string{
			"atlas env",
		},
	},
	"runtime": {
		Name:        "runtime",
		Category:    "Control Plane Runtime",
		Summary:     "Control plane runtime management",
		Usage:       "atlas runtime [subcommand] [args]",
		Description: "Inspects or resets runtime control plane state.",
		Flags: []string{
			"--json             Output runtime status as JSON",
		},
		Examples: []string{
			"atlas runtime status",
		},
	},
	"package": {
		Name:        "package",
		Category:    "Workforce & Compilation",
		Summary:     "Package management and distribution",
		Usage:       "atlas package [subcommand] [args]",
		Description: "Inspects or bundles skill packages and distribution archives.",
		Flags: []string{
			"--json             Output package info as JSON",
		},
		Examples: []string{
			"atlas package list",
		},
	},
	"automation": {
		Name:        "automation",
		Category:    "Platform & Automation",
		Summary:     "Automation workflow engine",
		Usage:       "atlas automation [subcommand] [args]",
		Description: "Manages scheduled and event-driven automation recipes.",
		Flags: []string{
			"--json             Output automation status as JSON",
		},
		Examples: []string{
			"atlas automation list",
		},
	},
}

// PrintGeneralHelp prints top-level CLI help to stdout.
func PrintGeneralHelp(asJSON bool) int {
	if asJSON {
		categories := map[string][]map[string]string{}
		for _, info := range commandRegistry {
			categories[info.Category] = append(categories[info.Category], map[string]string{
				"name":    info.Name,
				"summary": info.Summary,
				"usage":   info.Usage,
			})
		}
		return printEnvelope(protocol.OkEnvelope(map[string]any{
			"version":    protocol.CLIVersion,
			"categories": categories,
		}))
	}

	fmt.Printf("Project Atlas Framework CLI v%s\n\n", protocol.CLIVersion)
	fmt.Println("Usage:")
	fmt.Println("  atlas [--json] [--home <path>] <command> [subcommand] [flags]")
	fmt.Println()

	// Group by category
	categoryOrder := []string{
		"Project Lifecycle",
		"Goals & Execution",
		"Workforce & Compilation",
		"Tooling & Verification",
		"Governance & Traceability",
		"Control Plane Runtime",
		"Environment & Connectors",
		"General",
	}

	catMap := map[string][]CommandInfo{}
	for _, info := range commandRegistry {
		catMap[info.Category] = append(catMap[info.Category], info)
	}

	for _, cat := range categoryOrder {
		cmds, ok := catMap[cat]
		if !ok || len(cmds) == 0 {
			continue
		}
		sort.Slice(cmds, func(i, j int) bool { return cmds[i].Name < cmds[j].Name })

		fmt.Printf("%s:\n", cat)
		for _, cmd := range cmds {
			fmt.Printf("  %-18s %s\n", cmd.Name, cmd.Summary)
		}
		fmt.Println()
	}

	fmt.Println("Global Flags:")
	fmt.Println("  --json              Format command output as structured JSON envelope")
	fmt.Println("  --home <path>       Specify custom Atlas home directory (default: ~/.atlas)")
	fmt.Println("  --help, -h          Display help information for Atlas or any command")
	fmt.Println()
	fmt.Println("Quick Examples:")
	fmt.Println("  atlas init ./my-project --profile examples/brasa/project-profile.json")
	fmt.Println("  atlas tool check-escape-hatches .")
	fmt.Println("  atlas compile --target antigravity --path ./my-project")
	fmt.Println("  atlas doctor ./my-project")
	fmt.Println()
	fmt.Println("Run 'atlas <command> --help' or 'atlas help <command>' for detailed help on any command.")
	return exitOK
}

// PrintCommandHelp prints detailed help for a specific command to stdout.
func PrintCommandHelp(command string, asJSON bool) int {
	cmd := strings.ToLower(command)
	info, found := commandRegistry[cmd]
	if !found {
		fmt.Fprintf(os.Stderr, "error: unknown command '%s'\n\n", command)
		fmt.Fprintf(os.Stderr, "Run 'atlas --help' to view all available commands.\n")
		return exitUsage
	}

	if asJSON {
		return printEnvelope(protocol.OkEnvelope(map[string]any{
			"command":     info.Name,
			"category":    info.Category,
			"summary":     info.Summary,
			"usage":       info.Usage,
			"description": info.Description,
			"flags":       info.Flags,
			"subcommands": info.Subcommands,
			"examples":    info.Examples,
		}))
	}

	fmt.Printf("COMMAND: atlas %s\n\n", info.Name)
	fmt.Printf("Summary:\n  %s\n\n", info.Summary)
	fmt.Printf("Usage:\n  %s\n\n", info.Usage)

	if info.Description != "" {
		fmt.Printf("Description:\n")
		lines := strings.Split(info.Description, "\n")
		for _, line := range lines {
			fmt.Printf("  %s\n", line)
		}
		fmt.Println()
	}

	if len(info.Subcommands) > 0 {
		fmt.Println("Subcommands:")
		for _, sub := range info.Subcommands {
			fmt.Printf("  %s\n", sub)
		}
		fmt.Println()
	}

	if len(info.Flags) > 0 {
		fmt.Println("Flags & Options:")
		for _, flg := range info.Flags {
			fmt.Printf("  %s\n", flg)
		}
		fmt.Println()
	}

	if len(info.Examples) > 0 {
		fmt.Println("Examples:")
		for _, ex := range info.Examples {
			fmt.Printf("  $ %s\n", ex)
		}
		fmt.Println()
	}

	return exitOK
}
