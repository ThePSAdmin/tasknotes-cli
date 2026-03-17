package task

import (
	"testing"

	"github.com/thepsadmin/tasknotes-cli/internal/config"
)

func defaultFieldMapping() config.FieldMapping {
	return config.DefaultFieldMapping()
}

func TestParseTask_BasicFields(t *testing.T) {
	content := `---
title: Deploy authentication update
status: in-progress
priority: high
due: 2026-03-20
scheduled: 2026-03-19
tags:
  - task
  - deployment
  - backend
contexts:
  - "@dev"
projects:
  - "[[Backend Rewrite]]"
timeEstimate: 60
dateCreated: 2026-03-10T09:00:00Z
dateModified: 2026-03-17T14:22:00Z
---
Deploy the reviewed authentication changes to production.
`
	task, err := Parse("Tasks/deploy-auth.md", []byte(content), defaultFieldMapping())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "Deploy authentication update" {
		t.Errorf("title: got %q", task.Title)
	}
	if task.Status != "in-progress" {
		t.Errorf("status: got %q", task.Status)
	}
	if task.Priority != "high" {
		t.Errorf("priority: got %q", task.Priority)
	}
	if task.Due != "2026-03-20" {
		t.Errorf("due: got %q", task.Due)
	}
	if task.Scheduled != "2026-03-19" {
		t.Errorf("scheduled: got %q", task.Scheduled)
	}
	if task.TimeEstimate != 60 {
		t.Errorf("timeEstimate: got %d", task.TimeEstimate)
	}
	if task.Path != "Tasks/deploy-auth.md" {
		t.Errorf("path: got %q", task.Path)
	}
	if len(task.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d: %v", len(task.Tags), task.Tags)
	}
	if len(task.Contexts) != 1 || task.Contexts[0] != "@dev" {
		t.Errorf("contexts: got %v", task.Contexts)
	}
	if len(task.Projects) != 1 || task.Projects[0] != "[[Backend Rewrite]]" {
		t.Errorf("projects: got %v", task.Projects)
	}
	expected := "Deploy the reviewed authentication changes to production.\n"
	if task.Body != expected {
		t.Errorf("body: got %q, want %q", task.Body, expected)
	}
}

func TestParseTask_Dependencies(t *testing.T) {
	content := `---
title: Update API endpoints
status: open
priority: normal
tags:
  - task
blockedBy:
  - uid: "[[Migrate database schema]]"
    reltype: FINISHTOSTART
  - uid: "[[Code review]]"
    reltype: STARTTOSTART
    gap: P2D
---
`
	task, err := Parse("Tasks/update-api.md", []byte(content), defaultFieldMapping())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(task.BlockedBy) != 2 {
		t.Fatalf("expected 2 dependencies, got %d", len(task.BlockedBy))
	}
	dep := task.BlockedBy[0]
	if dep.UID != "Migrate database schema" {
		t.Errorf("dep[0] uid: got %q", dep.UID)
	}
	if dep.RelType != "FINISHTOSTART" {
		t.Errorf("dep[0] reltype: got %q", dep.RelType)
	}
	dep2 := task.BlockedBy[1]
	if dep2.UID != "Code review" {
		t.Errorf("dep[1] uid: got %q", dep2.UID)
	}
	if dep2.Gap != "P2D" {
		t.Errorf("dep[1] gap: got %q", dep2.Gap)
	}
}

func TestParseTask_TimeEntries(t *testing.T) {
	content := `---
title: Test task
status: open
priority: normal
tags:
  - task
timeEntries:
  - startTime: "2026-03-16T10:00:00Z"
    endTime: "2026-03-16T10:45:00Z"
    description: Initial setup
  - startTime: "2026-03-17T14:30:00Z"
    endTime: "2026-03-17T14:55:00Z"
    description: Configured staging env
---
`
	task, err := Parse("Tasks/test.md", []byte(content), defaultFieldMapping())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(task.TimeEntries) != 2 {
		t.Fatalf("expected 2 time entries, got %d", len(task.TimeEntries))
	}
	if task.TimeEntries[0].Description != "Initial setup" {
		t.Errorf("entry[0] description: got %q", task.TimeEntries[0].Description)
	}
	if task.TotalTrackedTime != 70 {
		t.Errorf("total tracked time: got %d, want 70", task.TotalTrackedTime)
	}
}

func TestParseTask_Recurrence(t *testing.T) {
	content := `---
title: Weekly standup
status: open
priority: normal
tags:
  - task
recurrence: "FREQ=WEEKLY;BYDAY=MO"
recurrence_anchor: scheduled
scheduled: 2026-03-23
complete_instances:
  - "2026-03-16"
  - "2026-03-09"
skipped_instances:
  - "2026-03-02"
---
`
	task, err := Parse("Tasks/standup.md", []byte(content), defaultFieldMapping())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Recurrence != "FREQ=WEEKLY;BYDAY=MO" {
		t.Errorf("recurrence: got %q", task.Recurrence)
	}
	if task.RecurrenceAnchor != "scheduled" {
		t.Errorf("recurrence anchor: got %q", task.RecurrenceAnchor)
	}
	if len(task.CompleteInstances) != 2 {
		t.Errorf("expected 2 complete instances, got %d", len(task.CompleteInstances))
	}
	if len(task.SkippedInstances) != 1 {
		t.Errorf("expected 1 skipped instance, got %d", len(task.SkippedInstances))
	}
}

func TestParseTask_TitleFromFilename(t *testing.T) {
	content := `---
status: open
priority: normal
tags:
  - task
---
`
	task, err := Parse("Tasks/My Great Task.md", []byte(content), defaultFieldMapping())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "My Great Task" {
		t.Errorf("title from filename: got %q", task.Title)
	}
}

func TestParseTask_CustomFieldMapping(t *testing.T) {
	fm := defaultFieldMapping()
	fm.Title = "name"
	fm.Status = "state"
	fm.Priority = "urgency"

	content := `---
name: Custom task
state: open
urgency: high
tags:
  - task
---
`
	task, err := Parse("Tasks/custom.md", []byte(content), fm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "Custom task" {
		t.Errorf("title: got %q", task.Title)
	}
	if task.Status != "open" {
		t.Errorf("status: got %q", task.Status)
	}
	if task.Priority != "high" {
		t.Errorf("priority: got %q", task.Priority)
	}
}

func TestParseTask_EmptyContent(t *testing.T) {
	content := `---
---
`
	task, err := Parse("Tasks/empty.md", []byte(content), defaultFieldMapping())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Title != "empty" {
		t.Errorf("expected title from filename, got %q", task.Title)
	}
}

func TestParseTask_NoFrontmatter(t *testing.T) {
	content := `Just some text without frontmatter.`
	_, err := Parse("Tasks/nofm.md", []byte(content), defaultFieldMapping())
	if err == nil {
		t.Error("expected error for missing frontmatter")
	}
}
