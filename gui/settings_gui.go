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
	selectFolderBtn := widget.NewButton("Choose Folder", func() {
		dialog.ShowFolderOpen(func(u fyne.ListableURI, err error) {
			if u != nil {
				folderLabel.SetText(u.Path())
			}
		}, w)
	})

	saveBtn := widget.NewButton("Save", func() {
		port, err := strconv.Atoi(portEntry.Text)
		if err == nil && port > 0 && port < 65536 {
			prefs.Port = port
			prefs.UploadDir = folderLabel.Text
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

	w.SetContent(container.NewVBox(
		widget.NewLabel("HTTP Port:"),
		portEntry,
		widget.NewLabel("Upload Folder:"),
		folderLabel,
		selectFolderBtn,
		showNotifCheckbox,
		autoUpdateCheckbox,
		autoOpenCheckbox,
		saveBtn,
	))
	w.Resize(fyne.NewSize(600, 400))
	w.Show()
}
