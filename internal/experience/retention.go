package experience

import (
	"time"
)

// RetentionPolicy governs the expiration, size limits, and archiving of experience data.
type RetentionPolicy struct {
	MaxAgeHours               int  `json:"max_age_hours"`
	MaxEventsPerSession       int  `json:"max_events_per_session"`
	KeepSummariesIndefinitely bool `json:"keep_summaries_indefinitely"`
	PruneRawEvents            bool `json:"prune_raw_events"`
}

// DefaultRetentionPolicy returns pragmatic default settings.
func DefaultRetentionPolicy() RetentionPolicy {
	return RetentionPolicy{
		MaxAgeHours:               720, // 30 days
		MaxEventsPerSession:       1000,
		KeepSummariesIndefinitely: true,
		PruneRawEvents:            true,
	}
}

// ShouldRetainEvent determines if an event is within the allowed retention window.
func (p RetentionPolicy) ShouldRetainEvent(event SessionEvent, now time.Time) bool {
	if p.MaxAgeHours <= 0 {
		return true
	}
	t, err := time.Parse(time.RFC3339Nano, event.Timestamp)
	if err != nil {
		t, err = time.Parse(time.RFC3339, event.Timestamp)
		if err != nil {
			return true
		}
	}
	age := now.Sub(t)
	return age <= time.Duration(p.MaxAgeHours)*time.Hour
}
