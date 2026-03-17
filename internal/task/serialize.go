package task

import (
	"bytes"
	"fmt"
	"time"

	"github.com/thepsadmin/tasknotes-cli/internal/config"
	"gopkg.in/yaml.v3"
)

// Serialize converts a Task into markdown with YAML frontmatter.
func Serialize(t *Task, fm config.FieldMapping) ([]byte, error) {
	// Build ordered frontmatter map
	m := yaml.Node{Kind: yaml.MappingNode}

	addString(&m, fm.Title, t.Title)
	addString(&m, fm.Status, t.Status)
	addString(&m, fm.Priority, t.Priority)

	if t.Due != "" {
		addQuotedString(&m, fm.Due, t.Due)
	}
	if t.Scheduled != "" {
		addQuotedString(&m, fm.Scheduled, t.Scheduled)
	}
	if t.CompletedDate != "" {
		addQuotedString(&m, fm.CompletedDate, t.CompletedDate)
	}

	if len(t.Tags) > 0 {
		addStringSlice(&m, "tags", t.Tags)
	}
	if len(t.Contexts) > 0 {
		addStringSlice(&m, fm.Contexts, t.Contexts)
	}
	if len(t.Projects) > 0 {
		addStringSlice(&m, fm.Projects, t.Projects)
	}

	if t.TimeEstimate > 0 {
		addInt(&m, fm.TimeEstimate, t.TimeEstimate)
	}

	if t.DateCreated != "" {
		addQuotedString(&m, fm.DateCreated, t.DateCreated)
	}
	if t.DateModified != "" {
		addQuotedString(&m, fm.DateModified, t.DateModified)
	}

	if t.Recurrence != "" {
		addQuotedString(&m, fm.Recurrence, t.Recurrence)
	}
	if t.RecurrenceAnchor != "" {
		addString(&m, fm.RecurrenceAnchor, t.RecurrenceAnchor)
	}
	if len(t.CompleteInstances) > 0 {
		addStringSlice(&m, fm.CompleteInstances, t.CompleteInstances)
	}
	if len(t.SkippedInstances) > 0 {
		addStringSlice(&m, fm.SkippedInstances, t.SkippedInstances)
	}

	if len(t.BlockedBy) > 0 {
		addDependencies(&m, fm.BlockedBy, t.BlockedBy)
	}

	if len(t.TimeEntries) > 0 {
		addTimeEntries(&m, fm.TimeEntries, t.TimeEntries)
	}

	var buf bytes.Buffer
	buf.WriteString("---\n")

	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(4)
	if err := enc.Encode(&m); err != nil {
		return nil, fmt.Errorf("encoding frontmatter: %w", err)
	}
	enc.Close()

	buf.WriteString("---\n")

	if t.Body != "" {
		buf.WriteString(t.Body)
	}

	return buf.Bytes(), nil
}

func addString(m *yaml.Node, key, value string) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value},
	)
}

func addQuotedString(m *yaml.Node, key, value string) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: value, Style: yaml.DoubleQuotedStyle},
	)
}

func addInt(m *yaml.Node, key string, value int) {
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		&yaml.Node{Kind: yaml.ScalarNode, Value: fmt.Sprintf("%d", value), Tag: "!!int"},
	)
}

func addStringSlice(m *yaml.Node, key string, values []string) {
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, v := range values {
		seq.Content = append(seq.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: v, Style: yaml.SingleQuotedStyle})
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		seq,
	)
}

func addDependencies(m *yaml.Node, key string, deps []Dependency) {
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, dep := range deps {
		entry := &yaml.Node{Kind: yaml.MappingNode}
		uid := dep.UID
		if uid != "" && !isWikilink(uid) {
			uid = "[[" + uid + "]]"
		}
		entry.Content = append(entry.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "uid"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: uid, Style: yaml.SingleQuotedStyle},
		)
		entry.Content = append(entry.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "reltype"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: dep.RelType},
		)
		if dep.Gap != "" {
			entry.Content = append(entry.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "gap"},
				&yaml.Node{Kind: yaml.ScalarNode, Value: dep.Gap},
			)
		}
		seq.Content = append(seq.Content, entry)
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		seq,
	)
}

func isWikilink(s string) bool {
	return len(s) > 4 && s[:2] == "[[" && s[len(s)-2:] == "]]"
}

func addTimeEntries(m *yaml.Node, key string, entries []TimeEntry) {
	seq := &yaml.Node{Kind: yaml.SequenceNode}
	for _, e := range entries {
		entry := &yaml.Node{Kind: yaml.MappingNode}
		entry.Content = append(entry.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "startTime"},
			&yaml.Node{Kind: yaml.ScalarNode, Value: e.StartTime.Format(time.RFC3339), Style: yaml.DoubleQuotedStyle},
		)
		if !e.EndTime.IsZero() {
			entry.Content = append(entry.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "endTime"},
				&yaml.Node{Kind: yaml.ScalarNode, Value: e.EndTime.Format(time.RFC3339), Style: yaml.DoubleQuotedStyle},
			)
		}
		if e.Description != "" {
			entry.Content = append(entry.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: "description"},
				&yaml.Node{Kind: yaml.ScalarNode, Value: e.Description},
			)
		}
		seq.Content = append(seq.Content, entry)
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key},
		seq,
	)
}
