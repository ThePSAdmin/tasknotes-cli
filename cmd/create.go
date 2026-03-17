package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new task",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		title, _ := cmd.Flags().GetString("title")
		if title == "" {
			return fmt.Errorf("--title is required")
		}

		tk := &task.Task{
			Title:       title,
			DateCreated: time.Now().UTC().Format(time.RFC3339),
		}

		tk.Status, _ = cmd.Flags().GetString("status")
		if tk.Status == "" {
			tk.Status = "open"
		}
		tk.Priority, _ = cmd.Flags().GetString("priority")
		if tk.Priority == "" {
			tk.Priority = "normal"
		}
		tk.Due, _ = cmd.Flags().GetString("due")
		tk.Scheduled, _ = cmd.Flags().GetString("scheduled")

		tags, _ := cmd.Flags().GetStringSlice("tag")
		tk.Tags = tags
		contexts, _ := cmd.Flags().GetStringSlice("context")
		tk.Contexts = contexts
		projects, _ := cmd.Flags().GetStringSlice("project")
		tk.Projects = projects

		timeEst, _ := cmd.Flags().GetInt("time-estimate")
		tk.TimeEstimate = timeEst

		recurrence, _ := cmd.Flags().GetString("recurrence")
		tk.Recurrence = recurrence

		body, _ := cmd.Flags().GetString("body")
		bodyFile, _ := cmd.Flags().GetString("body-file")
		if bodyFile != "" {
			data, err := os.ReadFile(bodyFile)
			if err != nil {
				return fmt.Errorf("reading body file: %w", err)
			}
			body = string(data)
		}
		if body != "" {
			tk.Body = body + "\n"
		}

		blockedBy, _ := cmd.Flags().GetStringSlice("blocked-by")
		for _, b := range blockedBy {
			tk.BlockedBy = append(tk.BlockedBy, task.Dependency{
				UID:     b,
				RelType: "FINISHTOSTART",
			})
		}

		path, err := v.CreateTask(tk)
		if err != nil {
			return err
		}
		fmt.Println(path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().String("title", "", "Task title (required)")
	createCmd.Flags().String("status", "", "Status (default: open)")
	createCmd.Flags().String("priority", "", "Priority (default: normal)")
	createCmd.Flags().String("due", "", "Due date (YYYY-MM-DD)")
	createCmd.Flags().String("scheduled", "", "Scheduled date")
	createCmd.Flags().StringSlice("tag", nil, "Add tag (repeatable)")
	createCmd.Flags().StringSlice("context", nil, "Add context (repeatable)")
	createCmd.Flags().StringSlice("project", nil, "Add project link (repeatable)")
	createCmd.Flags().Int("time-estimate", 0, "Time estimate in minutes")
	createCmd.Flags().StringSlice("blocked-by", nil, "Add dependency (repeatable)")
	createCmd.Flags().String("recurrence", "", "RFC 5545 recurrence rule")
	createCmd.Flags().String("body", "", "Task body content")
	createCmd.Flags().String("body-file", "", "Read body from file")
}
