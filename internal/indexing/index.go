package indexing

type Entry struct {
	Path      string `json:"path"`
	Hash      string `json:"hash"`
	Size      int64  `json:"size"`
	UpdatedAt string `json:"updated_at"`
}
type Index struct {
	Version  int     `json:"version"`
	Revision string  `json:"revision"`
	Entries  []Entry `json:"entries"`
}

func Changed(previous, current Index) []string {
	old := map[string]Entry{}
	for _, e := range previous.Entries {
		old[e.Path] = e
	}
	out := []string{}
	for _, e := range current.Entries {
		if before, ok := old[e.Path]; !ok || before.Hash != e.Hash || before.Size != e.Size {
			out = append(out, e.Path)
		}
	}
	return out
}
