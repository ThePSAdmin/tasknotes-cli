package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestVault(t *testing.T, dataJSON string) string {
	t.Helper()
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, ".obsidian", "plugins", "tasknotes")
	if err := os.MkdirAll(pluginDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pluginDir, "data.json"), []byte(dataJSON), 0644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoadConfig_Defaults(t *testing.T) {
	// Vault with no data.json should return defaults
	dir := t.TempDir()
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TasksFolder != "TaskNotes/Tasks" {
		t.Errorf("expected default tasks folder 'TaskNotes/Tasks', got %q", cfg.TasksFolder)
	}
	if cfg.ArchiveFolder != "TaskNotes/Archive" {
		t.Errorf("expected default archive folder 'TaskNotes/Archive', got %q", cfg.ArchiveFolder)
	}
	if cfg.TaskTag != "task" {
		t.Errorf("expected default task tag 'task', got %q", cfg.TaskTag)
	}
}

func TestLoadConfig_FromDataJSON(t *testing.T) {
	data := `{
		"tasksFolder": "MyTasks",
		"archiveFolder": "MyArchive",
		"taskTag": "todo",
		"fieldMapping": {
			"title": "name",
			"status": "state",
			"priority": "urgency"
		},
		"customStatuses": [
			{"id": "open", "value": "open", "label": "Open", "isCompleted": false, "order": 1},
			{"id": "done", "value": "done", "label": "Done", "isCompleted": true, "order": 2}
		],
		"customPriorities": [
			{"id": "low", "value": "low", "label": "Low", "weight": 1},
			{"id": "high", "value": "high", "label": "High", "weight": 3}
		]
	}`
	dir := setupTestVault(t, data)
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.TasksFolder != "MyTasks" {
		t.Errorf("expected 'MyTasks', got %q", cfg.TasksFolder)
	}
	if cfg.ArchiveFolder != "MyArchive" {
		t.Errorf("expected 'MyArchive', got %q", cfg.ArchiveFolder)
	}
	if cfg.TaskTag != "todo" {
		t.Errorf("expected 'todo', got %q", cfg.TaskTag)
	}
	if cfg.FieldMapping.Title != "name" {
		t.Errorf("expected field mapping title='name', got %q", cfg.FieldMapping.Title)
	}
	if cfg.FieldMapping.Status != "state" {
		t.Errorf("expected field mapping status='state', got %q", cfg.FieldMapping.Status)
	}
	if len(cfg.Statuses) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(cfg.Statuses))
	}
	if len(cfg.Priorities) != 2 {
		t.Errorf("expected 2 priorities, got %d", len(cfg.Priorities))
	}
}

func TestLoadConfig_FieldMappingDefaults(t *testing.T) {
	data := `{}`
	dir := setupTestVault(t, data)
	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.FieldMapping.Title != "title" {
		t.Errorf("expected default field mapping title='title', got %q", cfg.FieldMapping.Title)
	}
	if cfg.FieldMapping.BlockedBy != "blockedBy" {
		t.Errorf("expected default field mapping blockedBy='blockedBy', got %q", cfg.FieldMapping.BlockedBy)
	}
	if cfg.FieldMapping.TimeEntries != "timeEntries" {
		t.Errorf("expected default field mapping timeEntries='timeEntries', got %q", cfg.FieldMapping.TimeEntries)
	}
}

func TestStatusIsCompleted(t *testing.T) {
	cfg := &Config{
		Statuses: []StatusConfig{
			{ID: "open", Value: "open", IsCompleted: false},
			{ID: "done", Value: "done", IsCompleted: true},
		},
	}
	if cfg.IsCompletedStatus("open") {
		t.Error("open should not be completed")
	}
	if !cfg.IsCompletedStatus("done") {
		t.Error("done should be completed")
	}
}

func TestPriorityWeight(t *testing.T) {
	cfg := &Config{
		Priorities: []PriorityConfig{
			{ID: "low", Value: "low", Weight: 1},
			{ID: "normal", Value: "normal", Weight: 2},
			{ID: "high", Value: "high", Weight: 3},
		},
	}
	if w := cfg.PriorityWeight("low"); w != 1 {
		t.Errorf("expected weight 1, got %d", w)
	}
	if w := cfg.PriorityWeight("high"); w != 3 {
		t.Errorf("expected weight 3, got %d", w)
	}
	if w := cfg.PriorityWeight("unknown"); w != 0 {
		t.Errorf("expected weight 0 for unknown, got %d", w)
	}
}
