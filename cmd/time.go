package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Time tracking commands",
}

var timeStartCmd = &cobra.Command{
	Use:   "start <path>",
	Short: "Start time tracking",
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

		// Check if already tracking
		for _, e := range tk.TimeEntries {
			if e.EndTime.IsZero() {
				return fmt.Errorf("already tracking %s (started %s)", tk.Path, e.StartTime.Format(time.RFC3339))
			}
		}

		now := time.Now().UTC()
		tk.TimeEntries = append(tk.TimeEntries, task.TimeEntry{
			StartTime: now,
		})
		tk.DateModified = now.Format(time.RFC3339)

		if err := v.SaveTask(tk); err != nil {
			return err
		}
		fmt.Printf("started %s %s\n", tk.Path, now.Format(time.RFC3339))
		return nil
	},
}

var timeStopCmd = &cobra.Command{
	Use:   "stop <path>",
	Short: "Stop time tracking",
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

		now := time.Now().UTC()
		desc, _ := cmd.Flags().GetString("desc")

		found := false
		for i := range tk.TimeEntries {
			if tk.TimeEntries[i].EndTime.IsZero() {
				tk.TimeEntries[i].EndTime = now
				if desc != "" {
					tk.TimeEntries[i].Description = desc
				}
				duration := now.Sub(tk.TimeEntries[i].StartTime)
				fmt.Printf("stopped %s %dm\n", tk.Path, int(duration.Minutes()))
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("no active time tracking for %s", tk.Path)
		}

		tk.DateModified = now.Format(time.RFC3339)
		tk.TotalTrackedTime = computeTotal(tk.TimeEntries)

		return v.SaveTask(tk)
	},
}

var timeLogCmd = &cobra.Command{
	Use:   "log <path>",
	Short: "Show time entries",
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

		fmt.Println("start\tend\tduration\tdescription")
		var total int
		for _, e := range tk.TimeEntries {
			endStr := ""
			durStr := ""
			if !e.EndTime.IsZero() {
				endStr = e.EndTime.Format(time.RFC3339)
				dur := int(e.EndTime.Sub(e.StartTime).Minutes())
				durStr = fmt.Sprintf("%dm", dur)
				total += dur
			} else {
				endStr = "(active)"
				dur := int(time.Since(e.StartTime).Minutes())
				durStr = fmt.Sprintf("%dm", dur)
				total += dur
			}
			fmt.Printf("%s\t%s\t%s\t%s\n", e.StartTime.Format(time.RFC3339), endStr, durStr, e.Description)
		}

		totalStr := fmt.Sprintf("total: %dm", total)
		if tk.TimeEstimate > 0 {
			totalStr += fmt.Sprintf(" (estimate: %dm)", tk.TimeEstimate)
		}
		fmt.Println(totalStr)
		return nil
	},
}

var timeSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Aggregate time across tasks",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		tasks, err := v.ListTasks()
		if err != nil {
			return err
		}

		since, _ := cmd.Flags().GetString("since")
		tag, _ := cmd.Flags().GetString("tag")

		fmt.Println("path\ttitle\ttracked\testimate")
		var grandTotal int
		for _, tk := range tasks {
			if tag != "" && !containsStr(tk.Tags, tag) {
				continue
			}

			tracked := filteredTrackedTime(tk.TimeEntries, since)
			if tracked == 0 {
				continue
			}
			grandTotal += tracked

			est := ""
			if tk.TimeEstimate > 0 {
				est = fmt.Sprintf("%dm", tk.TimeEstimate)
			}
			fmt.Printf("%s\t%s\t%dm\t%s\n", tk.Path, tk.Title, tracked, est)
		}
		fmt.Printf("total: %dm\n", grandTotal)
		return nil
	},
}

func computeTotal(entries []task.TimeEntry) int {
	total := 0
	for _, e := range entries {
		if !e.EndTime.IsZero() {
			total += int(e.EndTime.Sub(e.StartTime).Minutes())
		}
	}
	return total
}

func filteredTrackedTime(entries []task.TimeEntry, since string) int {
	total := 0
	var sinceTime time.Time
	if since != "" {
		sinceTime, _ = time.Parse("2006-01-02", since)
	}

	for _, e := range entries {
		if !sinceTime.IsZero() && e.StartTime.Before(sinceTime) {
			continue
		}
		if !e.EndTime.IsZero() {
			total += int(e.EndTime.Sub(e.StartTime).Minutes())
		}
	}
	return total
}

func init() {
	rootCmd.AddCommand(timeCmd)
	timeCmd.AddCommand(timeStartCmd)
	timeCmd.AddCommand(timeStopCmd)
	timeCmd.AddCommand(timeLogCmd)
	timeCmd.AddCommand(timeSummaryCmd)

	timeStopCmd.Flags().String("desc", "", "Description for the time entry")
	timeSummaryCmd.Flags().String("since", "", "Only count entries after date (YYYY-MM-DD)")
	timeSummaryCmd.Flags().String("tag", "", "Filter by tag")
}
