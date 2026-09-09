package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/raillen/project-atlas-framework/internal/adoption"
	"github.com/raillen/project-atlas-framework/internal/protocol"
)

func runAdopt(asJSON bool, args []string) int {
	path := "."
	strict := false
	auditOnly := false
	interactive := false
	nonInteractive := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--path":
			if i+1 < len(args) {
				path = args[i+1]
				i++
			} else {
				fmt.Fprintf(os.Stderr, "error: --path requires a directory path\n")
				return exitUsage
			}
		case "--audit-only":
			auditOnly = true
		case "--strict":
			strict = true
		case "--interactive":
			interactive = true
		case "--non-interactive":
			nonInteractive = true
		default:
			if !strings.HasPrefix(args[i], "-") && path == "." {
				path = args[i]
			}
		}
	}

	_ = auditOnly
	_ = interactive
	_ = nonInteractive

	opts := adoption.ScanOptions{
		Budget:         adoption.DefaultBudget(),
		ParseManifests: true,
	}

	report, err := adoption.RunAdoptionAudit(path, opts)
	if err != nil {
		if asJSON {
			return envelopeError("adoption_error", err.Error())
		}
		fmt.Fprintf(os.Stderr, "error running adoption audit: %s\n", err)
		return exitInternal
	}

	if strict {
		if len(report.Contradictions) > 0 || report.Ledger.Summary.RequiresConfirmation > 0 {
			msg := fmt.Sprintf("strict adoption check failed: %d contradiction(s), %d unconfirmed inference(s)",
				len(report.Contradictions), report.Ledger.Summary.RequiresConfirmation)
			if asJSON {
				return printEnvelope(protocol.ErrEnvelope(protocol.Diagnostic{
					Code:    "adoption_strict_failure",
					Message: msg,
				}))
			}
			fmt.Fprintf(os.Stderr, "error: %s\n", msg)
			return exitValidation
		}
	}

	if asJSON {
		return printEnvelope(protocol.OkEnvelope(report))
	}

	fmt.Print(adoption.RenderHumanReport(report))
	return exitOK
}
