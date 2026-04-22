package ui

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

var (
	bold   = color.New(color.Bold).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	cyan   = color.New(color.FgCyan).SprintFunc()
)

// PrintBanner prints the tool banner
func PrintBanner(version string) {
	fmt.Println()
	fmt.Println(bold("═══════════════════════════════════════════════════════════════════"))
	fmt.Printf(bold("  Docker Pilot  |  Version: %s\n"), version)
	fmt.Println(bold("  Docker setup & TUI management for SLES 15+"))
	fmt.Println(bold("═══════════════════════════════════════════════════════════════════"))
	fmt.Println()
}

// PrintStep prints the step title
func PrintStep(current, total int, title string) {
	fmt.Printf("[%d/%d] %s\n", current, total, bold(title))
}

// PrintSuccess prints success message
func PrintSuccess(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", green("✓"), msg)
}

// PrintWarning prints warning message
func PrintWarning(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", yellow("→"), msg)
}

// PrintError prints error message
func PrintError(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", red("✗"), msg)
}

// PrintInfo prints info message
func PrintInfo(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Printf("  %s %s\n", cyan("→"), msg)
}

// AskInput asks for user input
func AskInput(prompt, defaultValue, help string) (string, error) {
	var answer string
	p := &survey.Input{
		Message: prompt,
		Default: defaultValue,
		Help:    help,
	}
	err := survey.AskOne(p, &answer)
	return answer, err
}

// AskSelect asks user to select from options
func AskSelect(prompt string, options []string, defaultIndex int) (int, error) {
	var answer int
	p := &survey.Select{
		Message: prompt,
		Options: options,
		Default: defaultIndex,
	}
	err := survey.AskOne(p, &answer)
	return answer, err
}

// AskConfirm asks for user confirmation
func AskConfirm(prompt string, defaultValue bool) bool {
	var answer bool
	p := &survey.Confirm{
		Message: prompt,
		Default: defaultValue,
	}
	survey.AskOne(p, &answer)
	return answer
}

// PrintCompletion prints completion message
func PrintCompletion() {
	fmt.Println()
	fmt.Println(green("───────────────────────────────────────────"))
	fmt.Println(green("✅ All configurations completed!"))
	fmt.Println()
	fmt.Println("   Run the following command to apply changes:")
	fmt.Println("     newgrp docker  # or re-login")
	fmt.Println()
	fmt.Println("   Test Docker:")
	fmt.Println("     docker run hello-world")
	fmt.Println()
	fmt.Println("   Configuration file locations:")
	fmt.Println("     - /etc/docker/daemon.json")
	fmt.Println("     - /etc/systemd/system/docker.service.d/http-proxy.conf")
	fmt.Println()
	fmt.Println("   Issues? Contact Ops Team #docker-support")
	fmt.Println(green("───────────────────────────────────────────"))
	fmt.Println()
}
