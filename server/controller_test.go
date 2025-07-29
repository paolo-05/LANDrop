package server

import (
	"bytes"
	"embed"
	"encoding/json"
	"lan-drop/config"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Create an empty embedded filesystem for testing
var testEmbeddedFiles embed.FS

// Mock status reporter for testing
type mockStatusReporter struct {
	messages []string
}

func (m *mockStatusReporter) recordMessage(msg string) {
	m.messages = append(m.messages, msg)
}

func (m *mockStatusReporter) getMessages() []string {
	return m.messages
}

func TestNewServerController(t *testing.T) {
	tempDir := t.TempDir()
	prefs := &config.Preferences{
		UploadDir:         tempDir,
		Port:              8080,
		ShowNotifications: true,
		AutoUpdateCheck:   true,
		AutoOpenFiles:     true,
	}

	controller := NewServerController(8080, tempDir, prefs, testEmbeddedFiles, "test-version")

	if controller == nil {
		t.Fatal("NewServerController returned nil")
	}

	if controller.port != 8080 {
		t.Errorf("Expected port 8080, got %d", controller.port)
	}

	if controller.folder != tempDir {
		t.Errorf("Expected folder %s, got %s", tempDir, controller.folder)
	}

	if controller.version != "test-version" {
		t.Errorf("Expected version 'test-version', got '%s'", controller.version)
	}

	if controller.prefs != prefs {
		t.Error("Preferences not set correctly")
	}
}

func TestServerControllerReportStatus(t *testing.T) {
	tempDir := t.TempDir()
	prefs := &config.Preferences{
		UploadDir:         tempDir,
		Port:              8080,
		ShowNotifications: true,
		AutoUpdateCheck:   true,
		AutoOpenFiles:     true,
	}

	controller := NewServerController(8080, tempDir, prefs, testEmbeddedFiles, "test-version")

	// Test with no callback set
	controller.ReportStatus("test message")

	// Test with callback set
	var receivedMessage string
	controller.OnStatus = func(msg string) {
		receivedMessage = msg
	}

	controller.ReportStatus("test message 2")
	if receivedMessage != "test message 2" {
		t.Errorf("Expected 'test message 2', got '%s'", receivedMessage)
	}
}

func TestServerControllerUpdate(t *testing.T) {
	tempDir := t.TempDir()
	prefs := &config.Preferences{
		UploadDir:         tempDir,
		Port:              8080,
		ShowNotifications: true,
		AutoUpdateCheck:   true,
		AutoOpenFiles:     true,
	}

	controller := NewServerController(8080, tempDir, prefs, testEmbeddedFiles, "test-version")

	newTempDir := t.TempDir()
	controller.Update(9090, newTempDir)

	if controller.port != 9090 {
		t.Errorf("Expected port 9090 after update, got %d", controller.port)
	}

	if controller.folder != newTempDir {
		t.Errorf("Expected folder %s after update, got %s", newTempDir, controller.folder)
	}

	if controller.prefs.Port != 9090 {
		t.Errorf("Expected preferences port 9090 after update, got %d", controller.prefs.Port)
	}

	if controller.prefs.UploadDir != newTempDir {
		t.Errorf("Expected preferences folder %s after update, got %s", newTempDir, controller.prefs.UploadDir)
	}
}

func TestSafeSavePath(t *testing.T) {
	tempDir := t.TempDir()
	prefs := &config.Preferences{
		UploadDir:         tempDir,
		Port:              8080,
		ShowNotifications: true,
		AutoUpdateCheck:   true,
		AutoOpenFiles:     true,
	}

	controller := NewServerController(8080, tempDir, prefs, testEmbeddedFiles, "test-version")

	// Test with non-existing file
	path1 := controller.safeSavePath("test.txt")
	expected1 := filepath.Join(tempDir, "test.txt")
	if path1 != expected1 {
		t.Errorf("Expected %s, got %s", expected1, path1)
	}

	// Create the file to test conflict resolution
	os.WriteFile(expected1, []byte("test"), 0644)

	// Test with existing file (should add _1)
	path2 := controller.safeSavePath("test.txt")
	expected2 := filepath.Join(tempDir, "test_1.txt")
	if path2 != expected2 {
		t.Errorf("Expected %s, got %s", expected2, path2)
	}

	// Create the _1 file too
	os.WriteFile(expected2, []byte("test"), 0644)

	// Test with both existing (should add _2)
	path3 := controller.safeSavePath("test.txt")
	expected3 := filepath.Join(tempDir, "test_2.txt")
	if path3 != expected3 {
		t.Errorf("Expected %s, got %s", expected3, path3)
	}
}

func TestSafeSavePathWithExtension(t *testing.T) {
	tempDir := t.TempDir()
	prefs := &config.Preferences{
		UploadDir:         tempDir,
		Port:              8080,
		ShowNotifications: true,
		AutoUpdateCheck:   true,
		AutoOpenFiles:     true,
	}

	controller := NewServerController(8080, tempDir, prefs, testEmbeddedFiles, "test-version")

	// Create original file
	originalPath := filepath.Join(tempDir, "document.pdf")
	os.WriteFile(originalPath, []byte("test"), 0644)

	// Test with existing file with extension
	path := controller.safeSavePath("document.pdf")
	expected := filepath.Join(tempDir, "document_1.pdf")
	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

func TestHandleVersion(t *testing.T) {
	tempDir := t.TempDir()
	prefs := &config.Preferences{
		UploadDir:         tempDir,
		Port:              8080,
		ShowNotifications: true,
		AutoUpdateCheck:   true,
		AutoOpenFiles:     true,
	}

	controller := NewServerController(8080, tempDir, prefs, testEmbeddedFiles, "1.2.3")

	req := httptest.NewRequest("GET", "/version", nil)
	w := httptest.NewRecorder()

	controller.handleVersion(w, req, controller.version)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	if response["version"] != "1.2.3" {
		t.Errorf("Expected version '1.2.3', got '%s'", response["version"])
	}
}

func TestHandleUploadPOST(t *testing.T) {
	tempDir := t.TempDir()
	prefs := &config.Preferences{
		UploadDir:         tempDir,
		Port:              8080,
		ShowNotifications: false, // Disable notifications for testing
	}

	controller := NewServerController(8080, tempDir, prefs, testEmbeddedFiles, "test-version")

	// Capture status messages
	var statusMessages []string
	controller.OnStatus = func(msg string) {
		statusMessages = append(statusMessages, msg)
	}

	// Create a multipart form with a file
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Add a file
	fileWriter, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	fileWriter.Write([]byte("Hello, World!"))

	writer.Close()

	req := httptest.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	controller.handleUpload(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	responseBody := w.Body.String()
	if responseBody != "Upload successful" {
		t.Errorf("Expected 'Upload successful', got '%s'", responseBody)
	}

	// Check if file was saved
	uploadedFile := filepath.Join(tempDir, "test.txt")
	if _, err := os.Stat(uploadedFile); os.IsNotExist(err) {
		t.Error("Uploaded file was not saved")
	}

	// Check file content
	content, err := os.ReadFile(uploadedFile)
	if err != nil {
		t.Fatalf("Failed to read uploaded file: %v", err)
	}
	if string(content) != "Hello, World!" {
		t.Errorf("Expected file content 'Hello, World!', got '%s'", string(content))
	}

	// Check status messages
	if len(statusMessages) < 2 {
		t.Errorf("Expected at least 2 status messages, got %d", len(statusMessages))
	}
}

func TestHandleUploadWrongMethod(t *testing.T) {
	tempDir := t.TempDir()
	prefs := &config.Preferences{
		UploadDir:         tempDir,
		Port:              8080,
		ShowNotifications: true,
		AutoUpdateCheck:   true,
		AutoOpenFiles:     true,
	}

	controller := NewServerController(8080, tempDir, prefs, testEmbeddedFiles, "test-version")

	req := httptest.NewRequest("GET", "/upload", nil)
	w := httptest.NewRecorder()

	controller.handleUpload(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Code)
	}
}

func TestHandleUploadNoFiles(t *testing.T) {
	tempDir := t.TempDir()
	prefs := &config.Preferences{
		UploadDir:         tempDir,
		Port:              8080,
		ShowNotifications: true,
		AutoUpdateCheck:   true,
		AutoOpenFiles:     true,
	}

	controller := NewServerController(8080, tempDir, prefs, testEmbeddedFiles, "test-version")

	// Create empty multipart form
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	controller.handleUpload(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
