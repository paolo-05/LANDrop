package update

import (
	"fmt"
	"log"
	"strings"

	"lan-drop/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowUpdateDialog displays a dialog prompting the user to update
func ShowUpdateDialog(app fyne.App, window fyne.Window, updateInfo *UpdateInfo, updateChecker *UpdateChecker) {
	// Prepare the content
	var title string

	if updateInfo.IsMinorUpdate {
		title = "Minor Update Available"
	} else {
		title = "Major Update Available"
	}

	// Create version info
	versionInfo := widget.NewRichTextFromMarkdown(
		"**Current version:** " + updateInfo.CurrentVersion + "\n\n" +
			"**New version:** " + updateInfo.LatestVersion,
	)

	// Create release notes (truncate if too long)
	releaseNotes := updateInfo.ReleaseNotes
	if len(releaseNotes) > 500 {
		releaseNotes = releaseNotes[:500] + "..."
	}

	// Clean up release notes for better display
	releaseNotes = strings.ReplaceAll(releaseNotes, "\r\n", "\n")
	releaseNotes = strings.TrimSpace(releaseNotes)

	var releaseNotesWidget *widget.RichText
	if releaseNotes != "" {
		releaseNotesWidget = widget.NewRichTextFromMarkdown("**What's new:**\n\n" + releaseNotes)
		releaseNotesWidget.Wrapping = fyne.TextWrapWord
	} else {
		releaseNotesWidget = widget.NewRichTextFromMarkdown("**What's new:**\n\nNo release notes available.")
	}

	// Create scrollable content for release notes
	scroll := container.NewScroll(releaseNotesWidget)
	scroll.SetMinSize(fyne.NewSize(400, 150))

	// Create main content
	content := container.NewVBox(
		versionInfo,
		widget.NewSeparator(),
		scroll,
	)

	// Create custom dialog
	d := dialog.NewCustom(title, "Close", content, window)
	d.Resize(fyne.NewSize(500, 350))

	// Add action buttons
	updateButton := widget.NewButton("Download Update", func() {
		d.Hide()
		// Open the GitHub release page
		if err := utils.OpenFile(updateInfo.DownloadURL); err != nil {
			log.Printf("Failed to open download URL: %v", err)
			// Fallback: show URL in a dialog
			urlDialog := dialog.NewInformation("Download URL",
				"Please visit: "+updateInfo.DownloadURL, window)
			urlDialog.Show()
		}
	})
	updateButton.Importance = widget.HighImportance

	skipButton := widget.NewButton("Skip This Version", func() {
		updateChecker.SetSkippedVersion(updateInfo.LatestVersion)
		d.Hide()
	})

	remindButton := widget.NewButton("Remind Me Later", func() {
		d.Hide()
	})

	// Create button container
	buttons := container.NewHBox(
		remindButton,
		skipButton,
		updateButton,
	)

	// Add buttons to the dialog content
	finalContent := container.NewVBox(
		content,
		widget.NewSeparator(),
		buttons,
	)

	// Create new dialog with custom content and no default buttons
	customDialog := dialog.NewCustom(title, "", finalContent, window)
	customDialog.Resize(fyne.NewSize(500, 400))
	customDialog.Show()
}

// ShowUpdateNotification shows a system notification about available update
func ShowUpdateNotification(app fyne.App, updateInfo *UpdateInfo, onAction func()) {
	var title, content string

	if updateInfo.IsMinorUpdate {
		title = "LANDrop - Minor Update Available"
		content = fmt.Sprintf("Version %s is now available (you have %s)",
			updateInfo.LatestVersion, updateInfo.CurrentVersion)
	} else {
		title = "LANDrop - Major Update Available"
		content = fmt.Sprintf("Version %s is now available with important improvements (you have %s)",
			updateInfo.LatestVersion, updateInfo.CurrentVersion)
	}

	// Use our enhanced notification system
	utils.SendNotificationWithAction(app, utils.NotificationConfig{
		Title:    title,
		Content:  content,
		FilePath: "", // No file associated
		Action:   "", // No file action
	})

	// If there's an action callback, we could potentially call it
	// but Fyne notifications don't support click handlers directly
	if onAction != nil {
		log.Printf("Update notification sent. Manual action required.")
	}
}

// CheckAndPromptForUpdates performs the complete update check flow
func CheckAndPromptForUpdates(app fyne.App, window fyne.Window, repoOwner, repoName, currentVersion string, showDialog bool) {
	updateChecker := NewUpdateChecker(repoOwner, repoName, currentVersion, app)

	// Check if we should check for updates
	if !updateChecker.ShouldCheckForUpdates() {
		log.Printf("Update check skipped (recently checked or disabled)")
		return
	}

	// Perform async update check
	updateChecker.CheckForUpdatesAsync(func(updateInfo *UpdateInfo, err error) {
		if err != nil {
			log.Printf("Update check failed: %v", err)
			return
		}

		if updateInfo == nil || !updateInfo.Available {
			log.Printf("No updates available")
			return
		}

		log.Printf("Update available: %s -> %s", updateInfo.CurrentVersion, updateInfo.LatestVersion)

		// Show notification first
		ShowUpdateNotification(app, updateInfo, nil)

		// Show dialog if requested (typically on startup)
		if showDialog && window != nil {
			ShowUpdateDialog(app, window, updateInfo, updateChecker)
		}
	})
}

// ManualUpdateCheck performs a manual update check (usually triggered by user)
func ManualUpdateCheck(app fyne.App, window fyne.Window, repoOwner, repoName, currentVersion string) {
	updateChecker := NewUpdateChecker(repoOwner, repoName, currentVersion, app)

	// Show progress dialog
	progress := dialog.NewProgressInfinite("Checking for updates...", "Please wait", window)
	progress.Show()

	updateChecker.CheckForUpdatesAsync(func(updateInfo *UpdateInfo, err error) {
		progress.Hide()

		if err != nil {
			dialog.ShowError(err, window)
			return
		}

		if updateInfo == nil || !updateInfo.Available {
			dialog.ShowInformation("No Updates", "You are using the latest version of LANDrop.", window)
			return
		}

		ShowUpdateDialog(app, window, updateInfo, updateChecker)
	})
}
