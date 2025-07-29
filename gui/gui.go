package gui

import (
	"fmt"
	"lan-drop/config"
	"lan-drop/qrcode"
	"lan-drop/server"
	"lan-drop/utils"
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func Start(a fyne.App, prefs config.Preferences, controller *server.ServerController, version string) {
	w := a.NewWindow("LAN Drop")

	url := fmt.Sprintf("http://%s:%d", utils.GetLocalIP(), prefs.Port)

	copyableURL := widget.NewButton(url, func() {
		a.Clipboard().SetContent(url)
		dialog.ShowInformation("Copied", "URL copied to clipboard", w)
	})
	copyableURL.Importance = widget.LowImportance // Makes it look more like a label
	// copyableURL.DisableableWidget.BaseWidget // Make it visually static

	// QR Code
	qrImg := canvas.NewImageFromImage(qrcode.GenerateQRImage(url))
	qrImg.FillMode = canvas.ImageFillContain
	qrImg.SetMinSize(fyne.NewSize(256, 256))
	qrContainer := container.NewCenter(qrImg)

	// Status label (updated dynamically)
	statusLabel := widget.NewLabel("")

	controller.OnStatus = func(msg string) {
		fyne.DoAndWait(func() {
			// Limit status message length to avoid UI overflow
			// This is a simple way to ensure the UI remains clean
			const maxStatusLength = 100
			safeMsg := msg
			if len(safeMsg) > maxStatusLength {
				safeMsg = safeMsg[:maxStatusLength] + "..."
			}
			statusLabel.SetText(safeMsg)
		})
	}

	// Buttons
	openBtn := widget.NewButton("Open Uploads Folder", func() {
		go func() {
			if err := utils.OpenFolder(prefs.UploadDir); err != nil {
				log.Println("Error opening folder:", err)
			}
		}()
	})

	settingsBtn := widget.NewButton("Settings", func() {
		showSettingsWindow(a, &prefs, func(port int, folder string) {
			// Update controller
			controller.Update(port, folder)

			// Update GUI
			url = fmt.Sprintf("http://%s:%d", utils.GetLocalIP(), port)
			copyableURL.SetText(url)
			qrImg.Image = qrcode.GenerateQRImage(url)
			qrImg.Refresh()
			statusLabel.SetText("Settings saved. Server updated.")
			w.SetTitle("LAN Drop - Port: " + strconv.Itoa(port) + ", Folder: " + folder)
			dialog.ShowInformation("Settings Updated", fmt.Sprintf("Server is now running on port %d and uploads are saved to %s", port, folder), w)
		})
	})

	// Website link
	websiteLink := widget.NewHyperlink("Having trouble? Visit LAN Drop to gather support", utils.ParseURL("https://landrop.bianchessipaolo.works"))

	versionLabel := widget.NewLabelWithStyle("LAN Drop v"+version, fyne.TextAlignCenter, fyne.TextStyle{Italic: true})

	// Layout
	content := container.NewVBox(
		// widget.NewLabelWithStyle("LAN Drop", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		qrContainer,
		widget.NewLabel("Click below to copy server URL:"),
		copyableURL,
		statusLabel,
		openBtn,
		settingsBtn,
		container.NewCenter(websiteLink),
		versionLabel,
	)

	w.SetContent(content)
	w.Resize(fyne.NewSize(400, 560))
	w.ShowAndRun()
}
