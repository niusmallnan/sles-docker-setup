package install

import (
	"os"
	"os/exec"
)

// IsDockerInstalled checks if Docker is installed
func IsDockerInstalled() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// InstallDocker installs Docker Engine
func InstallDocker() error {
	// Step 1: Refresh repositories
	if err := refreshRepo(); err != nil {
		return err
	}

	// Step 2: Install Docker packages (using SLES built-in sources)
	if err := installDockerPackages(); err != nil {
		return err
	}

	// Step 3: Start Docker service
	if err := startDockerService(); err != nil {
		return err
	}

	return nil
}

func refreshRepo() error {
	return exec.Command("zypper", "--non-interactive", "refresh").Run()
}

func installDockerPackages() error {
	// Install Docker using SLES built-in sources
	return exec.Command("zypper", "--non-interactive", "install",
		"docker",
		"docker-compose",
	).Run()
}

func startDockerService() error {
	if err := exec.Command("systemctl", "start", "docker").Run(); err != nil {
		return err
	}

	return exec.Command("systemctl", "enable", "docker").Run()
}

// UninstallDocker uninstalls Docker
func UninstallDocker() error {
	// Stop service
	exec.Command("systemctl", "stop", "docker").Run()

	// Uninstall packages
	if err := exec.Command("zypper", "--non-interactive", "remove",
		"docker-ce",
		"docker-ce-cli",
		"containerd.io",
		"docker-buildx-plugin",
		"docker-compose-plugin",
	).Run(); err != nil {
		return err
	}

	// Remove config files (optional)
	os.RemoveAll("/etc/docker")
	os.RemoveAll("/var/lib/docker")

	return nil
}
