package system

import (
	"errors"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
)

// CheckRequirements verifies system requirements
func CheckRequirements() error {
	// Check if running on SLES
	if err := checkSLES(); err != nil {
		return err
	}

	// Check for root privileges
	if os.Geteuid() != 0 {
		return errors.New("root privileges required, please run with sudo")
	}

	// Check if zypper is available
	if _, err := exec.LookPath("zypper"); err != nil {
		return errors.New("zypper command not found")
	}

	return nil
}

// checkSLES checks if running on SLES Linux
func checkSLES() error {
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return errors.New("cannot read /etc/os-release")
	}

	if !strings.Contains(string(data), "SLES") {
		return errors.New("this tool only supports SLES Linux")
	}

	return nil
}

// RestartDocker restarts Docker service
func RestartDocker() error {
	// Reload systemd configuration
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return err
	}

	// Restart Docker service
	if err := exec.Command("systemctl", "restart", "docker").Run(); err != nil {
		return err
	}

	// Enable Docker on boot
	return exec.Command("systemctl", "enable", "docker").Run()
}

// AddUserToDockerGroup adds current user to docker group
func AddUserToDockerGroup() error {
	// Get original user who invoked sudo
	originalUser := os.Getenv("SUDO_USER")
	if originalUser == "" || originalUser == "root" {
		return errors.New("cannot determine original user")
	}

	// Check if docker group exists
	if err := exec.Command("getent", "group", "docker").Run(); err != nil {
		// Create docker group
		if err := exec.Command("groupadd", "docker").Run(); err != nil {
			return err
		}
	}

	// Add user to docker group
	return exec.Command("usermod", "-aG", "docker", originalUser).Run()
}

// IsCIDROccupied checks if CIDR is already in use
func IsCIDROccupied(cidr string) bool {
	// Remove mask part, extract network address
	ip := strings.Split(cidr, "/")[0]

	// Check routing table
	cmd := exec.Command("ip", "route", "show", "to", ip+"/16")
	output, _ := cmd.Output()

	return len(output) > 0
}

// BackupFile creates a backup of the specified file
func BackupFile(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	return exec.Command("cp", path, path+".bak").Run()
}

// CurrentUserUID returns current user's UID
func CurrentUserUID() int {
	u, _ := user.Current()
	uid, _ := strconv.Atoi(u.Uid)
	return uid
}
