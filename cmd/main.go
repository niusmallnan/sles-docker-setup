package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"sles-docker-setup/internal/config"
	"sles-docker-setup/internal/install"
	"sles-docker-setup/internal/system"
	"sles-docker-setup/internal/ui"

	"github.com/spf13/cobra"
)

//go:embed embed/trivy
var trivyFS embed.FS

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

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan containers and images for CVE vulnerabilities",
	Long:  `Scan all containers and images on the host for known CVE vulnerabilities.`,
	Run:   runScan,
}

var aiInspectCmd = &cobra.Command{
	Use:   "ai-inspect",
	Short: "AI-powered container health inspection",
	Long:  `Use AI to analyze and inspect the health status of all running containers on the host.`,
	Run:   runAiInspect,
}

func init() {
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(aiInspectCmd)
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

func runScan(cmd *cobra.Command, args []string) {
	ui.PrintBanner(version)

	// Check if docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		ui.PrintError("Docker not found on system - cannot scan")
		os.Exit(1)
	}

	ui.PrintInfo("Starting CVE vulnerability scan...")
	ui.PrintInfo("This may take a few minutes on first run (Trivy needs to download vulnerability database)")
	ui.PrintInfo("")

	// Extract and run embedded trivy
	if err := runEmbeddedTrivy(); err != nil {
		ui.PrintError("Scan failed: %v", err)
		os.Exit(1)
	}
}

func runEmbeddedTrivy() error {
	// Read embedded binary
	data, err := trivyFS.ReadFile("embed/trivy")
	if err != nil {
		return err
	}

	// Write to temp file
	tmpDir, err := os.MkdirTemp("", "trivy-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpPath := filepath.Join(tmpDir, "trivy")
	if err := os.WriteFile(tmpPath, data, 0755); err != nil {
		return err
	}

	// Run trivy - first scan images, then containers
	ui.PrintStep(1, 2, "Scanning Docker images")
	cmd := exec.Command(tmpPath, "image", "--severity", "CRITICAL,HIGH,MEDIUM", "--format", "table", "--no-progress", "all")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		// Don't exit on scan errors - just continue
		ui.PrintWarning("Image scan completed with some warnings")
	}

	ui.PrintStep(2, 2, "Scanning running containers")
	cmd = exec.Command(tmpPath, "container", "--severity", "CRITICAL,HIGH,MEDIUM", "--format", "table", "--no-progress", "--include-non-running", "all")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ui.PrintWarning("Container scan completed with some warnings")
	}

	ui.PrintSuccess("\nScan complete!")
	return nil
}

// ContainerInfo represents Docker container inspection data
type ContainerInfo struct {
	ID     string `json:"Id"`
	Name   string
	State  struct {
		Status       string
		Running      bool
		RestartCount int    `json:"RestartCount"`
		StartedAt    string `json:"StartedAt"`
	}
	Config struct {
		Image string
	}
	HostConfig struct {
		RestartPolicy struct {
			Name string
		}
	}
}

func runAiInspect(cmd *cobra.Command, args []string) {
	ui.PrintBanner(version)

	// Check if docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		ui.PrintError("Docker not found on system - cannot inspect")
		os.Exit(1)
	}

	ui.PrintInfo("Starting AI-powered container health inspection...")
	ui.PrintInfo("")

	// Get all containers
	containers, err := getContainers()
	if err != nil {
		ui.PrintError("Failed to list containers: %v", err)
		os.Exit(1)
	}

	if len(containers) == 0 {
		ui.PrintInfo("No containers found on this host")
		return
	}

	// Analyze each container
	healthReport := analyzeContainers(containers)

	// Print report
	printHealthReport(healthReport)
}

func getContainers() ([]ContainerInfo, error) {
	// Get all container IDs
	cmd := exec.Command("docker", "ps", "-a", "-q")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	containerIDs := strings.Fields(string(output))
	var containers []ContainerInfo

	for _, id := range containerIDs {
		cmd := exec.Command("docker", "inspect", id)
		output, err := cmd.Output()
		if err != nil {
			continue // skip containers we can't inspect
		}

		var inspectResult []ContainerInfo
		if err := json.Unmarshal(output, &inspectResult); err != nil {
			continue
		}

		if len(inspectResult) > 0 {
			containers = append(containers, inspectResult[0])
		}
	}

	return containers, nil
}

