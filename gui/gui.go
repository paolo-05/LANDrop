package gui

import (
	"fmt"
	"image/color"
	"io"
	"lan-drop/config"
	"lan-drop/qrcode"
	"lan-drop/server"
	"lan-drop/update"
	"lan-drop/utils"
	"log"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// copyFileToShared copies a file to the shared directory
func copyFileToShared(sourcePath string, sharedDir string, statusLabel *widget.Label) error {
	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("cannot open source file: %w", err)
	}
	defer sourceFile.Close()

	// Get file info
	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("cannot get file info: %w", err)
	}

	// Skip directories
	if fileInfo.IsDir() {
		return fmt.Errorf("cannot copy directories")
	}

	// Create destination path
	destPath := filepath.Join(sharedDir, filepath.Base(sourcePath))

	// Check if file exists
	if _, err := os.Stat(destPath); err == nil {
		// File exists, add number suffix
		ext := filepath.Ext(filepath.Base(sourcePath))
		name := filepath.Base(sourcePath)
		name = name[:len(name)-len(ext)]

		for i := 1; ; i++ {
			destPath = filepath.Join(sharedDir, fmt.Sprintf("%s_%d%s", name, i, ext))
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				break
			}
		}
	}

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("cannot create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy data
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("cannot copy file data: %w", err)
	}

	// Update status
	if statusLabel != nil {
		fyne.DoAndWait(func() {
			statusLabel.SetText(fmt.Sprintf("Shared: %s", filepath.Base(destPath)))
		})
	}

	return nil
}

