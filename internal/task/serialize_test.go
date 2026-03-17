package task

import (
	"strings"
	"testing"
	"time"
)

func TestSerialize_BasicFields(t *testing.T) {
	task := &Task{
		Path:     "Tasks/test.md",
		Title:    "Test task",
		Status:   "open",
		Priority: "high",
		Due:      "2026-03-20",
		Tags:     []string{"task", "bugfix"},
		Contexts: []string{"@dev"},
		Body:     "Some body content.\n",
	}

	data, err := Serialize(task, defaultFieldMapping())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "title: Test task") {
		t.Error("missing title")
	}
	if !strings.Contains(content, "status: open") {
		t.Error("missing status")
	}
	if !strings.Contains(content, "priority: high") {
		t.Error("missing priority")
	}
	if !strings.Contains(content, "due: \"2026-03-20\"") && !strings.Contains(content, "due: 2026-03-20") {
		t.Error("missing due")
	}
	if !strings.Contains(content, "Some body content.") {
		t.Error("missing body")
	}
	if !strings.HasPrefix(content, "---\n") {
		t.Error("should start with ---")
	}
	if !strings.Contains(content, "\n---\n") {
		t.Error("should have closing ---")
	}
}

func TestSerialize_Dependencies(t *testing.T) {
	task := &Task{
		Path:     "Tasks/test.md",
		Title:    "Test",
		Status:   "open",
		Priority: "normal",
		Tags:     []string{"task"},
		BlockedBy: []Dependency{
			{UID: "Migrate database", RelType: "FINISHTOSTART"},
			{UID: "Code review", RelType: "STARTTOSTART", Gap: "P2D"},
		},
	}

	data, err := Serialize(task, defaultFieldMapping())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "[[Migrate database]]") {
		t.Error("dependency should be wrapped in wikilinks")
	}
	if !strings.Contains(content, "FINISHTOSTART") {
		t.Error("missing reltype")
	}
	if !strings.Contains(content, "P2D") {
		t.Error("missing gap")
	}
}

func TestSerialize_TimeEntries(t *testing.T) {
	start := time.Date(2026, 3, 16, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 16, 10, 45, 0, 0, time.UTC)
	task := &Task{
		Path:     "Tasks/test.md",
		Title:    "Test",
		Status:   "open",
		Priority: "normal",
		Tags:     []string{"task"},
		TimeEntries: []TimeEntry{
			{StartTime: start, EndTime: end, Description: "Initial setup"},
		},
	}

	data, err := Serialize(task, defaultFieldMapping())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "2026-03-16T10:00:00Z") {
		t.Errorf("missing startTime in:\n%s", content)
	}
	if !strings.Contains(content, "Initial setup") {
		t.Error("missing description")
	}
}

func TestSerialize_Roundtrip(t *testing.T) {
	original := `---
title: Roundtrip task
status: in-progress
priority: high
due: "2026-03-20"
scheduled: "2026-03-19"
tags:
    - task
    - backend
contexts:
    - '@dev'
projects:
    - '[[Backend Rewrite]]'
timeEstimate: 60
blockedBy:
    - uid: '[[Migrate database]]'
      reltype: FINISHTOSTART
---
Body content here.
`
	fm := defaultFieldMapping()
	task, err := Parse("Tasks/roundtrip.md", []byte(original), fm)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	data, err := Serialize(task, fm)
	if err != nil {
		t.Fatalf("serialize error: %v", err)
	}

	// Re-parse the serialized output
	task2, err := Parse("Tasks/roundtrip.md", data, fm)
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}

	if task.Title != task2.Title {
		t.Errorf("title mismatch: %q vs %q", task.Title, task2.Title)
	}
	if task.Status != task2.Status {
		t.Errorf("status mismatch: %q vs %q", task.Status, task2.Status)
	}
	if task.Due != task2.Due {
		t.Errorf("due mismatch: %q vs %q", task.Due, task2.Due)
	}
	if len(task.BlockedBy) != len(task2.BlockedBy) {
		t.Errorf("dependencies mismatch: %d vs %d", len(task.BlockedBy), len(task2.BlockedBy))
	}
	if task.Body != task2.Body {
		t.Errorf("body mismatch: %q vs %q", task.Body, task2.Body)
	}
}
