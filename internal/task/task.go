package task

import "time"

// Task represents a parsed task from a markdown file.
type Task struct {
	Path              string       `json:"path"`
	Title             string       `json:"title"`
	Status            string       `json:"status"`
	Priority          string       `json:"priority"`
	Due               string       `json:"due,omitempty"`
	Scheduled         string       `json:"scheduled,omitempty"`
	Tags              []string     `json:"tags,omitempty"`
	Contexts          []string     `json:"contexts,omitempty"`
	Projects          []string     `json:"projects,omitempty"`
	TimeEstimate      int          `json:"timeEstimate,omitempty"`
	CompletedDate     string       `json:"completedDate,omitempty"`
	DateCreated       string       `json:"dateCreated,omitempty"`
	DateModified      string       `json:"dateModified,omitempty"`
	Recurrence        string       `json:"recurrence,omitempty"`
	RecurrenceAnchor  string       `json:"recurrenceAnchor,omitempty"`
	CompleteInstances []string     `json:"completeInstances,omitempty"`
	SkippedInstances  []string     `json:"skippedInstances,omitempty"`
	BlockedBy         []Dependency `json:"blockedBy,omitempty"`
	TimeEntries       []TimeEntry  `json:"timeEntries,omitempty"`
	Body              string       `json:"body,omitempty"`

	// Computed fields (not stored in frontmatter)
	Blocking         []string `json:"blocking,omitempty"`
	TotalTrackedTime int      `json:"totalTrackedTime,omitempty"`
	IsBlocked        bool     `json:"isBlocked,omitempty"`
}

// Dependency represents a task dependency.
type Dependency struct {
	UID     string `yaml:"uid" json:"uid"`
	RelType string `yaml:"reltype" json:"reltype"`
	Gap     string `yaml:"gap,omitempty" json:"gap,omitempty"`
}

// TimeEntry represents a time tracking entry.
type TimeEntry struct {
	StartTime   time.Time `yaml:"startTime" json:"startTime"`
	EndTime     time.Time `yaml:"endTime,omitempty" json:"endTime,omitempty"`
	Description string    `yaml:"description,omitempty" json:"description,omitempty"`
}
