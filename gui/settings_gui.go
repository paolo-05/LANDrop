package gui

import (
	"fmt"
	"lan-drop/config"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func showSettingsWindow(a fyne.App, prefs *config.Preferences, onSave func(port int, folder string)) {
	w := a.NewWindow("Settings")

	portEntry := widget.NewEntry()
	portEntry.SetText(strconv.Itoa(prefs.Port))

	folderLabel := widget.NewLabel(prefs.UploadDir)
	selectFolderBtn := widget.NewButton("Choose Upload Folder", func() {
		dialog.ShowFolderOpen(func(u fyne.ListableURI, err error) {
			if u != nil {
				folderLabel.SetText(u.Path())
			}
		}, w)
	})

	// Shared folder for downloads
	sharedFolderLabel := widget.NewLabel(prefs.SharedDir)
	selectSharedFolderBtn := widget.NewButton("Choose Shared Folder", func() {
		dialog.ShowFolderOpen(func(u fyne.ListableURI, err error) {
			if u != nil {
				sharedFolderLabel.SetText(u.Path())
			}
		}, w)
	})

	saveBtn := widget.NewButton("Save", func() {
		port, err := strconv.Atoi(portEntry.Text)
		if err == nil && port > 0 && port < 65536 {
			prefs.Port = port
			prefs.UploadDir = folderLabel.Text
			prefs.SharedDir = sharedFolderLabel.Text
			config.SavePreferences(a, *prefs)
			onSave(prefs.Port, folderLabel.Text)
			w.Close()
		} else {
			dialog.ShowError(fmt.Errorf("invalid port number"), w)
		}
	})

	showNotifCheckbox := widget.NewCheck("Show upload notifications", func(checked bool) {
		prefs.ShowNotifications = checked
		config.SavePreferences(a, *prefs) // persist change
	})
	showNotifCheckbox.SetChecked(prefs.ShowNotifications)

	autoUpdateCheckbox := widget.NewCheck("Check for updates automatically", func(checked bool) {
		prefs.AutoUpdateCheck = checked
		config.SavePreferences(a, *prefs) // persist change
	})
	autoUpdateCheckbox.SetChecked(prefs.AutoUpdateCheck)

	autoOpenCheckbox := widget.NewCheck("Automatically open uploaded files", func(checked bool) {
		prefs.AutoOpenFiles = checked
		config.SavePreferences(a, *prefs) // persist change
	})
	autoOpenCheckbox.SetChecked(prefs.AutoOpenFiles)

	// Enable downloads checkbox
	enableDownloadsCheckbox := widget.NewCheck("Enable bidirectional transfers (downloads)", func(checked bool) {
		prefs.EnableDownloads = checked
		config.SavePreferences(a, *prefs) // persist change
		// Enable/disable shared folder selection based on this setting
		if checked {
			selectSharedFolderBtn.Enable()
		} else {
			selectSharedFolderBtn.Disable()
		}
	})
	enableDownloadsCheckbox.SetChecked(prefs.EnableDownloads)
	if prefs.EnableDownloads {
		selectSharedFolderBtn.Enable()
	} else {
		selectSharedFolderBtn.Disable()
	}

	w.SetContent(container.NewVBox(
		widget.NewLabelWithStyle("Server Configuration", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("HTTP Port:"),
		portEntry,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Upload Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Upload Folder (where files are saved):"),
		folderLabel,
		selectFolderBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Download Settings", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		enableDownloadsCheckbox,
		widget.NewLabel("Shared Folder (files available for download):"),
		sharedFolderLabel,
		selectSharedFolderBtn,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Notifications", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		showNotifCheckbox,
		autoOpenCheckbox,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Updates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		autoUpdateCheckbox,
		widget.NewSeparator(),
		saveBtn,
	))
	w.Resize(fyne.NewSize(650, 550))
	w.Show()
}
