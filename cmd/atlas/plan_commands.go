package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/raillen/project-atlas-framework/internal/app"
	"github.com/raillen/project-atlas-framework/internal/contextcompiler"
	"github.com/raillen/project-atlas-framework/internal/planning"
	"github.com/raillen/project-atlas-framework/internal/protocol"
)

// planSessionData is the harness-neutral payload of `atlas plan status`.
type planSessionData struct {
	Sessions []app.SessionStatus `json:"sessions"`
}

// planQuestionsData is the payload of `atlas plan questions`.
type planQuestionsData struct {
	Session   string                  `json:"session"`
	Questions []planning.OpenQuestion `json:"questions"`
}

// planResumeData is the payload of `atlas plan resume` (and `atlas plan
// --goal <id>` / `atlas plan --resume`). It is the interaction artifact a
// harness reads to drive the next turn: the reconstructed session, its status,
// the compiled context manifest with pressure, the open questions in impact
// order, and whether the run can still proceed.
type planResumeData struct {
	Session   planning.PlanningSession `json:"session"`
	Status    app.SessionStatus        `json:"status"`
	Budget    int                      `json:"budget"`
	Manifest  contextcompiler.Manifest `json:"manifest"`
	Questions []planning.OpenQuestion  `json:"questions"`
	Resumable bool                     `json:"resumable"`
}

// runPlan drives the `atlas plan` interaction protocol. The same application
// service serves both surfaces: `--json` emits the stable machine envelope,
// plain mode prints deterministic human text.
func runPlan(asJSON bool, args []string) int {
	path := "."
	sessionID := ""
	goal := ""
	budget := 0
	action := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "--") {
		action = args[0]
		args = args[1:]
	}
	for i := 0; i < len(args); i++ {
		value := func() (string, bool) {
			if i+1 >= len(args) {
				return "", false
			}
			v := args[i+1]
			i++
			return v, true
		}
		switch args[i] {
		case "--path":
			v, ok := value()
			if !ok {
				fmt.Fprintln(os.Stderr, "error: --path requires a value")
				return exitUsage
			}
			path = v
		case "--session":
			v, ok := value()
			if !ok {
				fmt.Fprintln(os.Stderr, "error: --session requires a value")
				return exitUsage
			}
			sessionID = v
		case "--goal":
			v, ok := value()
			if !ok {
				fmt.Fprintln(os.Stderr, "error: --goal requires a value")
				return exitUsage
			}
			goal = v
		case "--budget":
			v, ok := value()
			if !ok {
				fmt.Fprintln(os.Stderr, "error: --budget requires a value")
				return exitUsage
			}
			parsed, err := strconv.Atoi(v)
			if err != nil || parsed < 0 {
				fmt.Fprintln(os.Stderr, "error: --budget requires a non-negative integer")
				return exitUsage
			}
			budget = parsed
		case "--resume":
			if action == "" {
				action = "resume"
			}
		default:
			fmt.Fprintf(os.Stderr, "error: unsupported plan option: %s\n", args[i])
			return exitUsage
		}
	}
	if action == "" {
		switch {
		case goal != "":
			action = "resume"
		default:
			action = "status"
		}
	}
	switch action {
	case "status":
		return runPlanStatus(asJSON, path, sessionID)
	case "questions":
		return runPlanQuestions(asJSON, path, sessionID)
	case "resume":
		return runPlanResume(asJSON, path, sessionID, goal, budget)
	default:
		fmt.Fprintf(os.Stderr, "error: unknown plan subcommand: %s\n", action)
		return exitUsage
	}
}

func planSession(path, sessionID, goal string) (planning.PlanningSession, error) {
	if goal != "" {
		return app.FocusSession(path, goal, sessionID)
	}
	if sessionID != "" {
		return app.LoadSession(path, sessionID)
	}
	return app.LatestSession(path)
}

func runPlanStatus(asJSON bool, path, sessionID string) int {
	if sessionID != "" {
		session, err := app.LoadSession(path, sessionID)
		if err != nil {
			return serviceError(asJSON, err)
		}
		return printPlanStatus(asJSON, []app.SessionStatus{app.StatusOf(session)})
	}
	sessions, err := app.ListSessions(path)
	if err != nil {
		return serviceError(asJSON, err)
	}
	statuses := make([]app.SessionStatus, 0, len(sessions))
	for _, s := range sessions {
		statuses = append(statuses, app.StatusOf(s))
	}
	return printPlanStatus(asJSON, statuses)
}

func printPlanStatus(asJSON bool, statuses []app.SessionStatus) int {
	if asJSON {
		return printEnvelope(protocol.OkEnvelope(planSessionData{Sessions: statuses}))
	}
	if len(statuses) == 0 {
		fmt.Println("no planning sessions found")
		return exitOK
	}
	fmt.Println("SESSION ID    SCOPE          GOAL      DEC  OPEN  RESUMABLE")
	for _, s := range statuses {
		fmt.Printf("%-13s %-13s %-9s %3d %4d  %v\n",
			s.SessionID, s.Scope, s.Goal, s.Decisions, s.OpenQuestions, s.Resumable)
	}
	return exitOK
}

func runPlanQuestions(asJSON bool, path, sessionID string) int {
	session, err := planSession(path, sessionID, "")
	if err != nil {
		return serviceError(asJSON, err)
	}
	questions := app.SessionOpenQuestions(session)
	if asJSON {
		return printEnvelope(protocol.OkEnvelope(planQuestionsData{Session: session.ID, Questions: questions}))
	}
	if len(questions) == 0 {
		fmt.Println("no open questions")
		return exitOK
	}
	for _, q := range questions {
		fmt.Printf("- [%s] %s: %s\n", q.Priority, q.ID, questionText(q))
	}
	return exitOK
}

func runPlanResume(asJSON bool, path, sessionID, goal string, budget int) int {
	session, err := planSession(path, sessionID, goal)
	if err != nil {
		return serviceError(asJSON, err)
	}
	bindings, err := app.LoadSessionBindings(path)
	if err != nil {
		return serviceError(asJSON, err)
	}
	report := app.CompileResumeContext(session, bindings, budget)
	effective := budget
	if effective <= 0 {
		effective = app.DefaultResumeBudget
	}
	questions := app.SessionOpenQuestions(session)
	data := planResumeData{
		Session:   report.Session,
		Status:    app.StatusOf(session),
		Budget:    effective,
		Manifest:  report.Manifest,
		Questions: questions,
		Resumable: app.Resumable(report),
	}
	if asJSON {
		return printEnvelope(protocol.OkEnvelope(data))
	}
	fmt.Printf("Session %s | scope %s | decisions %d | open questions %d\n",
		session.ID, session.Scope, len(session.Decisions), len(questions))
	fmt.Printf("Context: %d/%d tokens (%s)\n",
		report.Manifest.EstimatedTokens, effective, report.Manifest.Pressure)
	fmt.Printf("Resumable: %v\n", data.Resumable)
	if q, found := planning.NextOpen(session.Open); found {
		fmt.Printf("Next question [%s %s]: %s\n", q.Priority, q.ID, questionText(q))
	}
	return exitOK
}

func questionText(q planning.OpenQuestion) string {
	if q.Question != "" {
		return q.Question
	}
	if q.Topic != "" {
		return q.Topic
	}
	if q.Reason != "" {
		return q.Reason
	}
	return "see question record"
}
