package filter

import (
	"testing"

	"github.com/thepsadmin/tasknotes-cli/internal/config"
	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

func cfg() *config.Config {
	return &config.Config{
		Priorities: config.DefaultPriorities(),
		Statuses:   config.DefaultStatuses(),
	}
}

func makeTasks() []*task.Task {
	return []*task.Task{
		{Path: "Tasks/a.md", Title: "Task A", Status: "open", Priority: "high", Due: "2026-03-20", Tags: []string{"task", "backend"}, Contexts: []string{"@dev"}, Projects: []string{"[[Backend Rewrite]]"}},
		{Path: "Tasks/b.md", Title: "Task B", Status: "done", Priority: "normal", Due: "2026-03-15", Tags: []string{"task"}},
		{Path: "Tasks/c.md", Title: "Task C", Status: "in-progress", Priority: "low", Due: "2026-03-25", Tags: []string{"task", "frontend"}, Contexts: []string{"@home"}},
		{Path: "Tasks/d.md", Title: "Task D", Status: "open", Priority: "normal", Tags: []string{"task"}, IsBlocked: true},
		{Path: "Tasks/e.md", Title: "Task E", Status: "open", Priority: "high", Due: "2026-03-10", Tags: []string{"task"}, Blocking: []string{"Tasks/d.md"}},
	}
}

func TestFilterByStatus(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Status: "open"}
	result := Apply(tasks, opts, cfg())
	if len(result) != 3 {
		t.Errorf("expected 3 open tasks, got %d", len(result))
	}
}

func TestFilterByPriority(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Priority: "high"}
	result := Apply(tasks, opts, cfg())
	if len(result) != 2 {
		t.Errorf("expected 2 high-priority tasks, got %d", len(result))
	}
}

func TestFilterByTag(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Tags: []string{"backend"}}
	result := Apply(tasks, opts, cfg())
	if len(result) != 1 {
		t.Errorf("expected 1 task with backend tag, got %d", len(result))
	}
}

func TestFilterByContext(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Context: "@dev"}
	result := Apply(tasks, opts, cfg())
	if len(result) != 1 {
		t.Errorf("expected 1 task with @dev context, got %d", len(result))
	}
}

func TestFilterByProject(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Project: "[[Backend Rewrite]]"}
	result := Apply(tasks, opts, cfg())
	if len(result) != 1 {
		t.Errorf("expected 1 task in Backend Rewrite, got %d", len(result))
	}
}

func TestFilterOverdue(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Overdue: true, Today: "2026-03-17"}
	result := Apply(tasks, opts, cfg())
	// Task B (due 2026-03-15) is done, so not overdue.
	// Task E (due 2026-03-10) is open, so overdue.
	if len(result) != 1 {
		t.Errorf("expected 1 overdue task, got %d", len(result))
		for _, tk := range result {
			t.Logf("  %s status=%s due=%s", tk.Title, tk.Status, tk.Due)
		}
	}
}

func TestFilterBlocked(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Blocked: true}
	result := Apply(tasks, opts, cfg())
	if len(result) != 1 || result[0].Title != "Task D" {
		t.Errorf("expected Task D as blocked, got %v", result)
	}
}

func TestFilterBlocking(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Blocking: true}
	result := Apply(tasks, opts, cfg())
	if len(result) != 1 || result[0].Title != "Task E" {
		t.Errorf("expected Task E as blocking, got %v", result)
	}
}

func TestFilterDueBefore(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{DueBefore: "2026-03-18"}
	result := Apply(tasks, opts, cfg())
	// B (2026-03-15) and E (2026-03-10) are before 2026-03-18
	if len(result) != 2 {
		t.Errorf("expected 2 tasks due before 2026-03-18, got %d", len(result))
	}
}

func TestFilterDueAfter(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{DueAfter: "2026-03-19"}
	result := Apply(tasks, opts, cfg())
	// A (2026-03-20) and C (2026-03-25) are after 2026-03-19
	if len(result) != 2 {
		t.Errorf("expected 2 tasks due after 2026-03-19, got %d", len(result))
	}
}

func TestSortByPriorityDesc(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Sort: "priority", SortDir: "desc"}
	result := Apply(tasks, opts, cfg())
	if result[0].Priority != "high" {
		t.Errorf("expected first task to be high priority, got %q", result[0].Priority)
	}
}

func TestSortByDueAsc(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Sort: "due", SortDir: "asc"}
	result := Apply(tasks, opts, cfg())
	// Tasks without due should be last
	if result[0].Due != "2026-03-10" {
		t.Errorf("expected first due to be 2026-03-10, got %q", result[0].Due)
	}
}

func TestLimit(t *testing.T) {
	tasks := makeTasks()
	opts := &Options{Limit: 2}
	result := Apply(tasks, opts, cfg())
	if len(result) != 2 {
		t.Errorf("expected 2 tasks with limit, got %d", len(result))
	}
}
