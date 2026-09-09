package goals

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

var States = []string{"DRAFT", "PLANNED", "LOCKED", "EXECUTING", "VERIFYING", "REVIEWING", "BLOCKED", "DONE"}

var Transitions = map[string][]string{
	"DRAFT":     {"PLANNED", "BLOCKED"},
	"PLANNED":   {"LOCKED", "DRAFT", "BLOCKED"},
	"LOCKED":    {"EXECUTING", "BLOCKED"},
	"EXECUTING": {"VERIFYING", "BLOCKED"},
	"VERIFYING": {"REVIEWING", "EXECUTING", "BLOCKED"},
	"REVIEWING": {"DONE", "EXECUTING", "BLOCKED"},
	"BLOCKED":   {"PLANNED", "LOCKED", "EXECUTING"},
	"DONE":      {},
}

type Goal map[string]any

func strList(value any) []string {
	out := []string{}
	switch items := value.(type) {
	case []any:
		for _, item := range items {
			out = append(out, strings.TrimSpace(fmt.Sprint(item)))
		}
	case []string:
		for _, item := range items {
			out = append(out, strings.TrimSpace(item))
		}
	}
	sort.Strings(out)
	return out
}

func ComputeDigest(goal Goal) string {
	payload := map[string]any{
		"acceptance":  strList(goal["acceptance"]),
		"constraints": strList(goal["constraints"]),
		"gates":       goal["gates"],
		"non_goals":   strList(goal["non_goals"]),
		"objective":   strings.TrimSpace(fmt.Sprint(goal["objective"])),
	}
	encoded, _ := json.Marshal(payload)
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func VerifyLock(goal Goal) (bool, string) {
	state := strings.ToUpper(fmt.Sprint(goal["state"]))
	lock, hasLock := goal["lock"].(map[string]any)
	if !hasLock {
		if state == "LOCKED" || state == "EXECUTING" || state == "VERIFYING" || state == "REVIEWING" || state == "DONE" {
			return false, "Goal is in locked/executing state but lacks a lock record."
		}
		return true, "Goal is not locked."
	}
	expected, _ := lock["digest"].(string)
	current := ComputeDigest(goal)
	if expected != current {
		return false, fmt.Sprintf("Lock digest mismatch: expected %s, got %s", expected, current)
	}
	return true, "Lock integrity valid."
}

func NewGoal(id, title, phase, objective string) Goal {
	if objective == "" {
		objective = fmt.Sprintf("Define the measurable outcome for %s.", title)
	}
	return Goal{
		"id":          id,
		"title":       title,
		"phase":       phase,
		"revision":    1,
		"state":       "DRAFT",
		"objective":   objective,
		"constraints": []string{},
		"non_goals":   []string{},
		"acceptance":  []string{"Replace this placeholder with objective acceptance criteria."},
		"gates": map[string]any{
			"build": "required", "tests": "required", "review": "required",
			"documentation_impact": "required", "project_intelligence": "required",
		},
		"context":      map[string]any{"budget_profile": "medium", "max_delegation_depth": 1},
		"dependencies": []string{},
		"evidence":     []string{},
		"history": []any{map[string]any{
			"at": time.Now().UTC().Format(time.RFC3339Nano), "event": "created", "state": "DRAFT",
		}},
	}
}

func allowedTransition(current, target string) bool {
	for _, next := range Transitions[current] {
		if next == target {
			return true
		}
	}
	return false
}
