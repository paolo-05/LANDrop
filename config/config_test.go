package config

import (
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestPreferencesDefaults(t *testing.T) {
	// Create a test app
	testApp := test.NewApp()
	defer testApp.Quit()

	// Load preferences (should return defaults)
	prefs := LoadPreferences(testApp)

	// Check default values
	if prefs.UploadDir != "./uploads" {
		t.Errorf("Expected default UploadDir to be './uploads', got '%s'", prefs.UploadDir)
	}

	if prefs.Port != 8080 {
		t.Errorf("Expected default Port to be 8080, got %d", prefs.Port)
	}

	if !prefs.ShowNotifications {
		t.Errorf("Expected default ShowNotifications to be true, got %v", prefs.ShowNotifications)
	}

	if !prefs.AutoUpdateCheck {
		t.Errorf("Expected default AutoUpdateCheck to be true, got %v", prefs.AutoUpdateCheck)
	}

	if !prefs.AutoOpenFiles {
		t.Errorf("Expected default AutoOpenFiles to be true, got %v", prefs.AutoOpenFiles)
	}
}

func TestSaveAndLoadPreferences(t *testing.T) {
	// Create a test app
	testApp := test.NewApp()
	defer testApp.Quit()

	// Create test preferences
	testPrefs := Preferences{
		UploadDir:         "/tmp/test-uploads",
		Port:              9090,
		ShowNotifications: false,
		AutoUpdateCheck:   false,
		AutoOpenFiles:     false,
	}

	// Save preferences
	SavePreferences(testApp, testPrefs)

	// Load preferences back
	loadedPrefs := LoadPreferences(testApp)

	// Verify saved values
	if loadedPrefs.UploadDir != testPrefs.UploadDir {
		t.Errorf("Expected UploadDir '%s', got '%s'", testPrefs.UploadDir, loadedPrefs.UploadDir)
	}

	if loadedPrefs.Port != testPrefs.Port {
		t.Errorf("Expected Port %d, got %d", testPrefs.Port, loadedPrefs.Port)
	}

	if loadedPrefs.ShowNotifications != testPrefs.ShowNotifications {
		t.Errorf("Expected ShowNotifications %v, got %v", testPrefs.ShowNotifications, loadedPrefs.ShowNotifications)
	}

	if loadedPrefs.AutoUpdateCheck != testPrefs.AutoUpdateCheck {
		t.Errorf("Expected AutoUpdateCheck %v, got %v", testPrefs.AutoUpdateCheck, loadedPrefs.AutoUpdateCheck)
	}

	if loadedPrefs.AutoOpenFiles != testPrefs.AutoOpenFiles {
		t.Errorf("Expected AutoOpenFiles %v, got %v", testPrefs.AutoOpenFiles, loadedPrefs.AutoOpenFiles)
	}
}

func TestEnsureUploadDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testUploadDir := filepath.Join(tempDir, "test-uploads", "nested", "path")

	// Create preferences with test directory
	prefs := Preferences{
		UploadDir: testUploadDir,
	}

	// Ensure upload directory exists
	EnsureUploadDir(prefs)

	// Check if directory was created
	if _, err := os.Stat(testUploadDir); os.IsNotExist(err) {
		t.Errorf("Upload directory was not created: %s", testUploadDir)
	}

	// Check if directory is writable
	testFile := filepath.Join(testUploadDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Errorf("Cannot write to upload directory: %v", err)
	}

	// Clean up
	os.Remove(testFile)
}

func TestEnsureUploadDirAlreadyExists(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	testUploadDir := filepath.Join(tempDir, "existing-uploads")

	// Create the directory first
	if err := os.MkdirAll(testUploadDir, os.ModePerm); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create preferences with existing directory
	prefs := Preferences{
		UploadDir: testUploadDir,
	}

	// Ensure upload directory exists (should not fail)
	EnsureUploadDir(prefs)

	// Check if directory still exists
	if _, err := os.Stat(testUploadDir); os.IsNotExist(err) {
		t.Errorf("Upload directory disappeared: %s", testUploadDir)
	}
}
