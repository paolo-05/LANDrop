package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsImageFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.jpg", true},
		{"test.jpeg", true},
		{"test.png", true},
		{"test.gif", true},
		{"test.bmp", true},
		{"test.tiff", true},
		{"test.webp", true},
		{"test.svg", true},
		{"test.txt", false},
		{"test.pdf", false},
		{"test.doc", false},
		{"test", false},
	}

	for _, test := range tests {
		result := IsImageFile(test.filename)
		if result != test.expected {
			t.Errorf("IsImageFile(%s) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}

func TestIsDocumentFile(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"test.pdf", true},
		{"test.doc", true},
		{"test.docx", true},
		{"test.txt", true},
		{"test.rtf", true},
		{"test.pages", true},
		{"test.jpg", false},
		{"test.png", false},
		{"test.mp4", false},
		{"test", false},
	}

	for _, test := range tests {
		result := IsDocumentFile(test.filename)
		if result != test.expected {
			t.Errorf("IsDocumentFile(%s) = %v, expected %v", test.filename, result, test.expected)
		}
	}
}

func TestGetBestActionForFile(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"test.jpg", "open"},
		{"test.pdf", "open"},
		{"test.png", "open"},
		{"test.doc", "open"},
		{"document.txt", "open"},
		{"video.mp4", "show"},
		{"archive.zip", "show"},
		{"executable.exe", "show"},
		{"unknown", "show"},
	}

	for _, test := range tests {
		result := GetBestActionForFile(test.filename)
		if result != test.expected {
			t.Errorf("GetBestActionForFile(%s) = %s, expected %s", test.filename, result, test.expected)
		}
	}
}

func TestHandleFileAction(t *testing.T) {
	// Create a temporary file for testing
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")

	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test that HandleFileAction doesn't panic with valid file
	// We can't easily test the actual opening behavior in unit tests
	// but we can ensure the function handles the calls without crashing
	HandleFileAction(testFile, "open")
	HandleFileAction(testFile, "show")
	HandleFileAction(testFile, "invalid_action")

	// Test with non-existent file (should not panic)
	HandleFileAction("/non/existent/file.txt", "open")
}
