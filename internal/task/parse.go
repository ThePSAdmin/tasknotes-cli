package task

import (
	"bytes"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/thepsadmin/tasknotes-cli/internal/config"
	"gopkg.in/yaml.v3"
)

// Parse parses a task markdown file into a Task struct.
// The file must have YAML frontmatter delimited by ---.
func Parse(path string, content []byte, fm config.FieldMapping) (*Task, error) {
	frontmatter, body, err := splitFrontmatter(content)
	if err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	if err := yaml.Unmarshal(frontmatter, &raw); err != nil {
		return nil, fmt.Errorf("parsing frontmatter: %w", err)
	}
	if raw == nil {
		raw = make(map[string]interface{})
	}

	t := &Task{
		Path: path,
		Body: body,
	}

	t.Title = getString(raw, fm.Title)
	if t.Title == "" {
		// Derive title from filename
		base := filepath.Base(path)
		t.Title = strings.TrimSuffix(base, filepath.Ext(base))
	}

	t.Status = getString(raw, fm.Status)
	t.Priority = getString(raw, fm.Priority)
	t.Due = getString(raw, fm.Due)
	t.Scheduled = getString(raw, fm.Scheduled)
	t.CompletedDate = getString(raw, fm.CompletedDate)
	t.DateCreated = getString(raw, fm.DateCreated)
	t.DateModified = getString(raw, fm.DateModified)
	t.Recurrence = getString(raw, fm.Recurrence)
	t.RecurrenceAnchor = getString(raw, fm.RecurrenceAnchor)
	t.TimeEstimate = getInt(raw, fm.TimeEstimate)

	t.Tags = getStringSlice(raw, "tags")
	t.Contexts = getStringSlice(raw, fm.Contexts)
	t.Projects = getStringSlice(raw, fm.Projects)
	t.CompleteInstances = getStringSlice(raw, fm.CompleteInstances)
	t.SkippedInstances = getStringSlice(raw, fm.SkippedInstances)

	t.BlockedBy = parseDependencies(raw, fm.BlockedBy)
	t.TimeEntries = parseTimeEntries(raw, fm.TimeEntries)
	t.TotalTrackedTime = computeTotalTrackedTime(t.TimeEntries)

	return t, nil
}

// splitFrontmatter splits YAML frontmatter from body content.
func splitFrontmatter(content []byte) ([]byte, string, error) {
	s := bytes.TrimLeft(content, " \t\r\n")
	if !bytes.HasPrefix(s, []byte("---")) {
		return nil, "", fmt.Errorf("no frontmatter found")
	}

	// Find the closing ---
	rest := s[3:]
	rest = bytes.TrimLeft(rest, " \t")
	if len(rest) > 0 && rest[0] == '\n' {
		rest = rest[1:]
	} else if len(rest) > 1 && rest[0] == '\r' && rest[1] == '\n' {
		rest = rest[2:]
	}

	idx := bytes.Index(rest, []byte("\n---"))
	if idx < 0 {
		// Frontmatter only, no body
		return rest, "", nil
	}

	fm := rest[:idx]
	after := rest[idx+4:] // skip \n---
	// Skip optional newline after closing ---
	after = bytes.TrimLeft(after, " \t")
	if len(after) > 0 && after[0] == '\n' {
		after = after[1:]
	} else if len(after) > 1 && after[0] == '\r' && after[1] == '\n' {
		after = after[2:]
	}

	return fm, string(after), nil
}

func getString(raw map[string]interface{}, key string) string {
	v, ok := raw[key]
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case float64:
		if val == math.Trunc(val) {
			return fmt.Sprintf("%d", int(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case time.Time:
		// YAML parser auto-parses dates; convert back to string
		if val.Hour() == 0 && val.Minute() == 0 && val.Second() == 0 && val.Nanosecond() == 0 {
			return val.Format("2006-01-02")
		}
		return val.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func getInt(raw map[string]interface{}, key string) int {
	v, ok := raw[key]
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case float64:
		return int(val)
	default:
		return 0
	}
}

func getStringSlice(raw map[string]interface{}, key string) []string {
	v, ok := raw[key]
	if !ok {
		return nil
	}
	switch val := v.(type) {
	case []interface{}:
		result := make([]string, 0, len(val))
		for _, item := range val {
			result = append(result, fmt.Sprintf("%v", item))
		}
		return result
	case string:
		return []string{val}
	default:
		return nil
	}
}

func parseDependencies(raw map[string]interface{}, key string) []Dependency {
	v, ok := raw[key]
	if !ok {
		return nil
	}
	items, ok := v.([]interface{})
	if !ok {
		return nil
	}

	deps := make([]Dependency, 0, len(items))
	for _, item := range items {
		switch val := item.(type) {
		case map[string]interface{}:
			dep := Dependency{
				UID:     stripWikilinks(getString(val, "uid")),
				RelType: getString(val, "reltype"),
			}
			if dep.RelType == "" {
				dep.RelType = "FINISHTOSTART"
			}
			if g := getString(val, "gap"); g != "" {
				dep.Gap = g
			}
			if dep.UID != "" {
				deps = append(deps, dep)
			}
		case string:
			uid := stripWikilinks(val)
			if uid != "" {
				deps = append(deps, Dependency{UID: uid, RelType: "FINISHTOSTART"})
			}
		}
	}
	return deps
}

func stripWikilinks(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "[[") && strings.HasSuffix(s, "]]") {
		return s[2 : len(s)-2]
	}
	return s
}

var timeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05Z",
	"2006-01-02T15:04:05",
	"2006-01-02T15:04:05-07:00",
	"2006-01-02 15:04:05",
}

func parseTime(s string) time.Time {
	for _, layout := range timeFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func parseTimeEntries(raw map[string]interface{}, key string) []TimeEntry {
	v, ok := raw[key]
	if !ok {
		return nil
	}
	items, ok := v.([]interface{})
	if !ok {
		return nil
	}

	entries := make([]TimeEntry, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		entry := TimeEntry{
			StartTime:   parseTime(getString(m, "startTime")),
			Description: getString(m, "description"),
		}
		if endStr := getString(m, "endTime"); endStr != "" {
			entry.EndTime = parseTime(endStr)
		}
		if !entry.StartTime.IsZero() {
			entries = append(entries, entry)
		}
	}
	return entries
}

func computeTotalTrackedTime(entries []TimeEntry) int {
	var total time.Duration
	for _, e := range entries {
		if !e.EndTime.IsZero() && !e.StartTime.IsZero() {
			total += e.EndTime.Sub(e.StartTime)
		}
	}
	return int(total.Minutes())
}
