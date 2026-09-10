package resources

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var (
	DefaultFS       fs.FS
	AdaptersFS      fs.FS
	WorkforceFS     fs.FS
	embeddedSchemas fs.FS
)

func safeResourceName(name string) (string, error) {
	if name == "" || name == "." || filepath.IsAbs(name) {
		return "", fs.ErrInvalid
	}
	if strings.ContainsRune(name, '\\') {
		return "", fs.ErrInvalid
	}
	clean := filepath.Clean(name)
	if clean != name || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fs.ErrInvalid
	}
	return name, nil
}

func FindSchemasDir(repoRoot string) string {
	return filepath.Join(repoRoot, "schemas")
}

func FindResourcesDir(repoRoot string) string {
	return filepath.Join(repoRoot, "src", "prumo", "resources")
}

func OpenSchema(repoRoot, schemaName string) ([]byte, error) {
	name, err := safeResourceName(schemaName)
	if err != nil {
		return nil, err
	}
	if DefaultFS != nil {
		return fs.ReadFile(DefaultFS, filepath.Join("schemas", name))
	}
	return os.ReadFile(filepath.Join(FindSchemasDir(repoRoot), name))
}

func OpenResourceFile(repoRoot string, parts ...string) ([]byte, error) {
	if len(parts) == 0 {
		return nil, fs.ErrInvalid
	}
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		name, err := safeResourceName(part)
		if err != nil {
			return nil, err
		}
		clean = append(clean, name)
	}
	if DefaultFS != nil {
		return fs.ReadFile(DefaultFS, filepath.Join(clean...))
	}
	return os.ReadFile(filepath.Join(append([]string{FindResourcesDir(repoRoot)}, clean...)...))
}

func ListResources(repoRoot, dir string) ([]string, error) {
	name, err := safeResourceName(dir)
	if err != nil {
		return nil, err
	}
	if DefaultFS != nil {
		entries, err := fs.ReadDir(DefaultFS, name)
		if err != nil {
			return nil, err
		}
		out := make([]string, 0, len(entries))
		for _, entry := range entries {
			out = append(out, entry.Name())
		}
		return out, nil
	}
	entries, err := os.ReadDir(filepath.Join(FindResourcesDir(repoRoot), name))
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry.Name())
	}
	return out, nil
}
