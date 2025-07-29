package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenFolder(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Test that OpenFolder doesn't panic with valid directory
	// We can't easily test the actual opening behavior in unit tests
	// but we can ensure the function handles the calls without crashing
	err := OpenFolder(tempDir)
	// On CI/testing environments, this might fail due to no GUI, which is expected
	if err != nil {
		t.Logf("OpenFolder failed (expected in headless environment): %v", err)
	}
}

func TestOpenFile(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test that OpenFile doesn't panic with valid file
	err = OpenFile(testFile)
	if err != nil {
		t.Logf("OpenFile failed (expected in headless environment): %v", err)
	}
}

func TestShowInFileManager(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test that ShowInFileManager doesn't panic with valid file
	err = ShowInFileManager(testFile)
	if err != nil {
		t.Logf("ShowInFileManager failed (expected in headless environment): %v", err)
	}
}

func TestFileOperationsWithNonExistentFiles(t *testing.T) {
	nonExistentFile := "/non/existent/file.txt"
	nonExistentDir := "/non/existent/directory"

	// These should either return errors or succeed silently (depending on the OS)
	// We mainly want to ensure they don't panic
	err := OpenFile(nonExistentFile)
	t.Logf("OpenFile with non-existent file returned: %v", err)

	err = OpenFolder(nonExistentDir)
	t.Logf("OpenFolder with non-existent directory returned: %v", err)

	err = ShowInFileManager(nonExistentFile)
	t.Logf("ShowInFileManager with non-existent file returned: %v", err)
}
