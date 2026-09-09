package traceability

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	GraphFileName     = "graph.json"
	JournalFileName   = "journal.json"
	RegistersFileName = "registers.json"
)

// SaveGraph writes the graph to the specified directory.
func SaveGraph(dir string, g *Graph) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, GraphFileName), data, 0644)
}

// LoadGraph reads the graph from the directory. If missing, returns a new empty graph.
func LoadGraph(dir string) (*Graph, error) {
	file := filepath.Join(dir, GraphFileName)
	data, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return NewGraph(), nil
	}
	if err != nil {
		return nil, err
	}
	var g Graph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	if g.Nodes == nil {
		g.Nodes = make(map[string]Node)
	}
	return &g, nil
}

// SaveJournal writes the journal to the directory.
func SaveJournal(dir string, j *ImplementationJournal) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, JournalFileName), data, 0644)
}

// LoadJournal reads the journal from the directory.
func LoadJournal(dir string) (*ImplementationJournal, error) {
	file := filepath.Join(dir, JournalFileName)
	data, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return NewJournal(), nil
	}
	if err != nil {
		return nil, err
	}
	var j ImplementationJournal
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, err
	}
	return &j, nil
}

// SaveRegisters writes registers to the directory.
func SaveRegisters(dir string, r *Registers) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, RegistersFileName), data, 0644)
}

// LoadRegisters reads registers from the directory.
func LoadRegisters(dir string) (*Registers, error) {
	file := filepath.Join(dir, RegistersFileName)
	data, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return NewRegisters(), nil
	}
	if err != nil {
		return nil, err
	}
	var r Registers
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
