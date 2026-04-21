package main

import (
	"os"

	"sles-docker-setup/internal/config"
	"sles-docker-setup/internal/install"
	"sles-docker-setup/internal/system"
	"sles-docker-setup/internal/ui"
)

func main() {
	ui.PrintBanner()

	// Step 1: System Check
	ui.PrintStep(1, 5, "System Check")
	if err := system.CheckRequirements(); err != nil {
		ui.PrintError("System check failed: %v", err)
		os.Exit(1)
	}
	ui.PrintSuccess("System check passed")

	// Step 2: Install Docker Engine
	ui.PrintStep(2, 5, "Install Docker Engine")
	if install.IsDockerInstalled() {
		ui.PrintWarning("Docker is already installed")
		if !ui.AskConfirm("Skip installation and proceed to configuration?", true) {
			if err := install.InstallDocker(); err != nil {
				ui.PrintError("Docker installation failed: %v", err)
				os.Exit(1)
			}
		}
	} else {
		if err := install.InstallDocker(); err != nil {
			ui.PrintError("Docker installation failed: %v", err)
			os.Exit(1)
		}
	}
	ui.PrintSuccess("Docker Engine installed")

	// Step 3: Configure Registry
	ui.PrintStep(3, 5, "Configure Registry")
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
	ui.PrintStep(4, 5, "Configure HTTP Proxy")
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
	ui.PrintStep(5, 5, "Configure Container Network CIDR")
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

	// Finalize
	if err := system.RestartDocker(); err != nil {
		ui.PrintError("Failed to restart Docker service: %v", err)
		os.Exit(1)
	}

	if err := system.AddUserToDockerGroup(); err != nil {
		ui.PrintWarning("Failed to add user to docker group: %v", err)
	}

	ui.PrintCompletion()
}
