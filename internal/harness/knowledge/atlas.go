// Memory Atlas (GAP-018 first slice, page 25 direction): a project-scoped,
// file-backed memory index feeding the Context Compiler. Records carry
// provenance/freshness; recall is lexical and bounded. Cross-project
// promotion with privacy gates remains future work.
package knowledge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// AtlasPath is the conventional project-local index location.
func AtlasPath(root string) string { return filepath.Join(root, ".prumo", "atlas.json") }

// Atlas is the serializable memory index.
type Atlas struct {
	Project string   `json:"project"`
	Records []Record `json:"records"`
	Updated string   `json:"updated"`
}

// LoadAtlas reads the index; missing file yields an empty atlas.
func LoadAtlas(path string) (Atlas, error) {
	var a Atlas
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return a, nil
		}
		return a, err
	}
	if err := json.Unmarshal(data, &a); err != nil {
		return a, err
	}
	return a, nil
}

// SaveAtlas writes atomically.
func SaveAtlas(path string, a Atlas) error {
	a.Updated = time.Now().UTC().Format(time.RFC3339Nano)
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// Remember upserts one memory record.
func Remember(s *Store, id, title, body, provenance string, tags []string) error {
	meta := map[string]any{}
	if len(tags) > 0 {
		joined := make([]any, 0, len(tags))
		for _, tg := range tags {
			joined = append(joined, tg)
		}
		meta["tags"] = joined
	}
	return s.Commit(Delta{ID: "remember-" + id, Author: "harness-run", Upserts: []Record{{
		ID: id, Kind: KindMemory, Title: title, Body: body,
		Authority: "reference", Trust: "medium", Status: "active",
		Provenance: provenance, UpdatedAt: time.Now().UTC().Format(time.RFC3339Nano), Meta: meta,
	}}})
}

// Recall returns up to n memory records overlapping query terms,
// freshest most-overlapping first.
func Recall(s *Store, query string, n int) []Record {
	if n <= 0 {
		n = 5
	}
	terms := map[string]bool{}
	for _, tok := range strings.Fields(strings.ToLower(query)) {
		if len(tok) > 1 {
			terms[tok] = true
		}
	}
	type hit struct {
		r     Record
		score int
	}
	var hits []hit
	for _, r := range s.Search("", []Kind{KindMemory}) {
		if r.Status != "active" {
			continue
		}
		hay := strings.ToLower(r.Title + " " + r.Body)
		score := 0
		for term := range terms {
			if strings.Contains(hay, term) {
				score++
			}
		}
		if score > 0 {
			hits = append(hits, hit{r, score})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		return hits[i].r.UpdatedAt > hits[j].r.UpdatedAt
	})
	out := []Record{}
	for i := range hits {
		if len(out) >= n {
			break
		}
		out = append(out, hits[i].r)
	}
	return out
}

// SnapshotAtlas exports a store's memory records as an Atlas.
func SnapshotAtlas(project string, s *Store) Atlas {
	records, _ := s.Snapshot()
	mem := []Record{}
	for _, r := range records {
		if r.Kind == KindMemory {
			mem = append(mem, r)
		}
	}
	sort.Slice(mem, func(i, j int) bool { return mem[i].ID < mem[j].ID })
	return Atlas{Project: project, Records: mem}
}
