package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thepsadmin/tasknotes-cli/internal/config"
	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

// Vault provides operations on a task vault.
type Vault struct {
	cfg *config.Config
}

// New creates a Vault instance from configuration.
func New(cfg *config.Config) *Vault {
	return &Vault{cfg: cfg}
}

// Config returns the vault configuration.
func (v *Vault) Config() *config.Config {
	return v.cfg
}

// ListTasks reads all task files from the tasks folder.
// Only files containing the configured task tag are returned.
func (v *Vault) ListTasks() ([]*task.Task, error) {
	return v.listTasksInFolder(v.cfg.TasksFolder, false)
}

// ListAllTasks reads tasks from both the tasks folder and archive folder.
func (v *Vault) ListAllTasks() ([]*task.Task, error) {
	tasks, err := v.listTasksInFolder(v.cfg.TasksFolder, false)
	if err != nil {
		return nil, err
	}
	archived, err := v.listTasksInFolder(v.cfg.ArchiveFolder, true)
	if err != nil {
		// Archive folder might not exist, that's okay
		return tasks, nil
	}
	return append(tasks, archived...), nil
}

func (v *Vault) listTasksInFolder(folder string, _ bool) ([]*task.Task, error) {
	absDir := filepath.Join(v.cfg.VaultPath, folder)
	entries, err := os.ReadDir(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var tasks []*task.Task
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		relPath := filepath.Join(folder, entry.Name())
		data, err := os.ReadFile(filepath.Join(absDir, entry.Name()))
		if err != nil {
			continue
		}
		t, err := task.Parse(relPath, data, v.cfg.FieldMapping)
		if err != nil {
			continue
		}
		// Only include files with the task tag
		if !hasTag(t.Tags, v.cfg.TaskTag) {
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func hasTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetTask reads a single task by its relative path.
func (v *Vault) GetTask(relPath string) (*task.Task, error) {
	absPath := filepath.Join(v.cfg.VaultPath, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("task not found: %s", relPath)
	}
	return task.Parse(relPath, data, v.cfg.FieldMapping)
}

// CreateTask creates a new task file and returns its relative path.
func (v *Vault) CreateTask(t *task.Task) (string, error) {
	// Ensure task has the task tag
	if !hasTag(t.Tags, v.cfg.TaskTag) {
		t.Tags = append([]string{v.cfg.TaskTag}, t.Tags...)
	}

	filename := filenameFromTitle(t.Title) + ".md"
	relPath := filepath.Join(v.cfg.TasksFolder, filename)
	absPath := filepath.Join(v.cfg.VaultPath, relPath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return "", err
	}

	// Check if file already exists
	if _, err := os.Stat(absPath); err == nil {
		return "", fmt.Errorf("task file already exists: %s", relPath)
	}

	t.Path = relPath

	data, err := task.Serialize(t, v.cfg.FieldMapping)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return "", err
	}
	return relPath, nil
}

// SaveTask writes a task back to its file.
func (v *Vault) SaveTask(t *task.Task) error {
	absPath := filepath.Join(v.cfg.VaultPath, t.Path)
	data, err := task.Serialize(t, v.cfg.FieldMapping)
	if err != nil {
		return err
	}
	return os.WriteFile(absPath, data, 0644)
}

// DeleteTask removes a task file.
func (v *Vault) DeleteTask(relPath string) error {
	absPath := filepath.Join(v.cfg.VaultPath, relPath)
	return os.Remove(absPath)
}

// ArchiveTask moves a task to the archive folder.
func (v *Vault) ArchiveTask(relPath string) (string, error) {
	filename := filepath.Base(relPath)
	newRelPath := filepath.Join(v.cfg.ArchiveFolder, filename)
	absOld := filepath.Join(v.cfg.VaultPath, relPath)
	absNew := filepath.Join(v.cfg.VaultPath, newRelPath)

	if err := os.MkdirAll(filepath.Dir(absNew), 0755); err != nil {
		return "", err
	}

	if err := os.Rename(absOld, absNew); err != nil {
		return "", err
	}
	return newRelPath, nil
}

// BuildBlockingMap populates the Blocking and IsBlocked fields on tasks
// by scanning all blockedBy references.
func (v *Vault) BuildBlockingMap(tasks []*task.Task) {
	// Build a map from title -> task for dependency resolution
	byTitle := make(map[string]*task.Task)
	byPath := make(map[string]*task.Task)
	for _, t := range tasks {
		byTitle[t.Title] = t
		byPath[t.Path] = t
	}

	for _, t := range tasks {
		if len(t.BlockedBy) > 0 {
			t.IsBlocked = true
		}
		for _, dep := range t.BlockedBy {
			// Try to resolve by title (UID is typically a title from wikilinks)
			if blocker, ok := byTitle[dep.UID]; ok {
				blocker.Blocking = append(blocker.Blocking, t.Path)
			}
		}
	}
}

var nonAlphanumeric = regexp.MustCompile(`[^a-z0-9]+`)

// filenameFromTitle converts a title to a kebab-case filename (without extension).
func filenameFromTitle(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	s = nonAlphanumeric.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	// Collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}
