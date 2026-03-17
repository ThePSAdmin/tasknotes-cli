package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show task statistics",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, cfg, err := loadVault()
		if err != nil {
			return err
		}

		tasks, err := v.ListTasks()
		if err != nil {
			return err
		}
		v.BuildBlockingMap(tasks)

		tag, _ := cmd.Flags().GetString("tag")
		project, _ := cmd.Flags().GetString("project")

		today := time.Now().Format("2006-01-02")
		sevenDaysAgo := time.Now().AddDate(0, 0, -7).Format("2006-01-02")

		counts := make(map[string]int)
		var overdue, blocked, totalTracked int
		var completed7d, total7d int

		for _, tk := range tasks {
			if tag != "" && !containsStr(tk.Tags, tag) {
				continue
			}
			if project != "" && !containsStr(tk.Projects, project) {
				continue
			}

			counts[tk.Status]++

			if tk.Due != "" && tk.Due < today && !cfg.IsCompletedStatus(tk.Status) {
				overdue++
			}
			if tk.IsBlocked {
				blocked++
			}
			totalTracked += tk.TotalTrackedTime

			// 7-day completion rate
			if tk.CompletedDate >= sevenDaysAgo {
				completed7d++
			}
			if tk.DateCreated >= sevenDaysAgo || (tk.CompletedDate >= sevenDaysAgo) {
				total7d++
			}
		}

		for _, s := range cfg.Statuses {
			if c, ok := counts[s.Value]; ok && c > 0 {
				fmt.Printf("%s: %d\n", s.Value, c)
			}
		}
		fmt.Printf("overdue: %d\n", overdue)
		fmt.Printf("blocked: %d\n", blocked)
		fmt.Printf("total_tracked: %dm\n", totalTracked)
		if total7d > 0 {
			fmt.Printf("completion_rate_7d: %d/%d\n", completed7d, total7d)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().String("since", "", "Stats since date")
	statsCmd.Flags().String("tag", "", "Filter by tag")
	statsCmd.Flags().String("project", "", "Filter by project")
}
