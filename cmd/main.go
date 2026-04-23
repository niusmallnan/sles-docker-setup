package main

import (
	"os"

	"sles-docker-setup/internal/config"
	"sles-docker-setup/internal/install"
	"sles-docker-setup/internal/system"
	"sles-docker-setup/internal/ui"

	"github.com/spf13/cobra"
)

var version = "Dev"

func main() {
	Execute()
}

var rootCmd = &cobra.Command{
	Use:   "docker-pilot",
	Short: "Docker Pilot - setup & manage Docker on SLES 15+",
	Long:  `Docker Pilot - Interactive Docker installation and TUI management tool for SUSE Linux Enterprise Server.`,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Docker (registry, proxy, network)",
	Long:  `Interactive configuration for Docker: registry, proxy, and container network CIDR. This is the default command.`,
	Run:   runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func Execute() {
	// If no subcommand is provided, default to "config"
	if len(os.Args) == 1 {
		os.Args = append(os.Args, "config")
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runConfig(cmd *cobra.Command, args []string) {
	ui.PrintBanner(version)

	// Check if running in container
	if system.IsRunningInContainer() {
		ui.PrintWarning("Running in a container environment - Docker installation will be skipped since we are already in Docker")
	}

	// Select mode
	modeOptions := []string{
		"Quick mode (install Docker only, skip configuration)",
		"Advanced mode (full setup: install + configure registry/proxy/network)",
	}
	modeChoice, err := ui.AskSelect("Select installation mode", modeOptions, 0)
	if err != nil {
		ui.PrintError("Failed to select mode: %v", err)
		os.Exit(1)
	}
	isAdvancedMode := modeChoice == 1

	// Step 1: System Check
	totalSteps := 2
	if isAdvancedMode {
		totalSteps = 5
	}
	ui.PrintStep(1, totalSteps, "System Check")
	if err := system.CheckRequirements(); err != nil {
		ui.PrintError("System check failed: %v", err)
		os.Exit(1)
	}
	ui.PrintSuccess("System check passed")

	// Step 2: Install Docker Engine
	ui.PrintStep(2, totalSteps, "Install Docker Engine")
	skipDockerInstall := false
	if system.IsRunningInContainer() {
		ui.PrintWarning("Detected container environment - skipping Docker installation")
		skipDockerInstall = true
	} else if install.IsDockerInstalled() {
		ui.PrintWarning("Docker is already installed")
		if !ui.AskConfirm("Skip installation and proceed?", true) {
			skipDockerInstall = true
		}
	}

	if !skipDockerInstall {
		if err := install.InstallDocker(); err != nil {
			ui.PrintError("Docker installation failed: %v", err)
			os.Exit(1)
		}
		ui.PrintSuccess("Docker Engine installed")
	}

	if isAdvancedMode {
		// Step 3: Configure Registry
		ui.PrintStep(3, totalSteps, "Configure Registry")
		registryConfig, err := config.AskRegistryConfig()
		if err != nil {
			ui.PrintError("Registry configuration failed: %v", err)
			os.Exit(1)
		}
		if registryConfig.Configured {
			if err := config.WriteRegistryConfig(registryConfig); err != nil {
				ui.PrintError("Failed to write Registry configuration: %v", err)
				os.Exit(1)
			}
			ui.PrintSuccess("Registry configuration written")
		} else {
			ui.PrintWarning("Skipping Registry configuration")
		}

		// Step 4: Configure HTTP Proxy
		ui.PrintStep(4, totalSteps, "Configure HTTP Proxy")
		proxyConfig, err := config.AskProxyConfig()
		if err != nil {
			ui.PrintError("Proxy configuration failed: %v", err)
			os.Exit(1)
		}
		if proxyConfig.Configured {
			if err := config.WriteProxyConfig(proxyConfig); err != nil {
				ui.PrintError("Failed to write Proxy configuration: %v", err)
				os.Exit(1)
			}
			ui.PrintSuccess("Proxy configuration written")
		} else {
			ui.PrintWarning("Skipping Proxy configuration")
		}

		// Step 5: Configure Container CIDR
		ui.PrintStep(5, totalSteps, "Configure Container Network CIDR")
		cidrConfig, err := config.AskCIDRConfig()
		if err != nil {
			ui.PrintError("CIDR configuration failed: %v", err)
			os.Exit(1)
		}
		if cidrConfig.Configured {
			if err := config.WriteCIDRConfig(cidrConfig); err != nil {
				ui.PrintError("Failed to write CIDR configuration: %v", err)
				os.Exit(1)
			}
			ui.PrintSuccess("Container network CIDR configuration written")
		} else {
			ui.PrintWarning("Skipping CIDR configuration")
		}
	}

	// Finalize
	if !system.IsRunningInContainer() {
		if err := system.RestartDocker(); err != nil {
			ui.PrintError("Failed to restart Docker service: %v", err)
			os.Exit(1)
		}

		if err := system.AddUserToDockerGroup(); err != nil {
			ui.PrintWarning("Failed to add user to docker group: %v", err)
		}
	} else {
		ui.PrintInfo("Running in container - skipping Docker service restart and group management")
	}

	ui.PrintCompletion()
}
