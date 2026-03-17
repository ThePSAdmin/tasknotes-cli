package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive <path>",
	Short: "Archive a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _, err := loadVault()
		if err != nil {
			return err
		}

		path := args[0]
		if _, err := v.GetTask(path); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}

		newPath, err := v.ArchiveTask(path)
		if err != nil {
			return err
		}
		fmt.Printf("%s -> %s\n", path, newPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(archiveCmd)
}
