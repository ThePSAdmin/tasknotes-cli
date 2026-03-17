package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// FieldMapping maps internal field names to user-configured YAML property names.
type FieldMapping struct {
	Title               string `json:"title"`
	Status              string `json:"status"`
	Priority            string `json:"priority"`
	Due                 string `json:"due"`
	Scheduled           string `json:"scheduled"`
	Contexts            string `json:"contexts"`
	Projects            string `json:"projects"`
	TimeEstimate        string `json:"timeEstimate"`
	CompletedDate       string `json:"completedDate"`
	DateCreated         string `json:"dateCreated"`
	DateModified        string `json:"dateModified"`
	Recurrence          string `json:"recurrence"`
	RecurrenceAnchor    string `json:"recurrenceAnchor"`
	ArchiveTag          string `json:"archiveTag"`
	TimeEntries         string `json:"timeEntries"`
	CompleteInstances   string `json:"completeInstances"`
	SkippedInstances    string `json:"skippedInstances"`
	BlockedBy           string `json:"blockedBy"`
	Pomodoros           string `json:"pomodoros"`
	ICSEventID          string `json:"icsEventId"`
	ICSEventTag         string `json:"icsEventTag"`
	GoogleCalendarEvtID string `json:"googleCalendarEventId"`
	Reminders           string `json:"reminders"`
}

// StatusConfig defines a task status.
type StatusConfig struct {
	ID          string `json:"id"`
	Value       string `json:"value"`
	Label       string `json:"label"`
	IsCompleted bool   `json:"isCompleted"`
	Order       int    `json:"order"`
}

// PriorityConfig defines a task priority.
type PriorityConfig struct {
	ID     string `json:"id"`
	Value  string `json:"value"`
	Label  string `json:"label"`
	Weight int    `json:"weight"`
}

// Config holds all settings read from the vault's plugin data.json.
type Config struct {
	VaultPath    string
	TasksFolder  string           `json:"tasksFolder"`
	ArchiveFolder string          `json:"archiveFolder"`
	TaskTag      string           `json:"taskTag"`
	FieldMapping FieldMapping     `json:"fieldMapping"`
	Statuses     []StatusConfig   `json:"customStatuses"`
	Priorities   []PriorityConfig `json:"customPriorities"`
}

// DefaultFieldMapping returns the default field mapping.
func DefaultFieldMapping() FieldMapping {
	return FieldMapping{
		Title:               "title",
		Status:              "status",
		Priority:            "priority",
		Due:                 "due",
		Scheduled:           "scheduled",
		Contexts:            "contexts",
		Projects:            "projects",
		TimeEstimate:        "timeEstimate",
		CompletedDate:       "completedDate",
		DateCreated:         "dateCreated",
		DateModified:        "dateModified",
		Recurrence:          "recurrence",
		RecurrenceAnchor:    "recurrence_anchor",
		ArchiveTag:          "archived",
		TimeEntries:         "timeEntries",
		CompleteInstances:   "complete_instances",
		SkippedInstances:    "skipped_instances",
		BlockedBy:           "blockedBy",
		Pomodoros:           "pomodoros",
		ICSEventID:          "icsEventId",
		ICSEventTag:         "ics_event",
		GoogleCalendarEvtID: "googleCalendarEventId",
		Reminders:           "reminders",
	}
}

// DefaultStatuses returns the default status configuration.
func DefaultStatuses() []StatusConfig {
	return []StatusConfig{
		{ID: "none", Value: "none", Label: "None", IsCompleted: false, Order: 0},
		{ID: "open", Value: "open", Label: "Open", IsCompleted: false, Order: 1},
		{ID: "in-progress", Value: "in-progress", Label: "In progress", IsCompleted: false, Order: 2},
		{ID: "done", Value: "done", Label: "Done", IsCompleted: true, Order: 3},
	}
}

// DefaultPriorities returns the default priority configuration.
func DefaultPriorities() []PriorityConfig {
	return []PriorityConfig{
		{ID: "none", Value: "none", Label: "None", Weight: 0},
		{ID: "low", Value: "low", Label: "Low", Weight: 1},
		{ID: "normal", Value: "normal", Label: "Normal", Weight: 2},
		{ID: "high", Value: "high", Label: "High", Weight: 3},
	}
}

// Load reads configuration from the vault's plugin data.json.
// If the file doesn't exist, defaults are returned.
func Load(vaultPath string) (*Config, error) {
	cfg := &Config{
		VaultPath:     vaultPath,
		TasksFolder:   "TaskNotes/Tasks",
		ArchiveFolder: "TaskNotes/Archive",
		TaskTag:       "task",
		FieldMapping:  DefaultFieldMapping(),
		Statuses:      DefaultStatuses(),
		Priorities:    DefaultPriorities(),
	}

	dataPath := filepath.Join(vaultPath, ".obsidian", "plugins", "tasknotes", "data.json")
	data, err := os.ReadFile(dataPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	// Parse into a raw structure to selectively override defaults
	var raw struct {
		TasksFolder   string           `json:"tasksFolder"`
		ArchiveFolder string           `json:"archiveFolder"`
		TaskTag       string           `json:"taskTag"`
		FieldMapping  *json.RawMessage `json:"fieldMapping"`
		Statuses      []StatusConfig   `json:"customStatuses"`
		Priorities    []PriorityConfig `json:"customPriorities"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	if raw.TasksFolder != "" {
		cfg.TasksFolder = raw.TasksFolder
	}
	if raw.ArchiveFolder != "" {
		cfg.ArchiveFolder = raw.ArchiveFolder
	}
	if raw.TaskTag != "" {
		cfg.TaskTag = raw.TaskTag
	}
	if raw.FieldMapping != nil {
		// Start with defaults, overlay with provided values
		fm := DefaultFieldMapping()
		if err := json.Unmarshal(*raw.FieldMapping, &fm); err != nil {
			return nil, err
		}
		cfg.FieldMapping = fm
	}
	if len(raw.Statuses) > 0 {
		cfg.Statuses = raw.Statuses
	}
	if len(raw.Priorities) > 0 {
		cfg.Priorities = raw.Priorities
	}

	return cfg, nil
}

// IsCompletedStatus returns true if the given status value is marked as completed.
func (c *Config) IsCompletedStatus(status string) bool {
	for _, s := range c.Statuses {
		if s.Value == status {
			return s.IsCompleted
		}
	}
	return false
}

// PriorityWeight returns the weight for a priority value, or 0 if unknown.
func (c *Config) PriorityWeight(priority string) int {
	for _, p := range c.Priorities {
		if p.Value == priority {
			return p.Weight
		}
	}
	return 0
}

// CompletedStatusValue returns the first status value marked as completed.
func (c *Config) CompletedStatusValue() string {
	for _, s := range c.Statuses {
		if s.IsCompleted {
			return s.Value
		}
	}
	return "done"
}
