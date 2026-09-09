package adoption

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ParsedManifest contains static metadata extracted from a manifest file.
type ParsedManifest struct {
	Name         string            `json:"name,omitempty"`
	Version      string            `json:"version,omitempty"`
	Dependencies []string          `json:"dependencies,omitempty"`
	Scripts      map[string]string `json:"scripts,omitempty"`
}

// ParseManifestFile extracts metadata and dependencies from a manifest file.
func ParseManifestFile(filename, fullPath string) (ParsedManifest, error) {
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return ParsedManifest{}, err
	}

	base := strings.ToLower(filepath.Base(filename))
	switch base {
	case "package.json":
		return parsePackageJSON(data)
	case "go.mod":
		return parseGoMod(data)
	case "pyproject.toml":
		return parsePyprojectTOML(data)
	case "requirements.txt":
		return parseRequirementsTxt(data)
	case "cargo.toml":
		return parseCargoTOML(data)
	default:
		return ParsedManifest{}, nil
	}
}

func parsePackageJSON(data []byte) (ParsedManifest, error) {
	var raw struct {
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
		Scripts         map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return ParsedManifest{}, err
	}

	deps := make([]string, 0, len(raw.Dependencies)+len(raw.DevDependencies))
	for k := range raw.Dependencies {
		deps = append(deps, k)
	}
	for k := range raw.DevDependencies {
		deps = append(deps, k)
	}

	return ParsedManifest{
		Name:         raw.Name,
		Version:      raw.Version,
		Dependencies: deps,
		Scripts:      raw.Scripts,
	}, nil
}

func parseGoMod(data []byte) (ParsedManifest, error) {
	pm := ParsedManifest{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	inRequireBlock := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "//") || line == "" {
			continue
		}

		if strings.HasPrefix(line, "module ") {
			pm.Name = strings.TrimSpace(strings.TrimPrefix(line, "module "))
			continue
		}

		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}
		if inRequireBlock {
			if line == ")" {
				inRequireBlock = false
				continue
			}
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				pm.Dependencies = append(pm.Dependencies, parts[0])
			}
			continue
		}

		if strings.HasPrefix(line, "require ") {
			rest := strings.TrimPrefix(line, "require ")
			parts := strings.Fields(rest)
			if len(parts) >= 1 {
				pm.Dependencies = append(pm.Dependencies, parts[0])
			}
		}
	}

	return pm, scanner.Err()
}

func parseRequirementsTxt(data []byte) (ParsedManifest, error) {
	pm := ParsedManifest{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" || strings.HasPrefix(line, "-") {
			continue
		}
		// Split by comparison operators: ==, >=, <=, ~=, !=, <, >
		dep := line
		for _, op := range []string{"==", ">=", "<=", "~=", "!=", "<", ">", ";", "["} {
			if idx := strings.Index(dep, op); idx != -1 {
				dep = dep[:idx]
			}
		}
		dep = strings.TrimSpace(dep)
		if dep != "" {
			pm.Dependencies = append(pm.Dependencies, dep)
		}
	}
	return pm, scanner.Err()
}

func parsePyprojectTOML(data []byte) (ParsedManifest, error) {
	pm := ParsedManifest{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	inDeps := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if strings.HasPrefix(line, "name =") || strings.HasPrefix(line, "name=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				pm.Name = strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
			continue
		}

		if strings.HasPrefix(line, "dependencies = [") || strings.HasPrefix(line, "dependencies=[") {
			if strings.HasSuffix(line, "]") {
				// Inline array
				content := line[strings.Index(line, "[")+1 : strings.LastIndex(line, "]")]
				for _, item := range strings.Split(content, ",") {
					item = strings.Trim(strings.TrimSpace(item), `"'`)
					dep := cleanDep(item)
					if dep != "" {
						pm.Dependencies = append(pm.Dependencies, dep)
					}
				}
			} else {
				inDeps = true
			}
			continue
		}

		if inDeps {
			if strings.HasPrefix(line, "]") {
				inDeps = false
				continue
			}
			item := strings.Trim(strings.TrimSpace(line), `"',`)
			dep := cleanDep(item)
			if dep != "" {
				pm.Dependencies = append(pm.Dependencies, dep)
			}
		}
	}
	return pm, scanner.Err()
}

func parseCargoTOML(data []byte) (ParsedManifest, error) {
	pm := ParsedManifest{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	inDeps := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		if line == "[dependencies]" || line == "[dev-dependencies]" {
			inDeps = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inDeps = false
			continue
		}

		if strings.HasPrefix(line, "name =") || strings.HasPrefix(line, "name=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				pm.Name = strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			}
			continue
		}

		if inDeps {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) >= 1 {
				dep := strings.TrimSpace(parts[0])
				if dep != "" {
					pm.Dependencies = append(pm.Dependencies, dep)
				}
			}
		}
	}
	return pm, scanner.Err()
}

func cleanDep(item string) string {
	for _, op := range []string{"==", ">=", "<=", "~=", "!=", "<", ">", ";", "["} {
		if idx := strings.Index(item, op); idx != -1 {
			item = item[:idx]
		}
	}
	return strings.TrimSpace(item)
}
