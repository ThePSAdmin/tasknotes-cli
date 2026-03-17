package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thepsadmin/tasknotes-cli/internal/filter"
	"github.com/thepsadmin/tasknotes-cli/internal/format"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, cfg, err := loadVault()
		if err != nil {
			return err
		}

		includeArchived, _ := cmd.Flags().GetBool("archived")

		tasks, err := v.ListTasks()
		if err != nil {
			return err
		}
		if includeArchived {
			tasks, err = v.ListAllTasks()
			if err != nil {
				return err
			}
		}

		v.BuildBlockingMap(tasks)

		opts := &filter.Options{}
		opts.Status, _ = cmd.Flags().GetString("status")
		opts.Priority, _ = cmd.Flags().GetString("priority")
		opts.Tags, _ = cmd.Flags().GetStringSlice("tag")
		opts.Context, _ = cmd.Flags().GetString("context")
		opts.Project, _ = cmd.Flags().GetString("project")
		opts.DueBefore, _ = cmd.Flags().GetString("due-before")
		opts.DueAfter, _ = cmd.Flags().GetString("due-after")
		opts.ScheduledBefore, _ = cmd.Flags().GetString("scheduled-before")
		opts.ScheduledAfter, _ = cmd.Flags().GetString("scheduled-after")
		opts.Overdue, _ = cmd.Flags().GetBool("overdue")
		opts.Blocked, _ = cmd.Flags().GetBool("blocked")
		opts.Blocking, _ = cmd.Flags().GetBool("blocking")
		opts.Sort, _ = cmd.Flags().GetString("sort")
		opts.SortDir, _ = cmd.Flags().GetString("sort-dir")
		opts.Limit, _ = cmd.Flags().GetInt("limit")

		filtered := filter.Apply(tasks, opts, cfg)

		fields := []string{"path", "title", "status", "priority", "due"}
		if f, _ := cmd.Flags().GetString("fields"); f != "" {
			fields = strings.Split(f, ",")
		}

		groupKey, _ := cmd.Flags().GetString("group")

		switch formatFlag {
		case "json":
			format.JSON(os.Stdout, filtered)
		default:
			if groupKey != "" {
				format.GroupedTSV(os.Stdout, filtered, groupKey, fields)
			} else {
				format.TSV(os.Stdout, filtered, fields)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().String("status", "", "Filter by status")
	listCmd.Flags().String("priority", "", "Filter by priority")
	listCmd.Flags().StringSlice("tag", nil, "Filter by tag (repeatable)")
	listCmd.Flags().String("context", "", "Filter by context")
	listCmd.Flags().String("project", "", "Filter by project wikilink")
	listCmd.Flags().String("due-before", "", "Tasks due before date (YYYY-MM-DD)")
	listCmd.Flags().String("due-after", "", "Tasks due after date")
	listCmd.Flags().String("scheduled-before", "", "Tasks scheduled before date")
	listCmd.Flags().String("scheduled-after", "", "Tasks scheduled after date")
	listCmd.Flags().Bool("overdue", false, "Tasks where due < today and status != done")
	listCmd.Flags().Bool("blocked", false, "Only show blocked tasks")
	listCmd.Flags().Bool("blocking", false, "Only show tasks that block others")
	listCmd.Flags().Bool("archived", false, "Include archived tasks")
	listCmd.Flags().String("sort", "due", "Sort by: due, scheduled, priority, status, title, dateCreated")
	listCmd.Flags().String("sort-dir", "asc", "Sort direction: asc or desc")
	listCmd.Flags().String("group", "", "Group by: priority, status, project, context, due, tags")
	listCmd.Flags().Int("limit", 0, "Max results")
	listCmd.Flags().String("fields", "", "Columns to show (comma-separated)")
}
