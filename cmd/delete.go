package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <path>",
	Short: "Delete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		path := args[0]

		// Check task exists
		tk, err := v.GetTask(path)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Fprintf(os.Stderr, "Delete %s (%s)? [y/N] ", path, tk.Title)
			reader := bufio.NewReader(os.Stdin)
			answer, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(answer)) != "y" {
				return nil
			}
		}

		// Remove this task from other tasks' blockedBy references
		tasks, _ := v.ListTasks()
		for _, other := range tasks {
			if other.Path == path {
				continue
			}
			modified := false
			var newDeps []task.Dependency
			for _, dep := range other.BlockedBy {
				if dep.UID == tk.Title {
					modified = true
					continue
				}
				newDeps = append(newDeps, dep)
			}
			if modified {
				other.BlockedBy = newDeps
				v.SaveTask(other)
			}
		}

		if err := v.DeleteTask(path); err != nil {
			return err
		}
		fmt.Printf("deleted %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")
}
