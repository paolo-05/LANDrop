package gui

import (
	"fmt"
	"lan-drop/config"
	"lan-drop/utils"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// ShowOnboardingWizard displays the initial setup wizard for new users
func ShowOnboardingWizard(a fyne.App, prefs *config.Preferences, onComplete func()) {
	wizardWindow := a.NewWindow("Welcome to LANDrop! 🎉")
	wizardWindow.Resize(fyne.NewSize(650, 550))

	currentStep := 0
	var content *fyne.Container
	var updateContent func()

	// Temporary preferences for the wizard
	tempPort := prefs.Port
	tempUploadDir := prefs.UploadDir
	tempSharedDir := prefs.SharedDir
	tempEnableDownloads := prefs.EnableDownloads
	tempShowNotifications := prefs.ShowNotifications
	tempAutoUpdateCheck := prefs.AutoUpdateCheck

	// Step pages
	steps := []func() *fyne.Container{
		// Step 0: Welcome
		func() *fyne.Container {
			return container.NewVBox(
				widget.NewLabelWithStyle("Welcome to LANDrop! 🎉", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewSeparator(),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("What is LANDrop?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("LANDrop is a simple and secure file transfer tool that works on your local network."),
				widget.NewLabel("Share files between your computer and any device with a web browser - no internet connection needed!"),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("How does it work?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("1. LANDrop runs a local server on your computer"),
				widget.NewLabel("2. Other devices connect to this server using a web browser"),
				widget.NewLabel("3. Transfer files in both directions - upload or download"),
				widget.NewLabel("4. Everything stays on your local network - fast and private"),
				widget.NewLabel(""),
				widget.NewLabel("This quick setup will help you configure LANDrop for your needs."),
				widget.NewLabel("Don't worry - you can change these settings anytime!"),
			)
		},

		// Step 1: Network Configuration
		func() *fyne.Container {
			portEntry := widget.NewEntry()
			portEntry.SetText(strconv.Itoa(tempPort))

			// Get local IP for display
			localIP := utils.GetLocalIP()
			exampleURL := fmt.Sprintf("http://%s:%d", localIP, tempPort)
			urlLabel := widget.NewLabel(exampleURL)
			urlLabel.Wrapping = fyne.TextWrapWord

			portEntry.OnChanged = func(value string) {
				if port, err := strconv.Atoi(value); err == nil && port > 0 && port < 65536 {
					tempPort = port
					urlLabel.SetText(fmt.Sprintf("http://%s:%d", localIP, port))
				}
			}

			return container.NewVBox(
				widget.NewLabelWithStyle("Network Configuration 🌐", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewSeparator(),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("Server Port", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("LANDrop needs to run a web server on your computer."),
				widget.NewLabel("Choose a port number (default is 8080):"),
				widget.NewLabel(""),
				widget.NewLabel("Port:"),
				portEntry,
				widget.NewLabel(""),
				widget.NewLabelWithStyle("Your Server URL", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("Other devices will connect to this address:"),
				urlLabel,
				widget.NewLabel(""),
				widget.NewLabel("💡 Tip: Keep the default port unless you have a conflict."),
				widget.NewLabel("Port numbers 8000-9000 are commonly used for local servers."),
			)
		},

		// Step 2: Folder Configuration
		func() *fyne.Container {
			uploadDirLabel := widget.NewLabel(tempUploadDir)
			uploadDirLabel.Wrapping = fyne.TextWrapWord

			selectUploadBtn := widget.NewButton("Choose Upload Folder", func() {
				dialog.ShowFolderOpen(func(u fyne.ListableURI, err error) {
					if u != nil {
						tempUploadDir = u.Path()
						uploadDirLabel.SetText(tempUploadDir)
					}
				}, wizardWindow)
			})

			sharedDirLabel := widget.NewLabel(tempSharedDir)
			sharedDirLabel.Wrapping = fyne.TextWrapWord

			selectSharedBtn := widget.NewButton("Choose Shared Folder", func() {
				dialog.ShowFolderOpen(func(u fyne.ListableURI, err error) {
					if u != nil {
						tempSharedDir = u.Path()
						sharedDirLabel.SetText(tempSharedDir)
					}
				}, wizardWindow)
			})

			return container.NewVBox(
				widget.NewLabelWithStyle("Folder Configuration 📁", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewSeparator(),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("Upload Folder", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("When other devices send files to you, where should they be saved?"),
				uploadDirLabel,
				selectUploadBtn,
				widget.NewLabel(""),
				widget.NewLabelWithStyle("Shared Folder", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("Which folder should be available for download by other devices?"),
				sharedDirLabel,
				selectSharedBtn,
				widget.NewLabel(""),
				widget.NewLabel("💡 Tip: You can use different folders or the same folder for both."),
			)
		},

		// Step 3: Features Configuration
		func() *fyne.Container {
			downloadsCheck := widget.NewCheck("Enable bidirectional transfers", func(checked bool) {
				tempEnableDownloads = checked
			})
			downloadsCheck.SetChecked(tempEnableDownloads)

			notificationsCheck := widget.NewCheck("Show notifications for uploads", func(checked bool) {
				tempShowNotifications = checked
			})
			notificationsCheck.SetChecked(tempShowNotifications)

			updatesCheck := widget.NewCheck("Check for updates automatically", func(checked bool) {
				tempAutoUpdateCheck = checked
			})
			updatesCheck.SetChecked(tempAutoUpdateCheck)

			return container.NewVBox(
				widget.NewLabelWithStyle("Features & Preferences ⚙️", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewSeparator(),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("File Transfer Options", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				downloadsCheck,
				widget.NewLabel("When enabled, other devices can browse and download files from your shared folder."),
				widget.NewLabel(""),
				widget.NewSeparator(),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("Notifications", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				notificationsCheck,
				widget.NewLabel("Get notified when files are uploaded to your computer."),
				widget.NewLabel(""),
				widget.NewSeparator(),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("Updates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				updatesCheck,
				widget.NewLabel("Automatically check for new versions of LANDrop."),
			)
		},

		// Step 4: Final Summary
		func() *fyne.Container {
			return container.NewVBox(
				widget.NewLabelWithStyle("You're All Set! ✅", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
				widget.NewSeparator(),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("Configuration Summary", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel(fmt.Sprintf("• Server Port: %d", tempPort)),
				widget.NewLabel(fmt.Sprintf("• Upload Folder: %s", tempUploadDir)),
				widget.NewLabel(fmt.Sprintf("• Shared Folder: %s", tempSharedDir)),
				widget.NewLabel(fmt.Sprintf("• Downloads Enabled: %v", tempEnableDownloads)),
				widget.NewLabel(""),
				widget.NewSeparator(),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("Quick Start Guide", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewLabel("1. Click 'Finish' to start LANDrop with your settings"),
				widget.NewLabel("2. A QR code will appear - scan it with your phone or tablet"),
				widget.NewLabel("3. Or manually enter the server URL in any web browser"),
				widget.NewLabel("4. Start transferring files!"),
				widget.NewLabel(""),
				widget.NewLabelWithStyle("Need Help?", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
				widget.NewHyperlink("Visit LANDrop website for documentation",
					utils.ParseURL("https://landrop.bianchessipaolo.works")),
				widget.NewLabel("You can always change these settings later in the Settings menu."),
			)
		},
	}

	// Navigation buttons
	var prevBtn, nextBtn, skipBtn *widget.Button

	// Progress indicator
	stepLabel := widget.NewLabel(fmt.Sprintf("Step %d of %d", currentStep+1, len(steps)))

	updateContent = func() {
		content.Objects = []fyne.CanvasObject{steps[currentStep]()}
		content.Refresh()

		// Update button states
		if currentStep == 0 {
			prevBtn.Disable()
		} else {
			prevBtn.Enable()
		}

		if currentStep == len(steps)-1 {
			nextBtn.SetText("Finish")
		} else {
			nextBtn.SetText("Next")
		}
		stepLabel.SetText(fmt.Sprintf("Step %d of %d", currentStep+1, len(steps)))
	}

	prevBtn = widget.NewButton("← Previous", func() {
		if currentStep > 0 {
			currentStep--
			updateContent()
		}
	})

	nextBtn = widget.NewButton("Next →", func() {
		if currentStep < len(steps)-1 {
			currentStep++
			updateContent()
		} else {
			// Save preferences
			prefs.Port = tempPort
			prefs.UploadDir = tempUploadDir
			prefs.SharedDir = tempSharedDir
			prefs.EnableDownloads = tempEnableDownloads
			prefs.ShowNotifications = tempShowNotifications
			prefs.AutoUpdateCheck = tempAutoUpdateCheck
			prefs.OnboardingCompleted = true

			config.SavePreferences(a, *prefs)
			config.MarkOnboardingCompleted(a)

			wizardWindow.Close()
			if onComplete != nil {
				onComplete()
			}
		}
	})

	skipBtn = widget.NewButton("Skip Setup", func() {
		dialog.ShowConfirm("Skip Setup?",
			"Are you sure you want to skip the setup wizard?\n\nYou can configure LANDrop later in the Settings menu.",
			func(skip bool) {
				if skip {
					prefs.OnboardingCompleted = true
					config.MarkOnboardingCompleted(a)
					wizardWindow.Close()
					if onComplete != nil {
						onComplete()
					}
				}
			}, wizardWindow)
	})

	// Initial content
	content = container.NewVBox(steps[currentStep]())

	// Update initial state
	updateContent()

	// Layout
	navigationBar := container.NewBorder(
		nil, nil,
		prevBtn,
		nextBtn,
		container.NewCenter(stepLabel),
	)

	mainContent := container.NewBorder(
		nil,
		container.NewVBox(
			widget.NewSeparator(),
			navigationBar,
			skipBtn,
		),
		nil, nil,
		container.NewVScroll(content),
	)

	wizardWindow.SetContent(mainContent)
	wizardWindow.CenterOnScreen()
	wizardWindow.Show()
}
