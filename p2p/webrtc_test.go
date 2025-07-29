package p2p

import (
	"os"
	"path/filepath"
	"testing"
)

// Mock status reporter for testing
type mockStatusReporter struct {
	messages []string
}

func (m *mockStatusReporter) ReportStatus(message string) {
	m.messages = append(m.messages, message)
}

func (m *mockStatusReporter) getMessages() []string {
	return m.messages
}

func (m *mockStatusReporter) reset() {
	m.messages = nil
}

func TestSetStatusReporter(t *testing.T) {
	// Reset any existing status reporter
	statusReporter = nil

	mock := &mockStatusReporter{}
	SetStatusReporter(mock)

	if statusReporter == nil {
		t.Error("SetStatusReporter did not set the status reporter")
	}

	if statusReporter != mock {
		t.Error("SetStatusReporter did not set the correct status reporter")
	}
}

func TestReportStatusWithReporter(t *testing.T) {
	mock := &mockStatusReporter{}
	SetStatusReporter(mock)

	testMessage := "test status message"
	reportStatus(testMessage)

	messages := mock.getMessages()
	if len(messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(messages))
	}

	if messages[0] != testMessage {
		t.Errorf("Expected '%s', got '%s'", testMessage, messages[0])
	}
}

func TestReportStatusWithoutReporter(t *testing.T) {
	// Set status reporter to nil
	statusReporter = nil

	// This should not panic
	reportStatus("test message")
}

func TestReportStatusMultipleMessages(t *testing.T) {
	mock := &mockStatusReporter{}
	SetStatusReporter(mock)

	messages := []string{"message 1", "message 2", "message 3"}
	for _, msg := range messages {
		reportStatus(msg)
	}

	receivedMessages := mock.getMessages()
	if len(receivedMessages) != len(messages) {
		t.Errorf("Expected %d messages, got %d", len(messages), len(receivedMessages))
	}

	for i, msg := range messages {
		if receivedMessages[i] != msg {
			t.Errorf("Message %d: expected '%s', got '%s'", i, msg, receivedMessages[i])
		}
	}
}

func TestSafeSavePath(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name     string
		filename string
		expected string
	}{
		{"simple filename", "test.txt", "test.txt"},
		{"no extension", "test", "test"},
		{"multiple dots", "test.backup.txt", "test.backup.txt"},
		{"hidden file", ".hidden", ".hidden"},
		{"complex filename", "My Document (1).pdf", "My Document (1).pdf"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := safeSavePath(tempDir, tc.filename)
			expectedPath := filepath.Join(tempDir, tc.expected)

			if result != expectedPath {
				t.Errorf("Expected %s, got %s", expectedPath, result)
			}
		})
	}
}

func TestSafeSevePathWithConflicts(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	originalFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(originalFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// First call should return the _1 version
	result1 := safeSavePath(tempDir, "test.txt")
	expected1 := filepath.Join(tempDir, "test_1.txt")
	if result1 != expected1 {
		t.Errorf("First conflict: expected %s, got %s", expected1, result1)
	}

	// Create the _1 file
	if err := os.WriteFile(expected1, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create _1 file: %v", err)
	}

	// Second call should return the _2 version
	result2 := safeSavePath(tempDir, "test.txt")
	expected2 := filepath.Join(tempDir, "test_2.txt")
	if result2 != expected2 {
		t.Errorf("Second conflict: expected %s, got %s", expected2, result2)
	}
}
