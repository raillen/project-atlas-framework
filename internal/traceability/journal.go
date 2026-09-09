package traceability

import (
	"fmt"
	"strings"
	"time"
)

// JournalFilter provides search parameters for querying the implementation journal.
type JournalFilter struct {
	Goal     string `json:"goal,omitempty"`
	Decision string `json:"decision,omitempty"`
	File     string `json:"file,omitempty"`
}

// ImplementationJournal manages structured synthesis entries across development goals.
type ImplementationJournal struct {
	Entries []JournalEntry `json:"entries"`
}

// NewJournal initializes an empty ImplementationJournal.
func NewJournal() *ImplementationJournal {
	return &ImplementationJournal{
		Entries: make([]JournalEntry, 0),
	}
}

// Add appends a new verified journal entry.
// Invariant: Summary must be synthesis; raw chain-of-thought or conversational telemetry is prohibited.
func (j *ImplementationJournal) Add(e JournalEntry) error {
	if strings.TrimSpace(e.ID) == "" {
		return fmt.Errorf("journal entry requires an ID")
	}
	if strings.TrimSpace(e.Summary) == "" {
		return fmt.Errorf("journal entry requires a summary")
	}
	if e.CreatedAt == "" {
		e.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	j.Entries = append(j.Entries, e)
	return nil
}

// FindByID retrieves a journal entry by its unique identifier.
func (j *ImplementationJournal) FindByID(id string) (JournalEntry, bool) {
	for _, e := range j.Entries {
		if e.ID == id {
			return e, true
		}
	}
	return JournalEntry{}, false
}

// Query filters journal entries matching the criteria.
func (j *ImplementationJournal) Query(filter JournalFilter) []JournalEntry {
	var matches []JournalEntry
	for _, e := range j.Entries {
		if filter.Goal != "" && !strings.EqualFold(e.Goal, filter.Goal) {
			continue
		}
		if filter.Decision != "" {
			foundDec := false
			for _, d := range e.Decisions {
				if strings.EqualFold(d, filter.Decision) {
					foundDec = true
					break
				}
			}
			if !foundDec {
				continue
			}
		}
		if filter.File != "" {
			foundFile := false
			for _, f := range e.CodeChanges {
				if strings.Contains(strings.ToLower(f), strings.ToLower(filter.File)) {
					foundFile = true
					break
				}
			}
			if !foundFile {
				for _, d := range e.ChangedDocs {
					if strings.Contains(strings.ToLower(d), strings.ToLower(filter.File)) {
						foundFile = true
						break
					}
				}
			}
			if !foundFile {
				continue
			}
		}
		matches = append(matches, e)
	}
	return matches
}