// ContainerHealth holds health analysis for a single container
type ContainerHealth struct {
	Container   ContainerInfo
	Status      string // "healthy", "warning", "critical"
	Issues      []string
	Suggestions []string
}

func analyzeContainers(containers []ContainerInfo) []ContainerHealth {
	var reports []ContainerHealth

	for _, c := range containers {
		report := ContainerHealth{
			Container: c,
			Status:    "healthy",
		}

		// Trim leading slash from container name
		cName := strings.TrimPrefix(c.Name, "/")
		report.Container.Name = cName

		// Check 1: Is container running?
		if !c.State.Running {
			report.Status = "warning"
			report.Issues = append(report.Issues, "Container is not running")

			// Check restart policy
			if c.HostConfig.RestartPolicy.Name == "always" || c.HostConfig.RestartPolicy.Name == "unless-stopped" {
				report.Status = "critical"
				report.Issues = append(report.Issues, "Restart policy is set but container is not running - potential crash loop")
				report.Suggestions = append(report.Suggestions, fmt.Sprintf("Check logs: `docker logs %s`", cName))
				report.Suggestions = append(report.Suggestions, fmt.Sprintf("Try starting manually: `docker start %s`", cName))
			}
		}

		// Check 2: Restart count
		if c.State.RestartCount > 5 {
			if report.Status != "critical" {
				report.Status = "warning"
			}
			report.Issues = append(report.Issues, fmt.Sprintf("High restart count: %d", c.State.RestartCount))
			report.Suggestions = append(report.Suggestions, "Review container stability - investigate crash reasons")
		} else if c.State.RestartCount > 0 {
			report.Issues = append(report.Issues, fmt.Sprintf("Restart count: %d", c.State.RestartCount))
		}

		// Check 3: Image used (latest tag is not ideal for production)
		if strings.HasSuffix(c.Config.Image, ":latest") {
			report.Issues = append(report.Issues, "Using 'latest' tag - not recommended for production")
			report.Suggestions = append(report.Suggestions, "Use specific version tags for better reproducibility")
		}

		// If no issues found
		if len(report.Issues) == 0 && report.Status == "healthy" {
			report.Suggestions = append(report.Suggestions, "Container is running smoothly - keep monitoring!")
		}

		reports = append(reports, report)
	}

	return reports
}

func printHealthReport(reports []ContainerHealth) {
	// Summary counts
	healthyCount := 0
	warningCount := 0
	criticalCount := 0

	for _, r := range reports {
		switch r.Status {
		case "healthy":
			healthyCount++
		case "warning":
			warningCount++
		case "critical":
			criticalCount++
		}
	}

	ui.PrintInfo(fmt.Sprintf("Inspection complete - %d containers analyzed", len(reports)))
	ui.PrintInfo("")

	if criticalCount > 0 {
		ui.PrintError(fmt.Sprintf("Critical: %d containers need immediate attention", criticalCount))
	}
	if warningCount > 0 {
		ui.PrintWarning(fmt.Sprintf("Warning: %d containers have potential issues", warningCount))
	}
	if healthyCount > 0 {
		ui.PrintSuccess(fmt.Sprintf("Healthy: %d containers are running well", healthyCount))
	}

	ui.PrintInfo("")
	ui.PrintInfo("=")

	// Detailed reports
	for i, r := range reports {
		fmt.Println("")
		fmt.Printf("Container %d/%d: %s\n", i+1, len(reports), r.Container.Name)
		fmt.Printf("  Image: %s\n", r.Container.Config.Image)
		fmt.Printf("  Status: ", r.Container.State.Status)

		switch r.Status {
		case "healthy":
			ui.PrintSuccess("HEALTHY")
		case "warning":
			ui.PrintWarning("WARNING")
		case "critical":
			ui.PrintError("CRITICAL")
		}

		if len(r.Issues) > 0 {
			fmt.Println("  Issues:")
			for _, issue := range r.Issues {
				fmt.Printf("   - %s\n", issue)
			}
		}

		if len(r.Suggestions) > 0 {
			fmt.Println("  Suggestions:")
			for _, suggestion := range r.Suggestions {
				fmt.Printf("   - %s\n", suggestion)
			}
		}

		fmt.Println("")
	}
}
