package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

var updateCmd = &cobra.Command{
	Use:   "update <path>",
	Short: "Update an existing task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		tk, err := v.GetTask(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		changed := false

		if cmd.Flags().Changed("title") {
			tk.Title, _ = cmd.Flags().GetString("title")
			changed = true
		}
		if cmd.Flags().Changed("status") {
			tk.Status, _ = cmd.Flags().GetString("status")
			changed = true
		}
		if cmd.Flags().Changed("priority") {
			tk.Priority, _ = cmd.Flags().GetString("priority")
			changed = true
		}
		if cmd.Flags().Changed("due") {
			tk.Due, _ = cmd.Flags().GetString("due")
			changed = true
		}
		if cmd.Flags().Changed("scheduled") {
			tk.Scheduled, _ = cmd.Flags().GetString("scheduled")
			changed = true
		}
		if cmd.Flags().Changed("time-estimate") {
			tk.TimeEstimate, _ = cmd.Flags().GetInt("time-estimate")
			changed = true
		}
		if cmd.Flags().Changed("recurrence") {
			tk.Recurrence, _ = cmd.Flags().GetString("recurrence")
			changed = true
		}

		// Tag operations
		if addTags, _ := cmd.Flags().GetStringSlice("add-tag"); len(addTags) > 0 {
			for _, tag := range addTags {
				if !containsStr(tk.Tags, tag) {
					tk.Tags = append(tk.Tags, tag)
				}
			}
			changed = true
		}
		if rmTags, _ := cmd.Flags().GetStringSlice("remove-tag"); len(rmTags) > 0 {
			tk.Tags = removeStrs(tk.Tags, rmTags)
			changed = true
		}

		// Context operations
		if addCtx, _ := cmd.Flags().GetStringSlice("add-context"); len(addCtx) > 0 {
			for _, ctx := range addCtx {
				if !containsStr(tk.Contexts, ctx) {
					tk.Contexts = append(tk.Contexts, ctx)
				}
			}
			changed = true
		}
		if rmCtx, _ := cmd.Flags().GetStringSlice("remove-context"); len(rmCtx) > 0 {
			tk.Contexts = removeStrs(tk.Contexts, rmCtx)
			changed = true
		}

		// Dependency operations
		if addDeps, _ := cmd.Flags().GetStringSlice("add-blocked-by"); len(addDeps) > 0 {
			for _, dep := range addDeps {
				tk.BlockedBy = append(tk.BlockedBy, task.Dependency{
					UID:     dep,
					RelType: "FINISHTOSTART",
				})
			}
			changed = true
		}
		if rmDeps, _ := cmd.Flags().GetStringSlice("remove-blocked-by"); len(rmDeps) > 0 {
			newDeps := make([]task.Dependency, 0)
			for _, dep := range tk.BlockedBy {
				if !containsStr(rmDeps, dep.UID) {
					newDeps = append(newDeps, dep)
				}
			}
			tk.BlockedBy = newDeps
			changed = true
		}

		// Body operations
		if cmd.Flags().Changed("body") {
			body, _ := cmd.Flags().GetString("body")
			tk.Body = body + "\n"
			changed = true
		}
		if appendBody, _ := cmd.Flags().GetString("append-body"); appendBody != "" {
			tk.Body = tk.Body + appendBody + "\n"
			changed = true
		}

		if !changed {
			return nil
		}

		tk.DateModified = time.Now().UTC().Format(time.RFC3339)

		if err := v.SaveTask(tk); err != nil {
			return err
		}
		fmt.Println(tk.Path)
		return nil
	},
}

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func removeStrs(slice []string, remove []string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if !containsStr(remove, s) {
			result = append(result, s)
		}
	}
	return result
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("title", "", "Update title")
	updateCmd.Flags().String("status", "", "Update status")
	updateCmd.Flags().String("priority", "", "Update priority")
	updateCmd.Flags().String("due", "", "Update due date")
	updateCmd.Flags().String("scheduled", "", "Update scheduled date")
	updateCmd.Flags().Int("time-estimate", 0, "Update time estimate")
	updateCmd.Flags().String("recurrence", "", "Update recurrence rule")
	updateCmd.Flags().StringSlice("add-tag", nil, "Append a tag")
	updateCmd.Flags().StringSlice("remove-tag", nil, "Remove a tag")
	updateCmd.Flags().StringSlice("add-context", nil, "Append a context")
	updateCmd.Flags().StringSlice("remove-context", nil, "Remove a context")
	updateCmd.Flags().StringSlice("add-blocked-by", nil, "Add a dependency")
	updateCmd.Flags().StringSlice("remove-blocked-by", nil, "Remove a dependency")
	updateCmd.Flags().String("body", "", "Replace body content")
	updateCmd.Flags().String("append-body", "", "Append to body")
}
