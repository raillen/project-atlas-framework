package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/raillen/project-atlas-framework/internal/connectors"
	_ "github.com/raillen/project-atlas-framework/internal/connectors/opencode"
	"github.com/raillen/project-atlas-framework/internal/protocol"
)

func runConnector(asJSON bool, args []string) int {
	if len(args) == 0 {
		return runConnectorList(asJSON)
	}

	sub := args[0]
	switch sub {
	case "list":
		return runConnectorList(asJSON)
	case "install":
		if len(args) < 2 {
			err := fmt.Errorf("usage: atlas connector install <name>")
			return serviceError(asJSON, err)
		}
		return runConnectorInstall(asJSON, args[1], args[2:])
	case "uninstall":
		if len(args) < 2 {
			err := fmt.Errorf("usage: atlas connector uninstall <name>")
			return serviceError(asJSON, err)
		}
		return runConnectorUninstall(asJSON, args[1], args[2:])
	case "validate":
		if len(args) < 2 {
			err := fmt.Errorf("usage: atlas connector validate <name>")
			return serviceError(asJSON, err)
		}
		return runConnectorValidate(asJSON, args[1], args[2:])
	default:
		err := fmt.Errorf("unknown connector subcommand: %s", sub)
		return serviceError(asJSON, err)
	}
}

func runConnectorList(asJSON bool) int {
	list := connectors.List()
	records := make([]map[string]any, 0, len(list))
	for _, c := range list {
		contract := c.Contract()
		records = append(records, map[string]any{
			"id":           c.ID(),
			"name":         c.Name(),
			"version":      contract.Version,
			"enforcement":  contract.Enforcement,
			"capabilities": contract.Capabilities,
			"hooks":        contract.Hooks,
		})
	}

	if asJSON {
		return printEnvelope(protocol.OkEnvelope(map[string]any{
			"connectors": records,
		}))
	}

	fmt.Println("Registered Project Atlas Connectors:")
	for _, r := range records {
		fmt.Printf("  • %s (%s) - v%s [%s]\n", r["name"], r["id"], r["version"], r["enforcement"])
		if caps, ok := r["capabilities"].([]string); ok && len(caps) > 0 {
			fmt.Printf("    Capabilities: %s\n", strings.Join(caps, ", "))
		}
		if hooks, ok := r["hooks"].([]string); ok && len(hooks) > 0 {
			fmt.Printf("    Hooks: %s\n", strings.Join(hooks, ", "))
		}
	}
	return exitOK
}

func runConnectorInstall(asJSON bool, id string, args []string) int {
	c, err := connectors.Get(id)
	if err != nil {
		return serviceError(asJSON, err)
	}

	home, err := installationHome("")
	if err != nil {
		return serviceError(asJSON, err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return serviceError(asJSON, err)
	}

	res, err := c.Install(home, cwd, connectors.InstallOptions{})
	if err != nil {
		return serviceError(asJSON, err)
	}

	if asJSON {
		return printEnvelope(protocol.OkEnvelope(res))
	}

	fmt.Printf("Connector %s (%s) installed successfully.\n", c.Name(), c.ID())
	fmt.Printf("Created %d artifacts in project.\n", len(res.CreatedPaths))
	fmt.Printf("Cleanup manifest: %s\n", res.CleanupPath)
	return exitOK
}

func runConnectorUninstall(asJSON bool, id string, args []string) int {
	c, err := connectors.Get(id)
	if err != nil {
		return serviceError(asJSON, err)
	}

	home, err := installationHome("")
	if err != nil {
		return serviceError(asJSON, err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return serviceError(asJSON, err)
	}

	res, err := c.Uninstall(home, cwd, connectors.UninstallOptions{})
	if err != nil {
		return serviceError(asJSON, err)
	}

	if asJSON {
		return printEnvelope(protocol.OkEnvelope(res))
	}

	fmt.Printf("Connector %s uninstalled.\n", id)
	fmt.Printf("Removed %d artifacts.\n", len(res.Removed))
	if len(res.Leftovers) > 0 {
		fmt.Printf("Leftovers retained: %v\n", res.Leftovers)
	}
	return exitOK
}

func runConnectorValidate(asJSON bool, id string, args []string) int {
	c, err := connectors.Get(id)
	if err != nil {
		return serviceError(asJSON, err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return serviceError(asJSON, err)
	}

	res, err := c.Validate(cwd)
	if err != nil {
		return serviceError(asJSON, err)
	}

	if asJSON {
		if !res.Valid {
			return envelopeError("CONNECTOR_INVALID", fmt.Sprintf("connector validation failed: %v", res.Errors))
		}
		return printEnvelope(protocol.OkEnvelope(res))
	}

	if !res.Valid {
		fmt.Printf("Validation failed for connector %s:\n", id)
		for _, err := range res.Errors {
			fmt.Printf("  [ERROR] %s\n", err)
		}
		return exitValidation
	}

	fmt.Printf("Connector %s validated successfully (%d files verified).\n", id, len(res.Files))
	return exitOK
}
