package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestVault(t *testing.T) string {
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

	// Create folders
	os.MkdirAll(filepath.Join(dir, "Tasks"), 0755)
	os.MkdirAll(filepath.Join(dir, "Archive"), 0755)

	// Create sample tasks
	writeTestTask(t, dir, "Tasks/deploy-auth.md", `---
title: Deploy authentication update
status: in-progress
priority: high
due: "2026-03-20"
scheduled: "2026-03-19"
tags:
    - task
    - deployment
    - backend
contexts:
    - '@dev'
projects:
    - '[[Backend Rewrite]]'
timeEstimate: 60
dateCreated: "2026-03-10T09:00:00Z"
blockedBy:
    - uid: '[[Migrate database schema]]'
      reltype: FINISHTOSTART
---
Deploy the reviewed authentication changes to production.
`)
	writeTestTask(t, dir, "Tasks/migrate-db.md", `---
title: Migrate database schema
status: in-progress
priority: high
due: "2026-03-18"
tags:
    - task
    - backend
dateCreated: "2026-03-08T09:00:00Z"
---
`)
	writeTestTask(t, dir, "Tasks/write-tests.md", `---
title: Write integration tests
status: open
priority: normal
due: "2026-03-25"
tags:
    - task
dateCreated: "2026-03-12T09:00:00Z"
---
`)
	writeTestTask(t, dir, "Tasks/update-docs.md", `---
title: Update API documentation
status: open
priority: low
tags:
    - task
dateCreated: "2026-03-14T09:00:00Z"
---
`)
	writeTestTask(t, dir, "Tasks/weekly-standup.md", `---
title: Weekly standup notes
status: open
priority: normal
scheduled: "2026-03-23"
tags:
    - task
recurrence: "FREQ=WEEKLY;BYDAY=MO"
recurrence_anchor: scheduled
complete_instances:
    - '2026-03-16'
    - '2026-03-09'
skipped_instances:
    - '2026-03-02'
dateCreated: "2026-01-01T09:00:00Z"
---
`)

	return dir
}

func writeTestTask(t *testing.T, dir, relPath, content string) {
	t.Helper()
	absPath := filepath.Join(dir, relPath)
	os.MkdirAll(filepath.Dir(absPath), 0755)
	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func executeCmd(t *testing.T, args ...string) (string, error) {
	t.Helper()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	return buf.String(), err
}

func TestListCmd(t *testing.T) {
	dir := setupTestVault(t)

	// Reset flags for fresh test
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"list", "--vault", dir})
	err := rootCmd.Execute()

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old

	if err != nil {
		t.Fatalf("list error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Deploy authentication update") {
		t.Errorf("missing deploy task in output:\n%s", output)
	}
	if !strings.Contains(output, "Write integration tests") {
		t.Errorf("missing write-tests in output:\n%s", output)
	}
}

func TestListCmd_FilterStatus(t *testing.T) {
	dir := setupTestVault(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"list", "--vault", dir, "--status", "open"})
	err := rootCmd.Execute()

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old

	if err != nil {
		t.Fatalf("list error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "Deploy authentication update") {
		t.Error("in-progress task should not appear with --status open")
	}
	if !strings.Contains(output, "Write integration tests") {
		t.Error("open task should appear")
	}
}

func TestListCmd_JSON(t *testing.T) {
	dir := setupTestVault(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"list", "--vault", dir, "--format", "json", "--limit", "1"})
	formatFlag = "json"
	err := rootCmd.Execute()
	formatFlag = "text"

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old

	if err != nil {
		t.Fatalf("list error: %v", err)
	}

	output := buf.String()
	if !strings.HasPrefix(output, "[") {
		t.Errorf("expected JSON array, got: %s", output[:50])
	}
}

func TestGetCmd(t *testing.T) {
	dir := setupTestVault(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"get", "--vault", dir, "Tasks/deploy-auth.md"})
	err := rootCmd.Execute()

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old

	if err != nil {
		t.Fatalf("get error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "path: Tasks/deploy-auth.md") {
		t.Error("missing path")
	}
	if !strings.Contains(output, "title: Deploy authentication update") {
		t.Error("missing title")
	}
	if !strings.Contains(output, "status: in-progress") {
		t.Error("missing status")
	}
}

func TestCreateCmd(t *testing.T) {
	dir := setupTestVault(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"create", "--vault", dir, "--title", "New test task", "--priority", "high", "--due", "2026-04-01"})
	err := rootCmd.Execute()

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old

	if err != nil {
		t.Fatalf("create error: %v", err)
	}

	output := strings.TrimSpace(buf.String())
	if output != "Tasks/new-test-task.md" {
		t.Errorf("expected 'Tasks/new-test-task.md', got %q", output)
	}

	// Verify file exists
	content, err := os.ReadFile(filepath.Join(dir, "Tasks/new-test-task.md"))
	if err != nil {
		t.Fatalf("created file not found: %v", err)
	}
	if !strings.Contains(string(content), "New test task") {
		t.Error("task title not in file")
	}
}

func TestRecurrenceCmd(t *testing.T) {
	dir := setupTestVault(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"recurrence", "--vault", dir, "Tasks/weekly-standup.md"})
	err := rootCmd.Execute()

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old

	if err != nil {
		t.Fatalf("recurrence error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "rule: FREQ=WEEKLY;BYDAY=MO") {
		t.Errorf("missing rule in output:\n%s", output)
	}
	if !strings.Contains(output, "completed: 2") {
		t.Errorf("missing completed count in output:\n%s", output)
	}
	if !strings.Contains(output, "skipped: 1") {
		t.Errorf("missing skipped count in output:\n%s", output)
	}
}

func TestDepsCmd_Roots(t *testing.T) {
	dir := setupTestVault(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"deps", "--vault", dir, "--roots"})
	err := rootCmd.Execute()

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old

	if err != nil {
		t.Fatalf("deps error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Migrate database schema") {
		t.Errorf("expected migrate-db as root task, got:\n%s", output)
	}
}

func TestInfoCmd(t *testing.T) {
	dir := setupTestVault(t)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.SetArgs([]string{"info", "--vault", dir})
	err := rootCmd.Execute()

	w.Close()
	var buf bytes.Buffer
	buf.ReadFrom(r)
	os.Stdout = old

	if err != nil {
		t.Fatalf("info error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "tasks_folder: Tasks") {
		t.Errorf("missing tasks folder in output:\n%s", output)
	}
	if !strings.Contains(output, "task_count: 5") {
		t.Errorf("unexpected task count in output:\n%s", output)
	}
}
