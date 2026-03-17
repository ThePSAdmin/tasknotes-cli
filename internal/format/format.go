package format

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

// TSV writes tasks as tab-separated values with a header row.
func TSV(w io.Writer, tasks []*task.Task, fields []string) {
	fmt.Fprintln(w, strings.Join(fields, "\t"))
	for _, t := range tasks {
		vals := make([]string, len(fields))
		for i, f := range fields {
			vals[i] = GetFieldValue(t, f)
		}
		fmt.Fprintln(w, strings.Join(vals, "\t"))
	}
}

// JSON writes tasks as a JSON array.
func JSON(w io.Writer, tasks []*task.Task) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(tasks)
}

// JSONTask writes a single task as JSON.
func JSONTask(w io.Writer, t *task.Task) {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(t)
}

// TaskDetail writes a single task in the key: value detail format.
func TaskDetail(w io.Writer, t *task.Task) {
	fmt.Fprintf(w, "path: %s\n", t.Path)
	fmt.Fprintf(w, "title: %s\n", t.Title)
	fmt.Fprintf(w, "status: %s\n", t.Status)
	fmt.Fprintf(w, "priority: %s\n", t.Priority)
	if t.Due != "" {
		fmt.Fprintf(w, "due: %s\n", t.Due)
	}
	if t.Scheduled != "" {
		fmt.Fprintf(w, "scheduled: %s\n", t.Scheduled)
	}
	if len(t.Tags) > 0 {
		fmt.Fprintf(w, "tags: %s\n", strings.Join(t.Tags, ", "))
	}
	if len(t.Contexts) > 0 {
		fmt.Fprintf(w, "contexts: %s\n", strings.Join(t.Contexts, ", "))
	}
	if len(t.Projects) > 0 {
		fmt.Fprintf(w, "projects: %s\n", strings.Join(t.Projects, ", "))
	}
	if t.TimeEstimate > 0 {
		fmt.Fprintf(w, "timeEstimate: %d\n", t.TimeEstimate)
	}
	if t.TotalTrackedTime > 0 {
		fmt.Fprintf(w, "totalTrackedTime: %d\n", t.TotalTrackedTime)
	}
	if t.DateCreated != "" {
		fmt.Fprintf(w, "dateCreated: %s\n", t.DateCreated)
	}
	if t.DateModified != "" {
		fmt.Fprintf(w, "dateModified: %s\n", t.DateModified)
	}
	if t.CompletedDate != "" {
		fmt.Fprintf(w, "completedDate: %s\n", t.CompletedDate)
	}
	if t.Recurrence != "" {
		fmt.Fprintf(w, "recurrence: %s\n", t.Recurrence)
	}
	if t.RecurrenceAnchor != "" {
		fmt.Fprintf(w, "recurrenceAnchor: %s\n", t.RecurrenceAnchor)
	}
	for _, dep := range t.BlockedBy {
		fmt.Fprintf(w, "blockedBy: [[%s]] (%s)\n", dep.UID, dep.RelType)
	}
	if len(t.Blocking) > 0 {
		fmt.Fprintf(w, "blocking: %s\n", strings.Join(t.Blocking, ", "))
	}
	if t.Body != "" {
		fmt.Fprintf(w, "body:\n")
		for _, line := range strings.Split(t.Body, "\n") {
			if line != "" {
				fmt.Fprintf(w, "  %s\n", line)
			}
		}
	}
}

// GroupedTSV writes tasks grouped by a key, each group with a ## header.
func GroupedTSV(w io.Writer, tasks []*task.Task, groupKey string, fields []string) {
	groups := make(map[string][]*task.Task)
	var order []string
	for _, t := range tasks {
		key := GetFieldValue(t, groupKey)
		if _, exists := groups[key]; !exists {
			order = append(order, key)
		}
		groups[key] = append(groups[key], t)
	}

	for i, key := range order {
		if i > 0 {
			fmt.Fprintln(w)
		}
		fmt.Fprintf(w, "## %s\n", key)
		TSV(w, groups[key], fields)
	}
}

// GetFieldValue returns a string value for a task field.
func GetFieldValue(t *task.Task, field string) string {
	switch field {
	case "path":
		return t.Path
	case "title":
		return t.Title
	case "status":
		return t.Status
	case "priority":
		return t.Priority
	case "due":
		return t.Due
	case "scheduled":
		return t.Scheduled
	case "tags":
		return strings.Join(t.Tags, ", ")
	case "contexts":
		return strings.Join(t.Contexts, ", ")
	case "projects":
		return strings.Join(t.Projects, ", ")
	case "timeEstimate":
		if t.TimeEstimate > 0 {
			return fmt.Sprintf("%d", t.TimeEstimate)
		}
		return ""
	case "totalTrackedTime":
		if t.TotalTrackedTime > 0 {
			return fmt.Sprintf("%d", t.TotalTrackedTime)
		}
		return ""
	case "dateCreated":
		return t.DateCreated
	case "dateModified":
		return t.DateModified
	case "completedDate":
		return t.CompletedDate
	case "recurrence":
		return t.Recurrence
	case "blockedBy":
		parts := make([]string, len(t.BlockedBy))
		for i, dep := range t.BlockedBy {
			parts[i] = fmt.Sprintf("[[%s]]", dep.UID)
		}
		return strings.Join(parts, ", ")
	case "blocking":
		return strings.Join(t.Blocking, ", ")
	case "blocking_count":
		return fmt.Sprintf("%d", len(t.Blocking))
	default:
		return ""
	}
}
