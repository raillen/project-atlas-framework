package plans

import "fmt"

type Task struct {
	ID           string   `json:"id"`
	Dependencies []string `json:"dependencies"`
}

func CheckDAGCycles(tasks []map[string]any) []string {
	errors := []string{}
	taskMap := map[string]map[string]any{}
	for _, task := range tasks {
		if id, ok := task["id"].(string); ok {
			taskMap[id] = task
		}
	}
	graph := map[string][]string{}
	for id, task := range taskMap {
		deps := []string{}
		if raw, ok := task["dependencies"].([]any); ok {
			for _, dep := range raw {
				deps = append(deps, fmt.Sprint(dep))
			}
		}
		graph[id] = deps
	}
	for id, deps := range graph {
		for _, dep := range deps {
			if dep == id {
				errors = append(errors, fmt.Sprintf("Task %s has a self-dependency.", id))
			}
			if _, ok := taskMap[dep]; !ok {
				errors = append(errors, fmt.Sprintf("Task %s depends on non-existent task '%s'.", id, dep))
			}
		}
	}
	visited := map[string]int{}
	var visit func(node string, path []string) bool
	visit = func(node string, path []string) bool {
		visited[node] = 1
		for _, neighbor := range graph[node] {
			if _, ok := graph[neighbor]; !ok {
				continue
			}
			if visited[neighbor] == 1 {
				cycle := append(append([]string{}, path...), neighbor)
				msg := ""
				for i, part := range cycle {
					if i > 0 {
						msg += " -> "
					}
					msg += part
				}
				errors = append(errors, "DAG cycle detected: "+msg)
				return true
			}
			if visited[neighbor] == 0 {
				if visit(neighbor, append(path, neighbor)) {
					return true
				}
			}
		}
		visited[node] = 2
		return false
	}
	for node := range graph {
		if visited[node] == 0 {
			visit(node, []string{node})
		}
	}
	return errors
}
