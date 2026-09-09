package goals

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func loadGoal(path string) (Goal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var goal Goal
	if err := json.Unmarshal(data, &goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func saveGoal(path string, goal Goal) error {
	if strings.ToLower(filepath.Ext(path)) != ".json" {
		return fmt.Errorf("Legacy YAML Goal is read-only in v0.2; migrate it to .goal.json before changing state.")
	}
	data, err := json.MarshalIndent(goal, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}

func TransitionGoal(path, target, reason string) (Goal, error) {
	goal, err := loadGoal(path)
	if err != nil {
		return nil, err
	}
	current := strings.ToUpper(fmt.Sprint(goal["state"]))
	target = strings.ToUpper(target)
	if !allowedTransition(current, target) {
		return nil, fmt.Errorf("Invalid goal transition: %s -> %s", current, target)
	}
	if target == "DONE" {
		if evidence, ok := goal["evidence"].([]any); !ok || len(evidence) == 0 {
			if ev, ok := goal["evidence"].([]string); !ok || len(ev) == 0 {
				return nil, fmt.Errorf("A goal cannot be marked DONE without evidence.")
			}
		}
	}
	if target == "LOCKED" {
		revision := 1
		if value, ok := goal["revision"].(float64); ok {
			revision = int(value)
		}
		goal["revision"] = revision
		goal["lock"] = map[string]any{
			"revision":  revision,
			"digest":    ComputeDigest(goal),
			"locked_at": time.Now().UTC().Format(time.RFC3339Nano),
		}
	} else if (current == "LOCKED" || current == "EXECUTING" || current == "VERIFYING" || current == "REVIEWING") && goal["lock"] != nil {
		if valid, msg := VerifyLock(goal); !valid {
			return nil, fmt.Errorf("Illegal goal mutation detected: %s", msg)
		}
	}
	goal["state"] = target
	history, _ := goal["history"].([]any)
	goal["history"] = append(history, map[string]any{
		"at": time.Now().UTC().Format(time.RFC3339Nano), "event": "state-transition",
		"from": current, "to": target, "reason": reason,
	})
	if err := saveGoal(path, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func AmendGoal(path string, amendment map[string]any) (Goal, error) {
	goal, err := loadGoal(path)
	if err != nil {
		return nil, err
	}
	current := strings.ToUpper(fmt.Sprint(goal["state"]))
	switch current {
	case "LOCKED", "EXECUTING", "VERIFYING", "REVIEWING", "BLOCKED":
	default:
		return nil, fmt.Errorf("Cannot amend goal in %s state (must be locked/active).", current)
	}
	revision := 1
	if value, ok := goal["revision"].(float64); ok {
		revision = int(value)
	}
	newRev := revision + 1
	goal["revision"] = newRev
	changes, _ := amendment["changes"].(map[string]any)
	if changes == nil {
		changes = map[string]any{}
		if raw, ok := amendment["changes"]; ok && raw != nil {
			_ = raw
		}
	}
	for _, key := range []string{"objective", "acceptance", "constraints", "non_goals", "gates"} {
		if value, ok := changes[key]; ok {
			goal[key] = value
		}
	}
	id, _ := amendment["id"].(string)
	if id == "" {
		id = fmt.Sprintf("AMD-%03d", newRev)
	}
	reason, _ := amendment["reason"].(string)
	if reason == "" {
		reason = "Formal amendment"
	}
	approvedBy, _ := amendment["approved_by"].(string)
	if approvedBy == "" {
		approvedBy = "human"
	}
	approvedAt, _ := amendment["approved_at"].(string)
	if approvedAt == "" {
		approvedAt = time.Now().UTC().Format(time.RFC3339Nano)
	}
	digest := ComputeDigest(goal)
	record := map[string]any{
		"id": id, "goal_id": goal["id"], "revision": newRev, "reason": reason,
		"approved_by": approvedBy, "approved_at": approvedAt, "changes": changes, "digest": digest,
	}
	amendments, _ := goal["amendments"].([]any)
	goal["amendments"] = append(amendments, record)
	goal["lock"] = map[string]any{"revision": newRev, "digest": digest, "locked_at": approvedAt}
	history, _ := goal["history"].([]any)
	goal["history"] = append(history, map[string]any{
		"at": approvedAt, "event": "amended", "revision": newRev, "reason": reason, "approved_by": approvedBy,
	})
	if strings.ToLower(filepath.Ext(path)) != ".json" {
		return nil, fmt.Errorf("Legacy YAML Goal is read-only; migrate to JSON first.")
	}
	if err := saveGoal(path, goal); err != nil {
		return nil, err
	}
	return goal, nil
}
