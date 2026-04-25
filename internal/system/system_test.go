package system

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsRunningInContainer(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The actual result depends on the test environment
	result := IsRunningInContainer()
	t.Logf("IsRunningInContainer returned: %v", result)
}

func TestCurrentUserUID(t *testing.T) {
	uid := CurrentUserUID()
	if uid < 0 {
		t.Errorf("Expected UID to be non-negative, got %d", uid)
	}
	t.Logf("Current user UID: %d", uid)
}

func TestBackupFile(t *testing.T) {
	// Test backup functionality with temp files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")

	// Create test file
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Test backup of existing file
	err := BackupFile(testFile)
	if err != nil {
		t.Errorf("BackupFile failed: %v", err)
	}

	// Check backup exists
	backupFile := testFile + ".bak"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("backup file was not created")
	} else {
		// Verify content
		data, err := os.ReadFile(backupFile)
		if err != nil {
			t.Errorf("failed to read backup file: %v", err)
		}
		if string(data) != string(testContent) {
			t.Error("backup file content mismatch")
		}
	}
}

func TestBackupFileNonExistent(t *testing.T) {
	// Test backup of non-existent file should return nil (no error)
	tmpDir := t.TempDir()
	nonExistentFile := filepath.Join(tmpDir, "nonexistent.txt")

	err := BackupFile(nonExistentFile)
	if err != nil {
		t.Errorf("BackupFile for non-existent file should return nil, got: %v", err)
	}
}

func TestCheckRequirementsPlatform(t *testing.T) {
	// On non-Linux platforms, this should fail (no /etc/os-release)
	err := CheckRequirements()
	if runtime.GOOS != "linux" {
		if err == nil {
			t.Error("Expected CheckRequirements to fail on non-Linux platform")
		}
	}
	// On Linux, result depends on distribution - just log it
	t.Logf("CheckRequirements result: %v", err)
}

func TestIsCIDROccupied(t *testing.T) {
	// Test with various CIDRs - the function should not panic
	// On non-Linux, ip command may not exist, so result will be false
	tests := []string{
		"172.31.0.1/16",
		"10.0.0.1/8",
		"192.168.1.1/24",
	}

	for _, cidr := range tests {
		result := IsCIDROccupied(cidr)
		t.Logf("IsCIDROccupied(%q) = %v", cidr, result)
	}
}

func TestCheckSLES(t *testing.T) {
	// Test that function doesn't panic and returns appropriate error
	err := checkSLES()
	if runtime.GOOS != "linux" {
		if err == nil {
			t.Error("Expected checkSLES to fail on non-Linux")
		}
	}
	t.Logf("checkSLES result: %v", err)
}
