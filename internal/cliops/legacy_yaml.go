package cliops

import (
	"fmt"
	"strconv"
	"strings"
)

func parseLegacyYAML(input string) (map[string]any, error) {
	lines := []string{}
	for _, raw := range strings.Split(strings.ReplaceAll(input, "\r\n", "\n"), "\n") {
		line := strings.TrimRight(raw, " \t")
		if strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		lines = append(lines, line)
	}
	value, next, err := parseYAMLBlock(lines, 0, indentation(lines, 0))
	if err != nil {
		return nil, err
	}
	if next != len(lines) {
		return nil, fmt.Errorf("unsupported YAML at line %d", next+1)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("legacy YAML root must be a map")
	}
	return object, nil
}

func parseYAMLBlock(lines []string, index, indent int) (any, int, error) {
	if index >= len(lines) {
		return map[string]any{}, index, nil
	}
	if indentation(lines, index) != indent {
		return nil, index, fmt.Errorf("invalid indentation at line %d", index+1)
	}
	if strings.HasPrefix(strings.TrimSpace(lines[index]), "- ") {
		items := []any{}
		for index < len(lines) && indentation(lines, index) == indent && strings.HasPrefix(strings.TrimSpace(lines[index]), "- ") {
			content := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[index]), "- "))
			if strings.Contains(content, ":") && !strings.HasPrefix(content, "[") {
				key, raw, _ := splitYAMLKey(content)
				item := map[string]any{}
				if raw == "" && index+1 < len(lines) && indentation(lines, index+1) > indent {
					child, next, err := parseYAMLBlock(lines, index+1, indentation(lines, index+1))
					if err != nil {
						return nil, index, err
					}
					item[key] = child
					index = next
				} else {
					item[key] = parseYAMLScalar(raw)
					index++
				}
				for index < len(lines) && indentation(lines, index) > indent {
					childKey, childRaw, ok := splitYAMLKey(strings.TrimSpace(lines[index]))
					if !ok {
						return nil, index, fmt.Errorf("unsupported list map at line %d", index+1)
					}
					childIndent := indentation(lines, index)
					if childRaw == "" && index+1 < len(lines) && indentation(lines, index+1) > childIndent {
						child, next, err := parseYAMLBlock(lines, index+1, indentation(lines, index+1))
						if err != nil {
							return nil, index, err
						}
						item[childKey] = child
						index = next
					} else {
						item[childKey] = parseYAMLScalar(childRaw)
						index++
					}
				}
				items = append(items, item)
				continue
			}
			items = append(items, parseYAMLScalar(content))
			index++
		}
		return items, index, nil
	}
	object := map[string]any{}
	for index < len(lines) && indentation(lines, index) == indent {
		key, raw, ok := splitYAMLKey(strings.TrimSpace(lines[index]))
		if !ok {
			return nil, index, fmt.Errorf("unsupported YAML at line %d", index+1)
		}
		if raw == "" && index+1 < len(lines) && indentation(lines, index+1) > indent {
			child, next, err := parseYAMLBlock(lines, index+1, indentation(lines, index+1))
			if err != nil {
				return nil, index, err
			}
			object[key] = child
			index = next
		} else {
			object[key] = parseYAMLScalar(raw)
			index++
		}
	}
	return object, index, nil
}

func indentation(lines []string, index int) int {
	return len(lines[index]) - len(strings.TrimLeft(lines[index], " "))
}

func splitYAMLKey(line string) (string, string, bool) {
	depth := 0
	quote := rune(0)
	for index, char := range line {
		if quote != 0 {
			if char == quote {
				quote = 0
			}
			continue
		}
		if char == '\'' || char == '"' {
			quote = char
			continue
		}
		switch char {
		case '[', '{':
			depth++
		case ']', '}':
			depth--
		case ':':
			if depth == 0 {
				return strings.TrimSpace(line[:index]), strings.TrimSpace(line[index+1:]), true
			}
		}
	}
	return "", "", false
}

func parseYAMLScalar(value string) any {
	value = strings.TrimSpace(value)
	if len(value) >= 2 && ((value[0] == '\'' && value[len(value)-1] == '\'') || (value[0] == '"' && value[len(value)-1] == '"')) {
		return value[1 : len(value)-1]
	}
	if value == "[]" {
		return []any{}
	}
	if value == "{}" {
		return map[string]any{}
	}
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		items := []any{}
		for _, item := range splitYAMLValues(value[1 : len(value)-1]) {
			if strings.TrimSpace(item) != "" {
				items = append(items, parseYAMLScalar(item))
			}
		}
		return items
	}
	if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
		object := map[string]any{}
		for _, item := range splitYAMLValues(value[1 : len(value)-1]) {
			key, raw, ok := splitYAMLKey(strings.TrimSpace(item))
			if ok {
				object[key] = parseYAMLScalar(raw)
			}
		}
		return object
	}
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}
	if value == "null" || value == "~" {
		return nil
	}
	if number, err := strconv.Atoi(value); err == nil {
		return number
	}
	return value
}

func splitYAMLValues(value string) []string {
	items := []string{}
	start := 0
	depth := 0
	quote := rune(0)
	for index, char := range value {
		if quote != 0 {
			if char == quote {
				quote = 0
			}
			continue
		}
		if char == '\'' || char == '"' {
			quote = char
			continue
		}
		switch char {
		case '[', '{':
			depth++
		case ']', '}':
			depth--
		case ',':
			if depth == 0 {
				items = append(items, value[start:index])
				start = index + 1
			}
		}
	}
	items = append(items, value[start:])
	return items
}
