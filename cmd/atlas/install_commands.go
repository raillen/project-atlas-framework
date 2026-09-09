package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/raillen/project-atlas-framework/internal/install"
	"github.com/raillen/project-atlas-framework/internal/protocol"
)

func installationHome(explicit string) (string, error) {
	return install.HomeDir(explicit)
}

func runSetup(asJSON bool, explicitHome string) int {
	home, err := installationHome(explicitHome)
	if err != nil {
		return serviceError(asJSON, err)
	}
	manifest, err := install.LoadManifest(home)
	if err != nil {
		return serviceError(asJSON, err)
	}
	if manifest.AtlasVersion == "" {
		manifest.AtlasVersion = protocol.CLIVersion
	}
	if manifest.BinaryPath == "" {
		if executable, err := os.Executable(); err == nil {
			manifest.BinaryPath = executable
		}
	}
	if err := install.SaveManifest(home, manifest); err != nil {
		return serviceError(asJSON, err)
	}
	harnesses := map[string]bool{}
	for _, candidate := range []string{"opencode", "codex", "code", "gemini", "copilot", "kiro"} {
		harnesses[candidate] = onPath(candidate)
	}
	result := map[string]any{"home": home, "manifest": manifest, "harnesses": harnesses, "path_valid": pathContains(home), "idempotent": true}
	if asJSON {
		return printEnvelope(protocol.OkEnvelope(result))
	}
	fmt.Printf("Atlas setup complete: %s\n", home)
	fmt.Printf("Detected harnesses:")
	detected := false
	for name, present := range harnesses {
		if present {
			fmt.Printf(" %s", name)
			detected = true
		}
	}
	if !detected {
		fmt.Printf(" none")
	}
	fmt.Printf("\n")
	return exitOK
}

func onPath(name string) bool {
	for _, dir := range pathDirs() {
		if _, err := os.Stat(dir + string(os.PathSeparator) + name); err == nil {
			return true
		}
	}
	return false
}

func pathDirs() []string {
	separator := ":"
	if os.PathSeparator == '\\' {
		separator = ";"
	}
	out := []string{}
	for _, dir := range splitEnv(os.Getenv("PATH"), separator) {
		if dir != "" {
			out = append(out, dir)
		}
	}
	return out
}

func splitEnv(value, separator string) []string {
	if value == "" {
		return []string{}
	}
	return strings.Split(value, separator)
}

func pathContains(home string) bool {
	bin := home + string(os.PathSeparator) + "bin"
	for _, dir := range pathDirs() {
		if dir == bin {
			return true
		}
	}
	return false
}

func runInstall(asJSON bool, explicitHome string, args []string) int {
	home, err := installationHome(explicitHome)
	if err != nil {
		return serviceError(asJSON, err)
	}
	manifest, err := install.LoadManifest(home)
	if err != nil {
		return serviceError(asJSON, err)
	}
	if manifest.AtlasVersion == "" {
		manifest.AtlasVersion = protocol.CLIVersion
	}
	if executable, err := os.Executable(); err == nil && manifest.BinaryPath == "" {
		manifest.BinaryPath = executable
	}
	if len(args) >= 2 && args[0] == "connector" {
		return runConnectorInstall(asJSON, args[1], args[2:])
	}
	if err := install.SaveManifest(home, manifest); err != nil {
		return serviceError(asJSON, err)
	}
	result := map[string]any{"home": home, "manifest": manifest}
	if asJSON {
		return printEnvelope(protocol.OkEnvelope(result))
	}
	fmt.Printf("Atlas installation state updated: %s\n", home)
	return exitOK
}

func runUninstall(asJSON bool, explicitHome string, args []string) int {
	home, err := installationHome(explicitHome)
	if err != nil {
		return serviceError(asJSON, err)
	}
	manifest, err := install.LoadManifest(home)
	if err != nil {
		return serviceError(asJSON, err)
	}
	removed := []string{}
	leftovers := []string{}
	if hasFlag(args, "--connectors") {
		for connector := range manifest.Connectors {
			cleanupPath := install.CleanupPath(home, connector)
			cleanupData, readErr := os.ReadFile(cleanupPath)
			if readErr != nil {
				continue
			}
			var cleanup install.CleanupManifest
			if json.Unmarshal(cleanupData, &cleanup) == nil {
				gone, remaining := install.RemoveManagedPaths(cleanup.CreatedPaths)
				removed = append(removed, gone...)
				leftovers = append(leftovers, remaining...)
			}
		}
		manifest.Connectors = map[string]string{}
	}
	if hasFlag(args, "--purge-cache") {
		if err := install.PurgeDirectory(home + "/cache"); err != nil {
			return serviceError(asJSON, err)
		}
		removed = append(removed, home+"/cache")
	}
	if hasFlag(args, "--purge-global-config") {
		if err := install.PurgeDirectory(home + "/config"); err != nil {
			return serviceError(asJSON, err)
		}
		removed = append(removed, home+"/config")
	}
	if !hasFlag(args, "--purge-global-config") {
		if err := install.SaveManifest(home, manifest); err != nil {
			return serviceError(asJSON, err)
		}
	}
	result := map[string]any{"home": home, "removed": removed, "leftovers": leftovers, "project_data_preserved": true}
	if asJSON {
		return printEnvelope(protocol.OkEnvelope(result))
	}
	fmt.Printf("Atlas uninstall completed; project data preserved.\n")
	return exitOK
}
