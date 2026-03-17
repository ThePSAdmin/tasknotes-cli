package vault

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thepsadmin/tasknotes-cli/internal/config"
	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

func setupVault(t *testing.T) (string, *config.Config) {
	t.Helper()
	dir := t.TempDir()

	// Create plugin settings
	pluginDir := filepath.Join(dir, ".obsidian", "plugins", "tasknotes")
	os.MkdirAll(pluginDir, 0755)
	os.WriteFile(filepath.Join(pluginDir, "data.json"), []byte(`{
		"tasksFolder": "Tasks",
		"archiveFolder": "Archive",
		"taskTag": "task"
	}`), 0644)

	// Create tasks folder
	tasksDir := filepath.Join(dir, "Tasks")
	os.MkdirAll(tasksDir, 0755)
	os.MkdirAll(filepath.Join(dir, "Archive"), 0755)

	cfg, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	return dir, cfg
}

func writeTask(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, "Tasks", name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestVault_ListTasks(t *testing.T) {
	dir, cfg := setupVault(t)
	writeTask(t, dir, "task1.md", `---
title: Task One
status: open
priority: high
tags:
  - task
---
`)
	writeTask(t, dir, "task2.md", `---
title: Task Two
status: done
priority: normal
tags:
  - task
---
`)
	writeTask(t, dir, "not-a-task.md", `---
title: Not a task
---
`)

	v := New(cfg)
	tasks, err := v.ListTasks()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should only include files with the task tag
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
}

func TestVault_GetTask(t *testing.T) {
	dir, cfg := setupVault(t)
	writeTask(t, dir, "deploy.md", `---
title: Deploy auth
status: in-progress
priority: high
due: 2026-03-20
tags:
  - task
---
Deploy the changes.
`)

	v := New(cfg)
	tk, err := v.GetTask("Tasks/deploy.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tk.Title != "Deploy auth" {
		t.Errorf("title: got %q", tk.Title)
	}
	if tk.Status != "in-progress" {
		t.Errorf("status: got %q", tk.Status)
	}
}

func TestVault_GetTask_NotFound(t *testing.T) {
	_, cfg := setupVault(t)
	v := New(cfg)
	_, err := v.GetTask("Tasks/nonexistent.md")
	if err == nil {
		t.Error("expected error for nonexistent task")
	}
}

func TestVault_CreateTask(t *testing.T) {
	_, cfg := setupVault(t)
	v := New(cfg)

	tk := &task.Task{
		Title:    "Fix login redirect bug",
		Status:   "open",
		Priority: "high",
		Due:      "2026-03-20",
		Tags:     []string{"task", "bugfix"},
		Contexts: []string{"@dev"},
	}

	path, err := v.CreateTask(tk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "Tasks/fix-login-redirect-bug.md" {
		t.Errorf("expected 'Tasks/fix-login-redirect-bug.md', got %q", path)
	}

	// Verify the file exists and can be read back
	tk2, err := v.GetTask(path)
	if err != nil {
		t.Fatalf("error reading created task: %v", err)
	}
	if tk2.Title != "Fix login redirect bug" {
		t.Errorf("title: got %q", tk2.Title)
	}
	if tk2.Priority != "high" {
		t.Errorf("priority: got %q", tk2.Priority)
	}
}

func TestVault_UpdateTask(t *testing.T) {
	dir, cfg := setupVault(t)
	writeTask(t, dir, "update-me.md", `---
title: Update me
status: open
priority: normal
tags:
  - task
---
`)
	v := New(cfg)
	tk, _ := v.GetTask("Tasks/update-me.md")
	tk.Status = "done"
	tk.Priority = "high"

	err := v.SaveTask(tk)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tk2, _ := v.GetTask("Tasks/update-me.md")
	if tk2.Status != "done" {
		t.Errorf("status not updated: got %q", tk2.Status)
	}
	if tk2.Priority != "high" {
		t.Errorf("priority not updated: got %q", tk2.Priority)
	}
}

func TestVault_DeleteTask(t *testing.T) {
	dir, cfg := setupVault(t)
	writeTask(t, dir, "delete-me.md", `---
title: Delete me
status: open
priority: normal
tags:
  - task
---
`)
	v := New(cfg)
	err := v.DeleteTask("Tasks/delete-me.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = v.GetTask("Tasks/delete-me.md")
	if err == nil {
		t.Error("task should be deleted")
	}
}

func TestVault_ArchiveTask(t *testing.T) {
	dir, cfg := setupVault(t)
	writeTask(t, dir, "archive-me.md", `---
title: Archive me
status: done
priority: normal
tags:
  - task
---
`)
	v := New(cfg)
	newPath, err := v.ArchiveTask("Tasks/archive-me.md")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newPath != "Archive/archive-me.md" {
		t.Errorf("expected 'Archive/archive-me.md', got %q", newPath)
	}

	// Old path should not exist
	_, err = v.GetTask("Tasks/archive-me.md")
	if err == nil {
		t.Error("old path should not exist")
	}

	// New path should exist
	tk, err := v.GetTask("Archive/archive-me.md")
	if err != nil {
		t.Fatalf("archived task not found: %v", err)
	}
	if tk.Title != "Archive me" {
		t.Errorf("title: got %q", tk.Title)
	}
}

func TestVault_BuildBlockingMap(t *testing.T) {
	dir, cfg := setupVault(t)
	writeTask(t, dir, "parent.md", `---
title: Parent task
status: open
priority: normal
tags:
  - task
---
`)
	writeTask(t, dir, "child.md", `---
title: Child task
status: open
priority: normal
tags:
  - task
blockedBy:
  - uid: "[[Parent task]]"
    reltype: FINISHTOSTART
---
`)

	v := New(cfg)
	tasks, _ := v.ListTasks()
	v.BuildBlockingMap(tasks)

	var parent, child *task.Task
	for i := range tasks {
		if tasks[i].Title == "Parent task" {
			parent = tasks[i]
		}
		if tasks[i].Title == "Child task" {
			child = tasks[i]
		}
	}

	if parent == nil || child == nil {
		t.Fatal("tasks not found")
	}
	if len(parent.Blocking) != 1 {
		t.Errorf("parent should block 1 task, got %d", len(parent.Blocking))
	}
	if !child.IsBlocked {
		t.Error("child should be blocked")
	}
}

func TestVault_FilenameFromTitle(t *testing.T) {
	tests := []struct {
		title    string
		expected string
	}{
		{"Fix Login Redirect Bug", "fix-login-redirect-bug"},
		{"Hello World!", "hello-world"},
		{"  spaces  around  ", "spaces-around"},
		{"Special @#$ chars", "special-chars"},
		{"Multiple---dashes", "multiple-dashes"},
	}

	for _, tt := range tests {
		got := filenameFromTitle(tt.title)
		if got != tt.expected {
			t.Errorf("filenameFromTitle(%q) = %q, want %q", tt.title, got, tt.expected)
		}
	}
}
