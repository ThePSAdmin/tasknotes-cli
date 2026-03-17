package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	rrule "github.com/teambition/rrule-go"
)

var completeCmd = &cobra.Command{
	Use:   "complete <path>",
	Short: "Complete a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, cfg, err := loadVault()
		if err != nil {
			return err
		}

		tk, err := v.GetTask(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		now := time.Now().UTC()
		today := now.Format("2006-01-02")

		if tk.Recurrence != "" {
			// Recurring task: record the completed instance and advance
			tk.CompleteInstances = append(tk.CompleteInstances, today)

			// Compute next scheduled date
			nextDate, err := computeNextRecurrence(tk.Recurrence, tk.Scheduled, today)
			if err != nil {
				return fmt.Errorf("computing next recurrence: %w", err)
			}
			tk.Scheduled = nextDate

			// Keep status as-is (typically open)
			if cfg.IsCompletedStatus(tk.Status) {
				tk.Status = "open"
			}

			tk.DateModified = now.Format(time.RFC3339)
			if err := v.SaveTask(tk); err != nil {
				return err
			}
			fmt.Printf("%s\t%s\tscheduled=%s\tcompleted_instance=%s\n", tk.Path, tk.Status, nextDate, today)
		} else {
			// Non-recurring: mark as done
			tk.Status = cfg.CompletedStatusValue()
			tk.CompletedDate = today
			tk.DateModified = now.Format(time.RFC3339)

			if err := v.SaveTask(tk); err != nil {
				return err
			}
			fmt.Printf("%s\t%s\tcompletedDate=%s\n", tk.Path, tk.Status, today)
		}
		return nil
	},
}

func computeNextRecurrence(rule string, scheduled string, today string) (string, error) {
	// Parse the scheduled date to use as dtstart
	var dtstart time.Time
	if scheduled != "" {
		// Try date-only first, then with time
		var err error
		dtstart, err = time.Parse("2006-01-02", scheduled)
		if err != nil {
			dtstart, err = time.Parse(time.RFC3339, scheduled)
			if err != nil {
				// Try datetime without timezone
				dtstart, err = time.Parse("2006-01-02T15:04", scheduled)
				if err != nil {
					dtstart = time.Now().UTC()
				}
			}
		}
	} else {
		dtstart = time.Now().UTC()
	}

	todayDate, _ := time.Parse("2006-01-02", today)

	ruleStr := rule
	if !strings.Contains(strings.ToUpper(ruleStr), "DTSTART") {
		ruleStr = fmt.Sprintf("DTSTART:%s\nRRULE:%s", dtstart.Format("20060102T150405Z"), ruleStr)
	}

	r, err := rrule.StrToRRule(ruleStr)
	if err != nil {
		return "", err
	}

	// Get occurrences after today
	after := todayDate.Add(24 * time.Hour)
	instances := r.Between(after, after.AddDate(1, 0, 0), true)
	if len(instances) > 0 {
		return instances[0].Format("2006-01-02"), nil
	}

	// Fallback: just return tomorrow
	return todayDate.Add(24 * time.Hour).Format("2006-01-02"), nil
}

func init() {
	rootCmd.AddCommand(completeCmd)
}
