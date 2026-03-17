package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show detected vault settings",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, cfg, err := loadVault()
		if err != nil {
			return err
		}

		tasks, err := v.ListTasks()
		if err != nil {
			return err
		}

		fmt.Printf("vault: %s\n", cfg.VaultPath)
		fmt.Printf("tasks_folder: %s\n", cfg.TasksFolder)
		fmt.Printf("archive_folder: %s\n", cfg.ArchiveFolder)

		statuses := make([]string, len(cfg.Statuses))
		for i, s := range cfg.Statuses {
			statuses[i] = s.Value
		}
		fmt.Printf("statuses: %s\n", strings.Join(statuses, ", "))

		priorities := make([]string, len(cfg.Priorities))
		for i, p := range cfg.Priorities {
			priorities[i] = fmt.Sprintf("%s (%d)", p.Value, p.Weight)
		}
		fmt.Printf("priorities: %s\n", strings.Join(priorities, ", "))

		fmt.Printf("task_tag: %s\n", cfg.TaskTag)
		fmt.Printf("task_count: %d\n", len(tasks))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
