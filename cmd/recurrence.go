package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var recurrenceCmd = &cobra.Command{
	Use:   "recurrence <path>",
	Short: "Show recurrence info",
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

		if tk.Recurrence == "" {
			return fmt.Errorf("task has no recurrence rule")
		}

		fmt.Printf("path: %s\n", tk.Path)
		fmt.Printf("rule: %s\n", tk.Recurrence)
		anchor := tk.RecurrenceAnchor
		if anchor == "" {
			anchor = "scheduled"
		}
		fmt.Printf("anchor: %s\n", anchor)
		fmt.Printf("next: %s\n", tk.Scheduled)
		fmt.Printf("completed: %d\n", len(tk.CompleteInstances))
		fmt.Printf("skipped: %d\n", len(tk.SkippedInstances))
		return nil
	},
}

var recurrenceSkipCmd = &cobra.Command{
	Use:   "skip <path>",
	Short: "Skip current recurrence instance",
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

		if tk.Recurrence == "" {
			return fmt.Errorf("task has no recurrence rule")
		}

		today := time.Now().UTC().Format("2006-01-02")
		skippedDate := tk.Scheduled
		if skippedDate == "" {
			skippedDate = today
		}

		tk.SkippedInstances = append(tk.SkippedInstances, skippedDate)

		nextDate, err := computeNextRecurrence(tk.Recurrence, tk.Scheduled, today)
		if err != nil {
			return err
		}
		tk.Scheduled = nextDate
		tk.DateModified = time.Now().UTC().Format(time.RFC3339)

		if err := v.SaveTask(tk); err != nil {
			return err
		}
		fmt.Printf("%s\tskipped=%s\tnext=%s\n", tk.Path, skippedDate, nextDate)
		return nil
	},
}

var recurrenceHistoryCmd = &cobra.Command{
	Use:   "history <path>",
	Short: "Show completed/skipped instances",
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

		fmt.Println("date\taction")

		// Merge and sort completed/skipped by date descending
		type entry struct {
			date   string
			action string
		}
		var entries []entry
		for _, d := range tk.CompleteInstances {
			entries = append(entries, entry{d, "completed"})
		}
		for _, d := range tk.SkippedInstances {
			entries = append(entries, entry{d, "skipped"})
		}

		// Sort descending by date
		for i := 0; i < len(entries); i++ {
			for j := i + 1; j < len(entries); j++ {
				if entries[j].date > entries[i].date {
					entries[i], entries[j] = entries[j], entries[i]
				}
			}
		}

		for _, e := range entries {
			fmt.Printf("%s\t%s\n", e.date, e.action)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(recurrenceCmd)
	recurrenceCmd.AddCommand(recurrenceSkipCmd)
	recurrenceCmd.AddCommand(recurrenceHistoryCmd)
}
