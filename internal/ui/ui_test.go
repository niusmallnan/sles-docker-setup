package ui

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"testing"
)

// captureOutput captures stdout during test
func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintBanner(t *testing.T) {
	output := captureOutput(func() {
		PrintBanner("1.0.0")
	})

	// Verify output contains expected content
	if !contains(output, "Docker Pilot") {
		t.Error("PrintBanner should contain 'Docker Pilot'")
	}

	if !contains(output, "1.0.0") {
		t.Error("PrintBanner should contain version")
	}

	if !contains(output, "SLES 15+") {
		t.Error("PrintBanner should mention SLES 15+")
	}

	t.Logf("Banner output:\n%s", output)
}

func TestPrintStep(t *testing.T) {
	output := captureOutput(func() {
		PrintStep(1, 5, "Test Step")
	})

	if !contains(output, "[1/5]") {
		t.Error("PrintStep should contain step indicator")
	}

	if !contains(output, "Test Step") {
		t.Error("PrintStep should contain step title")
	}
}

func TestPrintSuccess(t *testing.T) {
	output := captureOutput(func() {
		PrintSuccess("Operation completed: %s", "test")
	})

	if !contains(output, "Operation completed: test") {
		t.Error("PrintSuccess should contain formatted message")
	}
}

func TestPrintWarning(t *testing.T) {
	output := captureOutput(func() {
		PrintWarning("Warning: %d items", 5)
	})

	if !contains(output, "Warning: 5 items") {
		t.Error("PrintWarning should contain formatted message")
	}
}

func TestPrintError(t *testing.T) {
	output := captureOutput(func() {
		PrintError("Error code: %d", 404)
	})

	if !contains(output, "Error code: 404") {
		t.Error("PrintError should contain formatted message")
	}
}

func TestPrintInfo(t *testing.T) {
	output := captureOutput(func() {
		PrintInfo("Info: %s", "test message")
	})

	if !contains(output, "Info: test message") {
		t.Error("PrintInfo should contain formatted message")
	}
}

func TestPrintCompletion(t *testing.T) {
	output := captureOutput(func() {
		PrintCompletion()
	})

	expectedPhrases := []string{
		"All configurations completed",
		"newgrp docker",
		"hello-world",
		"daemon.json",
		"http-proxy.conf",
	}

	for _, phrase := range expectedPhrases {
		if !contains(output, phrase) {
			t.Errorf("PrintCompletion should contain '%s'", phrase)
		}
	}
}

// contains is a helper to check if string contains substring,
// ignoring ANSI color codes
func contains(s, substr string) bool {
	// Strip ANSI color codes for comparison
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	stripped := re.ReplaceAllString(s, "")
	return len(stripped) >= len(substr) && containsHelper(stripped, substr)
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
