package main

import (
	"embed"
	"os"
	"os/exec"
	"path/filepath"

	"docker-pilot/internal/ui"
	"github.com/spf13/cobra"
)

//go:embed embed/lazydocker
var lazyDockerFS embed.FS

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch lazydocker TUI for Docker management",
	Long:  `Launch lazydocker - a powerful terminal UI for Docker and Docker Compose. Built-in, no external dependencies.`,
	Run:   runTui,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTui(cmd *cobra.Command, args []string) {
	ui.PrintBanner(version)

	// Check if docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		ui.PrintWarning("Docker not found on system")
		if !ui.AskConfirm("Continue anyway?", false) {
			return
		}
	}

	ui.PrintInfo("Starting built-in lazydocker TUI...")
	ui.PrintInfo("Press 'q' to quit")
	ui.PrintInfo("")

	// Extract and run embedded lazydocker
	if err := runEmbeddedLazyDocker(); err != nil {
		ui.PrintError("lazydocker exited with error: %v", err)
	}
}

func runEmbeddedLazyDocker() error {
	// Read embedded binary
	data, err := lazyDockerFS.ReadFile("embed/lazydocker")
	if err != nil {
		return err
	}

	// Write to temp file
	tmpDir, err := os.MkdirTemp("", "lazydocker-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpPath := filepath.Join(tmpDir, "lazydocker")
	if err := os.WriteFile(tmpPath, data, 0755); err != nil {
		return err
	}

	// Run lazydocker
	cmd := exec.Command(tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
