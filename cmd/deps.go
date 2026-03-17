package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thepsadmin/tasknotes-cli/internal/format"
	"github.com/thepsadmin/tasknotes-cli/internal/task"
)

var depsCmd = &cobra.Command{
	Use:   "deps [path]",
	Short: "Show dependency information",
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		tasks, err := v.ListTasks()
		if err != nil {
			return err
		}
		v.BuildBlockingMap(tasks)

		// Build lookup maps
		byTitle := make(map[string]*task.Task)
		byPath := make(map[string]*task.Task)
		for _, t := range tasks {
			byTitle[t.Title] = t
			byPath[t.Path] = t
		}

		showRoots, _ := cmd.Flags().GetBool("roots")
		showLeaves, _ := cmd.Flags().GetBool("leaves")
		chainPath, _ := cmd.Flags().GetString("chain")

		if showRoots {
			// Root tasks: not blocked by anything and blocking something
			var roots []*task.Task
			for _, t := range tasks {
				if !t.IsBlocked && len(t.Blocking) > 0 {
					roots = append(roots, t)
				}
			}
			fields := []string{"path", "title", "status", "blocking_count"}
			format.TSV(os.Stdout, roots, fields)
			return nil
		}

		if showLeaves {
			// Leaf tasks: blocked but not blocking anything
			var leaves []*task.Task
			for _, t := range tasks {
				if len(t.Blocking) == 0 && t.IsBlocked {
					leaves = append(leaves, t)
				}
			}
			fields := []string{"path", "title", "status"}
			format.TSV(os.Stdout, leaves, fields)
			return nil
		}

		if chainPath != "" {
			// Show full chain from root to leaf through the given task
			chain := buildChain(byTitle, byPath, chainPath)
			parts := make([]string, len(chain))
			for i, t := range chain {
				parts[i] = t.Path
			}
			fmt.Println(strings.Join(parts, " -> "))
			return nil
		}

		// Default: show dependency tree for a specific task
		if len(args) == 0 {
			return fmt.Errorf("path argument required (or use --roots, --leaves, --chain)")
		}

		tk, err := v.GetTask(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		// Re-fetch from the loaded tasks to get blocking info
		for _, t := range tasks {
			if t.Path == tk.Path {
				tk = t
				break
			}
		}

		printDepTree(tk, byTitle, 0)
		return nil
	},
}

func printDepTree(t *task.Task, byTitle map[string]*task.Task, indent int) {
	prefix := strings.Repeat("  ", indent)
	fmt.Printf("%s%s (%s)\n", prefix, t.Path, t.Status)
	for _, dep := range t.BlockedBy {
		reltype := dep.RelType
		if blocker, ok := byTitle[dep.UID]; ok {
			fmt.Printf("%s  blocked-by: %s (%s) [%s]\n", prefix, blocker.Path, blocker.Status, reltype)
			printDepTree(blocker, byTitle, indent+2)
		} else {
			fmt.Printf("%s  blocked-by: [[%s]] (unresolved) [%s]\n", prefix, dep.UID, reltype)
		}
	}
}

func buildChain(byTitle map[string]*task.Task, byPath map[string]*task.Task, targetPath string) []*task.Task {
	target := byPath[targetPath]
	if target == nil {
		return nil
	}

	// Walk up to find root
	var chain []*task.Task
	visited := make(map[string]bool)
	current := target

	// Build path from target to root
	var upward []*task.Task
	for current != nil {
		if visited[current.Path] {
			break
		}
		visited[current.Path] = true
		upward = append(upward, current)
		if len(current.BlockedBy) == 0 {
			break
		}
		// Follow first dependency
		current = byTitle[current.BlockedBy[0].UID]
	}

	// Reverse to get root-first order
	for i := len(upward) - 1; i >= 0; i-- {
		chain = append(chain, upward[i])
	}

	// Walk down from target to find leaves
	current = target
	visited = make(map[string]bool)
	visited[target.Path] = true
	for len(current.Blocking) > 0 {
		next := byPath[current.Blocking[0]]
		if next == nil || visited[next.Path] {
			break
		}
		visited[next.Path] = true
		chain = append(chain, next)
		current = next
	}

	return chain
}

func init() {
	rootCmd.AddCommand(depsCmd)
	depsCmd.Flags().Bool("roots", false, "Show all unblocked root tasks")
	depsCmd.Flags().Bool("leaves", false, "Show all leaf tasks")
	depsCmd.Flags().String("chain", "", "Show full chain through task path")
}
