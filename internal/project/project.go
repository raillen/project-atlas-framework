package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/raillen/project-atlas-framework/internal/protocol"
)

var (
	ErrProjectNotFound = protocol.ErrProjectNotFound
	ErrInvalidProject  = protocol.ErrInvalidProject
)

func FindRoot(start string) (string, error) {
	curr, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		manifest := filepath.Join(curr, "atlas.json")
		if _, err := os.Stat(manifest); err == nil {
			return curr, nil
		}
		parent := filepath.Dir(curr)
		if parent == curr {
			return "", ErrProjectNotFound
		}
		curr = parent
	}
}

type Manifest struct {
	Version  int            `json:"version"`
	Protocol map[string]any `json:"protocol"`
	Project  map[string]any `json:"project"`
}

func LoadManifest(root string) (Manifest, error) {
	var manifest Manifest
	data, err := os.ReadFile(filepath.Join(root, "atlas.json"))
	if err != nil {
		return manifest, fmt.Errorf("%w: %s", ErrProjectNotFound, err)
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("%w: %s", ErrInvalidProject, err)
	}
	if manifest.Version < 2 {
		return manifest, fmt.Errorf("%w: unsupported atlas.json version %d", ErrInvalidProject, manifest.Version)
	}
	return manifest, nil
}
