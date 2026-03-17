package format

import (
	"bytes"
	"strings"
	"testing"

	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

func TestFormatTSV(t *testing.T) {
	tasks := []*task.Task{
		{Path: "Tasks/a.md", Title: "Task A", Status: "open", Priority: "high", Due: "2026-03-20"},
		{Path: "Tasks/b.md", Title: "Task B", Status: "done", Priority: "normal", Due: "2026-03-15"},
	}
	fields := []string{"path", "title", "status", "priority", "due"}

	var buf bytes.Buffer
	TSV(&buf, tasks, fields)
	output := buf.String()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines (header + 2 tasks), got %d", len(lines))
	}
	if lines[0] != "path\ttitle\tstatus\tpriority\tdue" {
		t.Errorf("unexpected header: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "Tasks/a.md\tTask A") {
		t.Errorf("unexpected first row: %q", lines[1])
	}
}

func TestFormatJSON(t *testing.T) {
	tasks := []*task.Task{
		{Path: "Tasks/a.md", Title: "Task A", Status: "open", Priority: "high"},
	}

	var buf bytes.Buffer
	JSON(&buf, tasks)
	output := buf.String()

	if !strings.Contains(output, `"path":"Tasks/a.md"`) {
		t.Errorf("missing path in JSON: %s", output)
	}
	if !strings.Contains(output, `"title":"Task A"`) {
		t.Errorf("missing title in JSON: %s", output)
	}
}

func TestFormatTaskDetail(t *testing.T) {
	tk := &task.Task{
		Path:     "Tasks/deploy.md",
		Title:    "Deploy auth",
		Status:   "in-progress",
		Priority: "high",
		Due:      "2026-03-20",
		Tags:     []string{"task", "backend"},
		Contexts: []string{"@dev"},
		Projects: []string{"[[Backend Rewrite]]"},
		BlockedBy: []task.Dependency{
			{UID: "Migrate database", RelType: "FINISHTOSTART"},
		},
		Body: "Deploy the changes.\n",
	}

	var buf bytes.Buffer
	TaskDetail(&buf, tk)
	output := buf.String()

	if !strings.Contains(output, "path: Tasks/deploy.md") {
		t.Error("missing path")
	}
	if !strings.Contains(output, "title: Deploy auth") {
		t.Error("missing title")
	}
	if !strings.Contains(output, "status: in-progress") {
		t.Error("missing status")
	}
	if !strings.Contains(output, "tags: task, backend") {
		t.Error("missing tags")
	}
	if !strings.Contains(output, "blockedBy: [[Migrate database]] (FINISHTOSTART)") {
		t.Errorf("missing blockedBy in:\n%s", output)
	}
	if !strings.Contains(output, "body:") {
		t.Error("missing body")
	}
}

func TestFormatGrouped(t *testing.T) {
	tasks := []*task.Task{
		{Path: "Tasks/a.md", Title: "Task A", Status: "open", Priority: "high"},
		{Path: "Tasks/b.md", Title: "Task B", Status: "open", Priority: "normal"},
		{Path: "Tasks/c.md", Title: "Task C", Status: "open", Priority: "high"},
	}
	fields := []string{"path", "title", "status"}

	var buf bytes.Buffer
	GroupedTSV(&buf, tasks, "priority", fields)
	output := buf.String()

	if !strings.Contains(output, "## high") {
		t.Error("missing high group header")
	}
	if !strings.Contains(output, "## normal") {
		t.Error("missing normal group header")
	}
}

func TestGetFieldValue(t *testing.T) {
	tk := &task.Task{
		Path:     "Tasks/a.md",
		Title:    "Test",
		Status:   "open",
		Priority: "high",
		Due:      "2026-03-20",
		Tags:     []string{"task", "bug"},
		Blocking: []string{"Tasks/b.md", "Tasks/c.md"},
	}
	tests := []struct {
		field    string
		expected string
	}{
		{"path", "Tasks/a.md"},
		{"title", "Test"},
		{"tags", "task, bug"},
		{"blocking", "Tasks/b.md, Tasks/c.md"},
	}
	for _, tt := range tests {
		got := GetFieldValue(tk, tt.field)
		if got != tt.expected {
			t.Errorf("GetFieldValue(%q) = %q, want %q", tt.field, got, tt.expected)
		}
	}
}
