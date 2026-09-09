package docengine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

var deltaIDPattern = regexp.MustCompile(`^DD-[0-9a-f]{12}\.json$`)

func NewDelta(source string, impacts []Impact, now time.Time) Delta {
	proposal := MakeDelta(source, impacts)
	stamp := now.UTC().Format(time.RFC3339Nano)
	return Delta{
		ID:        DeltaID(source, proposal.Contracts, proposal.Documents, proposal.Reason),
		Version:   1,
		State:     proposal.State,
		Source:    proposal.Source,
		Contracts: proposal.Contracts,
		Documents: proposal.Documents,
		Reason:    proposal.Reason,
		Evidence:  proposal.Evidence,
		CreatedAt: stamp,
		UpdatedAt: stamp,
	}
}

func DeltaID(source string, contracts, documents []string, reason string) string {
	payload := map[string]any{
		"contracts": append([]string{}, contracts...),
		"documents": append([]string{}, documents...),
		"reason":    reason,
		"source":    source,
	}
	sort.Strings(payload["contracts"].([]string))
	sort.Strings(payload["documents"].([]string))
	encoded, _ := json.Marshal(payload)
	sum := sha256.Sum256(encoded)
	return "DD-" + hex.EncodeToString(sum[:])[:12]
}

func (d Delta) Transition(target string, evidence []string, now time.Time) (Delta, error) {
	allowed := map[string][]string{
		"proposed": {"reviewed", "rejected"},
		"reviewed": {"accepted", "rejected"},
		"accepted": {"applied"},
	}
	next := false
	for _, candidate := range allowed[d.State] {
		if candidate == target {
			next = true
			break
		}
	}
	if !next {
		return Delta{}, fmt.Errorf("invalid documentation delta transition: %s -> %s", d.State, target)
	}
	merged := append([]string{}, d.Evidence...)
	for _, item := range evidence {
		found := false
		for _, existing := range merged {
			if existing == item {
				found = true
				break
			}
		}
		if !found {
			merged = append(merged, item)
		}
	}
	updated := d
	updated.State = target
	updated.Evidence = merged
	updated.Version = d.Version + 1
	updated.UpdatedAt = now.UTC().Format(time.RFC3339Nano)
	if err := updated.Validate(); err != nil {
		return Delta{}, err
	}
	return updated, nil
}

func deltaPath(root, id string) (string, error) {
	filename := id + ".json"
	if !deltaIDPattern.MatchString(filename) {
		return "", fmt.Errorf("invalid documentation delta id: %s", id)
	}
	return filepath.Join(root, ".ai", "docs", "deltas", filename), nil
}

func SaveDelta(root string, delta Delta) error {
	if err := delta.Validate(); err != nil {
		return err
	}
	if delta.Version < 1 {
		return fmt.Errorf("tracked documentation delta requires version 1 or later")
	}
	path, err := deltaPath(root, delta.ID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	encoded, err := json.MarshalIndent(delta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0644)
}

func LoadDelta(root, id string) (Delta, error) {
	path, err := deltaPath(root, id)
	if err != nil {
		return Delta{}, err
	}
	return loadDeltaFile(path)
}

func ListDeltas(root string) ([]Delta, error) {
	entries, err := os.ReadDir(filepath.Join(root, ".ai", "docs", "deltas"))
	if err != nil {
		if os.IsNotExist(err) {
			return []Delta{}, nil
		}
		return nil, err
	}
	deltas := []Delta{}
	for _, entry := range entries {
		if entry.IsDir() || !deltaIDPattern.MatchString(entry.Name()) {
			continue
		}
		delta, err := loadDeltaFile(filepath.Join(root, ".ai", "docs", "deltas", entry.Name()))
		if err != nil {
			return nil, err
		}
		deltas = append(deltas, delta)
	}
	return deltas, nil
}

func TransitionDelta(root, id, target string, evidence []string, now time.Time) (Delta, error) {
	current, err := LoadDelta(root, id)
	if err != nil {
		return Delta{}, err
	}
	updated, err := current.Transition(target, evidence, now)
	if err != nil {
		return Delta{}, err
	}
	if err := SaveDelta(root, updated); err != nil {
		return Delta{}, err
	}
	return updated, nil
}

func loadDeltaFile(path string) (Delta, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Delta{}, err
	}
	var delta Delta
	if err := json.Unmarshal(data, &delta); err != nil {
		return Delta{}, fmt.Errorf("invalid documentation delta %s: %w", path, err)
	}
	if delta.ID+".json" != filepath.Base(path) {
		return Delta{}, fmt.Errorf("documentation delta id does not match filename: %s", path)
	}
	if err := delta.Validate(); err != nil {
		return Delta{}, fmt.Errorf("invalid documentation delta %s: %w", path, err)
	}
	return delta, nil
}
