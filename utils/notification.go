package utils

import (
	"log"
	"path/filepath"

	"fyne.io/fyne/v2"
)

// NotificationConfig holds configuration for sending notifications
type NotificationConfig struct {
	Title    string
	Content  string
	FilePath string // Optional: path to the received file
	Action   string // "open" to open file, "show" to show in file manager
}

// SendNotificationWithAction sends a notification and stores the action for later use
func SendNotificationWithAction(app fyne.App, config NotificationConfig) {
	notification := &fyne.Notification{
		Title:   config.Title,
		Content: config.Content,
	}

	// Send the notification
	app.SendNotification(notification)

	// If there's a file path, we'll handle the action when the user interacts with the app
	// Since Fyne doesn't support click handlers on notifications directly,
	// we'll provide an alternative approach through the app interface
	if config.FilePath != "" {
		// Log the action for potential future handling
		log.Printf("File ready for action: %s (action: %s)", config.FilePath, config.Action)
	}
}

// HandleFileAction handles the action for a received file
func HandleFileAction(filePath string, action string) {
	switch action {
	case "open":
		// Try to open the file with default application (Preview for images/PDFs, etc.)
		if err := OpenFile(filePath); err != nil {
			log.Printf("Failed to open file %s: %v", filePath, err)
			// Fallback to showing in file manager
			if err := ShowInFileManager(filePath); err != nil {
				log.Printf("Failed to show file in file manager %s: %v", filePath, err)
			}
		}
	case "show":
		// Show the file in file manager (Finder/Explorer)
		if err := ShowInFileManager(filePath); err != nil {
			log.Printf("Failed to show file in file manager %s: %v", filePath, err)
		}
	default:
		// Default behavior: try to open file, fallback to showing in file manager
		if err := OpenFile(filePath); err != nil {
			log.Printf("Failed to open file %s: %v", filePath, err)
			if err := ShowInFileManager(filePath); err != nil {
				log.Printf("Failed to show file in file manager %s: %v", filePath, err)
			}
		}
	}
}

// IsImageFile checks if the file is an image that should be opened in Preview/default viewer
func IsImageFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".webp", ".svg"}

	for _, imgExt := range imageExts {
		if ext == imgExt {
			return true
		}
	}
	return false
}

// IsDocumentFile checks if the file is a document that should be opened in default viewer
func IsDocumentFile(filePath string) bool {
	ext := filepath.Ext(filePath)
	docExts := []string{".pdf", ".doc", ".docx", ".txt", ".rtf", ".pages"}

	for _, docExt := range docExts {
		if ext == docExt {
			return true
		}
	}
	return false
}

// GetBestActionForFile determines the best action for a file type
func GetBestActionForFile(filePath string) string {
	if IsImageFile(filePath) || IsDocumentFile(filePath) {
		return "open" // These should open in Preview/default viewer
	}
	return "show" // For other files, show in file manager
}
