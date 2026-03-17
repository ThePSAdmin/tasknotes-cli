package filter

import (
	"sort"
	"strings"
	"time"

	"github.com/thepsadmin/tasknotes-cli/internal/config"
	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

// Options specifies filtering, sorting, and limiting criteria.
type Options struct {
	Status         string
	Priority       string
	Tags           []string
	Context        string
	Project        string
	DueBefore      string
	DueAfter       string
	ScheduledBefore string
	ScheduledAfter  string
	Overdue        bool
	Blocked        bool
	Blocking       bool
	Sort           string // due, scheduled, priority, status, title, dateCreated
	SortDir        string // asc, desc
	Limit          int
	Today          string // override for testing (YYYY-MM-DD)
}

// Apply filters, sorts, and limits a list of tasks.
func Apply(tasks []*task.Task, opts *Options, cfg *config.Config) []*task.Task {
	if opts == nil {
		return tasks
	}

	result := make([]*task.Task, 0, len(tasks))
	for _, t := range tasks {
		if matchesFilters(t, opts, cfg) {
			result = append(result, t)
		}
	}

	sortTasks(result, opts, cfg)

	if opts.Limit > 0 && len(result) > opts.Limit {
		result = result[:opts.Limit]
	}

	return result
}

func matchesFilters(t *task.Task, opts *Options, cfg *config.Config) bool {
	if opts.Status != "" && t.Status != opts.Status {
		return false
	}
	if opts.Priority != "" && t.Priority != opts.Priority {
		return false
	}
	if len(opts.Tags) > 0 {
		for _, tag := range opts.Tags {
			if !containsStr(t.Tags, tag) {
				return false
			}
		}
	}
	if opts.Context != "" && !containsStr(t.Contexts, opts.Context) {
		return false
	}
	if opts.Project != "" && !containsStr(t.Projects, opts.Project) {
		return false
	}
	if opts.DueBefore != "" && (t.Due == "" || t.Due >= opts.DueBefore) {
		return false
	}
	if opts.DueAfter != "" && (t.Due == "" || t.Due <= opts.DueAfter) {
		return false
	}
	if opts.ScheduledBefore != "" && (t.Scheduled == "" || t.Scheduled >= opts.ScheduledBefore) {
		return false
	}
	if opts.ScheduledAfter != "" && (t.Scheduled == "" || t.Scheduled <= opts.ScheduledAfter) {
		return false
	}
	if opts.Overdue {
		today := opts.Today
		if today == "" {
			today = time.Now().Format("2006-01-02")
		}
		if t.Due == "" || t.Due >= today || cfg.IsCompletedStatus(t.Status) {
			return false
		}
	}
	if opts.Blocked && !t.IsBlocked {
		return false
	}
	if opts.Blocking && len(t.Blocking) == 0 {
		return false
	}
	return true
}

func sortTasks(tasks []*task.Task, opts *Options, cfg *config.Config) {
	sortKey := opts.Sort
	if sortKey == "" {
		sortKey = "due"
	}
	asc := true
	if strings.ToLower(opts.SortDir) == "desc" {
		asc = false
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		cmp := compareTasks(tasks[i], tasks[j], sortKey, cfg)
		if asc {
			return cmp < 0
		}
		return cmp > 0
	})
}

func compareTasks(a, b *task.Task, key string, cfg *config.Config) int {
	switch key {
	case "due":
		return compareDateStrings(a.Due, b.Due)
	case "scheduled":
		return compareDateStrings(a.Scheduled, b.Scheduled)
	case "priority":
		wa := cfg.PriorityWeight(a.Priority)
		wb := cfg.PriorityWeight(b.Priority)
		return wa - wb
	case "status":
		return strings.Compare(a.Status, b.Status)
	case "title":
		return strings.Compare(a.Title, b.Title)
	case "dateCreated":
		return compareDateStrings(a.DateCreated, b.DateCreated)
	default:
		return 0
	}
}

// compareDateStrings compares two date strings. Empty strings sort last.
func compareDateStrings(a, b string) int {
	if a == "" && b == "" {
		return 0
	}
	if a == "" {
		return 1 // empty sorts last
	}
	if b == "" {
		return -1
	}
	return strings.Compare(a, b)
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
