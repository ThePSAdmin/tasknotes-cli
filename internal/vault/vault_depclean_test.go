package vault

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thepsadmin/tasknotes-cli/internal/config"
	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

func TestDeleteTask_RemovesDependencyReferences(t *testing.T) {
	dir, cfg := setupVault(t)

	// Task A blocks Task B
	writeTask(t, dir, "task-a.md", `---
title: Task A
status: open
priority: normal
tags:
  - task
---
`)
	writeTask(t, dir, "task-b.md", `---
title: Task B
status: open
priority: normal
tags:
  - task
blockedBy:
  - uid: "[[Task A]]"
    reltype: FINISHTOSTART
---
`)

	v := New(cfg)

	// Simulate what delete command does: clean up references
	tasks, _ := v.ListTasks()
	deletedTitle := "Task A"
	for _, other := range tasks {
		if other.Title == deletedTitle {
			continue
		}
		var newDeps []task.Dependency
		for _, dep := range other.BlockedBy {
			if dep.UID != deletedTitle {
				newDeps = append(newDeps, dep)
			}
		}
		if len(newDeps) != len(other.BlockedBy) {
			other.BlockedBy = newDeps
			v.SaveTask(other)
		}
	}
	v.DeleteTask("Tasks/task-a.md")

	// Verify Task B no longer has the blockedBy reference
	taskB, err := v.GetTask("Tasks/task-b.md")
	if err != nil {
		t.Fatalf("error reading task B: %v", err)
	}
	if len(taskB.BlockedBy) != 0 {
		t.Errorf("expected 0 dependencies, got %d", len(taskB.BlockedBy))
	}
}

func TestCompleteTask_Recurring(t *testing.T) {
	dir := t.TempDir()
	pluginDir := filepath.Join(dir, ".obsidian", "plugins", "tasknotes")
	os.MkdirAll(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, "data.json"), []byte(`{
		"tasksFolder": "Tasks",
		"archiveFolder": "Archive",
		"taskTag": "task"
	}`), 0644)
	os.MkdirAll(filepath.Join(dir, "Tasks"), 0755)

	// Create a recurring task
	os.WriteFile(filepath.Join(dir, "Tasks", "standup.md"), []byte(`---
title: Weekly standup
status: open
priority: normal
scheduled: "2026-03-23"
tags:
    - task
recurrence: "FREQ=WEEKLY;BYDAY=MO"
recurrence_anchor: scheduled
complete_instances:
    - '2026-03-16'
---
`), 0644)

	cfg, _ := config.Load(dir)
	v := New(cfg)

	tk, _ := v.GetTask("Tasks/standup.md")

	// Simulate completing the recurring task
	tk.CompleteInstances = append(tk.CompleteInstances, "2026-03-23")
	if err := v.SaveTask(tk); err != nil {
		t.Fatalf("error saving: %v", err)
	}

	// Verify
	tk2, _ := v.GetTask("Tasks/standup.md")
	if len(tk2.CompleteInstances) != 2 {
		t.Errorf("expected 2 complete instances, got %d", len(tk2.CompleteInstances))
	}
}
