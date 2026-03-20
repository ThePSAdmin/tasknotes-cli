package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/thepsadmin/tasknotes-cli/internal/config"
	"github.com/thepsadmin/tasknotes-cli/internal/vault"
)

var (
	Version   = "dev"
	CommitSHA = "none"

	vaultPath  string
	formatFlag string
)

var rootCmd = &cobra.Command{
	Use:   "tasknotes",
	Short: "CLI for managing TaskNotes tasks",
	Long:  "Manage tasks in an Obsidian vault with the TaskNotes plugin, optimized for LLM usage.",
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&vaultPath, "vault", "", "Path to the obsidian vault root (or set TASKNOTES_VAULT env var)")
	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "text", "Output format: text, json, tsv")
}

func getVaultPath() (string, error) {
	if vaultPath != "" {
		return vaultPath, nil
	}
	if v := os.Getenv("TASKNOTES_VAULT"); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("vault path required: use --vault or set TASKNOTES_VAULT")
}

func loadVault() (*vault.Vault, *config.Config, error) {
	vp, err := getVaultPath()
	if err != nil {
		return nil, nil, err
	}
	cfg, err := config.Load(vp)
	if err != nil {
		return nil, nil, fmt.Errorf("loading config: %w", err)
	}
	return vault.New(cfg), cfg, nil
}