func Start(a fyne.App, prefs *config.Preferences, controller *server.ServerController, version string) {
	// Check if onboarding has been completed
	if !prefs.OnboardingCompleted {
		// Show onboarding wizard before creating the main window
		ShowOnboardingWizard(a, prefs, func() {
			// After onboarding completes, reload preferences and continue
			*prefs = config.LoadPreferences(a)
			// Ensure directories exist with new settings
			config.EnsureUploadDir(*prefs)
			config.EnsureSharedDir(*prefs)
			// Update controller with new settings
			controller.Update(prefs.Port, prefs.UploadDir)
		})
	}

	w := a.NewWindow("LAN Drop v" + version)

	url := fmt.Sprintf("http://%s:%d", utils.GetLocalIP(), prefs.Port)

	// Header section with title and version
	titleLabel := widget.NewLabelWithStyle("LANDrop", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	titleLabel.TextStyle.Bold = true

	versionLabel := widget.NewLabelWithStyle("v"+version, fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// Server URL with copy button
	urlLabel := widget.NewLabel("Server URL:")
	copyableURL := widget.NewButton(url, func() {
		a.Clipboard().SetContent(url)
		dialog.ShowInformation("Copied", "URL copied to clipboard", w)
	})
	copyableURL.Importance = widget.LowImportance

	// QR Code
	qrImg := canvas.NewImageFromImage(qrcode.GenerateQRImage(url))
	qrImg.FillMode = canvas.ImageFillContain
	qrImg.SetMinSize(fyne.NewSize(200, 200))
	qrContainer := container.NewCenter(qrImg)

	// Status label (updated dynamically)
	statusLabel := widget.NewLabel("Ready to receive files")
	statusLabel.Wrapping = fyne.TextWrapWord

	controller.OnStatus = func(msg string) {
		fyne.DoAndWait(func() {
			// Limit status message length to avoid UI overflow
			const maxStatusLength = 150
			safeMsg := msg
			if len(safeMsg) > maxStatusLength {
				safeMsg = safeMsg[:maxStatusLength] + "..."
			}
			statusLabel.SetText(safeMsg)
		})
	}

	// Share Files Section (to shared folder for download by peers)
	shareLabel := widget.NewLabelWithStyle("📤 Share Files with Connected Peers", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	shareHint := widget.NewLabel("Select files to add to your shared folder")
	shareHint.Alignment = fyne.TextAlignCenter
	shareHint.TextStyle.Italic = true

	selectFilesBtn := widget.NewButton("Select Files to Share", func() {
		fd := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}
			if reader == nil {
				return
			}
			defer reader.Close()

			// Ensure shared directory exists
			config.EnsureSharedDir(*prefs)

			// Copy file to shared directory
			sourcePath := reader.URI().Path()
			err = copyFileToShared(sourcePath, prefs.SharedDir, statusLabel)
			if err != nil {
				log.Printf("Error sharing file: %v", err)
				dialog.ShowError(fmt.Errorf("failed to share file: %v", err), w)
				return
			}

			dialog.ShowInformation("File Shared",
				fmt.Sprintf("File '%s' is now available for download by connected peers", filepath.Base(sourcePath)), w)
		}, w)
		fd.Show()
	})

	shareArea := container.NewVBox(
		container.NewPadded(shareLabel),
		shareHint,
		selectFilesBtn,
	)

	// Create bordered container for share area
	borderRect := canvas.NewRectangle(color.RGBA{R: 33, G: 147, B: 176, A: 50})
	shareContainer := container.NewStack(
		borderRect,
		container.NewPadded(shareArea),
	)

	// Action buttons section
	openBtn := widget.NewButton("📂 Open Uploads Folder", func() {
		go func() {
			config.EnsureUploadDir(*prefs)
			if err := utils.OpenFolder(prefs.UploadDir); err != nil {
				log.Printf("Error opening folder %s: %v", prefs.UploadDir, err)
				fyne.DoAndWait(func() {
					dialog.ShowError(fmt.Errorf("could not open uploads folder: %s\nError: %v", prefs.UploadDir, err), w)
				})
			}
		}()
	})

	var openSharedBtn *widget.Button
	if prefs.EnableDownloads {
		openSharedBtn = widget.NewButton("📁 Open Shared Folder", func() {
			go func() {
				config.EnsureSharedDir(*prefs)
				if err := utils.OpenFolder(prefs.SharedDir); err != nil {
					log.Printf("Error opening shared folder %s: %v", prefs.SharedDir, err)
					fyne.DoAndWait(func() {
						dialog.ShowError(fmt.Errorf("could not open shared folder: %s\nError: %v", prefs.SharedDir, err), w)
					})
				}
			}()
		})
	}

	settingsBtn := widget.NewButton("⚙️ Settings", func() {
		showSettingsWindow(a, prefs, func(port int, folder string) {
			controller.Update(port, folder)
			url = fmt.Sprintf("http://%s:%d", utils.GetLocalIP(), port)
			copyableURL.SetText(url)
			qrImg.Image = qrcode.GenerateQRImage(url)
			qrImg.Refresh()
			statusLabel.SetText("Settings saved. Server updated.")
			w.SetTitle("LAN Drop v" + version)
			dialog.ShowInformation("Settings Updated",
				fmt.Sprintf("Server is now running on port %d and uploads are saved to %s", port, folder), w)
		})
	})

	updateBtn := widget.NewButton("🔄 Check for Updates", func() {
		update.ManualUpdateCheck(a, w, "paolo-05", "LANDrop", version)
	})

	// Website link
	websiteLink := widget.NewHyperlink("Need help? Visit LANDrop website",
		utils.ParseURL("https://landrop.bianchessipaolo.works"))
	websiteLink.Alignment = fyne.TextAlignCenter

	// Build layout
	topSection := container.NewVBox(
		container.NewCenter(titleLabel),
		container.NewCenter(versionLabel),
		widget.NewSeparator(),
	)

	qrSection := container.NewVBox(
		qrContainer,
		container.NewCenter(urlLabel),
		container.NewCenter(copyableURL),
	)

	shareSection := container.NewVBox(
		widget.NewLabelWithStyle("Share Files", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		shareContainer,
		widget.NewSeparator(),
	)

	statusSection := container.NewVBox(
		widget.NewLabelWithStyle("Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		statusLabel,
		widget.NewSeparator(),
	)

	buttonsSection := container.NewVBox(
		widget.NewLabelWithStyle("Quick Actions", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		openBtn,
	)

	if openSharedBtn != nil {
		buttonsSection.Add(openSharedBtn)
	}

	buttonsSection.Add(widget.NewSeparator())
	buttonsSection.Add(settingsBtn)
	buttonsSection.Add(updateBtn)

	footerSection := container.NewVBox(
		widget.NewSeparator(),
		container.NewCenter(websiteLink),
	)

	// Main layout
	content := container.NewVBox(
		topSection,
		qrSection,
		shareSection,
		statusSection,
		buttonsSection,
		footerSection,
	)

	scrollContent := container.NewVScroll(content)
	scrollContent.SetMinSize(fyne.NewSize(450, 700))

	w.SetContent(scrollContent)
	w.Resize(fyne.NewSize(480, 750))

	// Perform automatic update check on startup (if enabled)
	if prefs.AutoUpdateCheck {
		go func() {
			// Small delay to let the UI fully load
			time.Sleep(2 * time.Second)
			update.CheckAndPromptForUpdates(a, w, "paolo-05", "LANDrop", version, true)
		}()
	}

	w.ShowAndRun()
}
