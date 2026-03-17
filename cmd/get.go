package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thepsadmin/tasknotes-cli/internal/format"
)

var getCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Get details for a single task",
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

		// Populate blocking info
		tasks, _ := v.ListTasks()
		v.BuildBlockingMap(tasks)
		// Find the task in the list to get blocking info
		for _, t := range tasks {
			if t.Path == tk.Path {
				tk.Blocking = t.Blocking
				tk.IsBlocked = t.IsBlocked
				break
			}
		}

		switch formatFlag {
		case "json":
			format.JSONTask(os.Stdout, tk)
		default:
			format.TaskDetail(os.Stdout, tk)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
