package install

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestIsDockerInstalled(t *testing.T) {
	// Test that the function doesn't panic
	result := IsDockerInstalled()
	t.Logf("IsDockerInstalled returned: %v", result)

	// Verify by checking the command directly
	_, err := exec.LookPath("docker")
	expected := err == nil
	if result != expected {
		t.Errorf("IsDockerInstalled() = %v, but exec.LookPath indicates %v", result, expected)
	}
}

func TestInstallDockerPlatform(t *testing.T) {
	// On non-Linux or non-SLES, this should fail
	// We just verify it doesn't panic
	if runtime.GOOS != "linux" {
		err := InstallDocker()
		if err == nil {
			t.Error("Expected InstallDocker to fail on non-Linux platform")
		}
		t.Logf("InstallDocker on %s: %v", runtime.GOOS, err)
	}
}

func TestUninstallDockerPlatform(t *testing.T) {
	// Just verify no panic
	if runtime.GOOS != "linux" {
		err := UninstallDocker()
		// May or may not error depending on tools, just ensure no panic
		t.Logf("UninstallDocker on %s: %v", runtime.GOOS, err)
	}
}
