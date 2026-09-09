package main

import (
	"bufio"
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

	// F-G07: Handle uncertainty resolution
	if nonInteractive {
		session := adoption.ResolveNonInteractive(&report.Ledger)
		if !asJSON && session.ResolvedCount > 0 {
			fmt.Printf("Non-interactive pass: resolved %d inference(s) (%d unconfirmed remain)\n\n",
				session.ResolvedCount, session.UnresolvedCount)
		}
	} else if interactive {
		questions := adoption.GenerateQuestions(report.Ledger)
		if len(questions) > 0 {
			scanner := bufio.NewScanner(os.Stdin)
			for _, q := range questions {
				if !asJSON {
					fmt.Printf("\n[Adoption Interview] %s\n", q.Question)
					if q.Context != "" {
						fmt.Printf("  Context: %s\n", q.Context)
					}
					fmt.Printf("  Options: [%s] (default: %s)\n", strings.Join(q.Options, ", "), q.DefaultChoice)
					fmt.Print("  Choice > ")
				}

				var choiceStr string
				if scanner.Scan() {
					choiceStr = strings.TrimSpace(scanner.Text())
				}
				if choiceStr == "" {
					choiceStr = q.DefaultChoice
				}

				choice := adoption.ResolutionChoice{
					QuestionID:     q.ID,
					SelectedOption: choiceStr,
					Actor:          "human-operator",
					Rationale:      "Confirmed via interactive adoption interview",
				}
				_ = adoption.ApplyChoice(&report.Ledger, q, choice)
			}
		}
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
