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
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func Start(prefs config.Preferences, controller *server.ServerController) {
	a := app.NewWithID("lan-drop")
	a.SetIcon(resourceLogoPng)
	w := a.NewWindow("LAN Drop")

	url := fmt.Sprintf("http://%s:%d", utils.GetLocalIP(), prefs.Port)

	urlLabel := widget.NewLabel("Server running at: " + url)

	qrImg := canvas.NewImageFromImage(qrcode.GenerateQRImage(url))
	qrImg.FillMode = canvas.ImageFillContain
	qrImg.SetMinSize(fyne.NewSize(256, 256))
	qrContainer := container.NewCenter(qrImg)

	statusLabel := widget.NewLabel("")

	controller.OnStatus = func(msg string) {
		fyne.DoAndWait(func() {
			const maxStatusLength = 100
			safeMsg := msg
			if len(safeMsg) > maxStatusLength {
				safeMsg = safeMsg[:maxStatusLength] + "..."
			}
			statusLabel.SetText(safeMsg)
		})
	}

	openBtn := widget.NewButton("Open Uploads Folder", func() {
		go func() {
			if err := utils.OpenFolder(prefs.UploadDir); err != nil {
				log.Println("Error opening folder:", err)
			}
		}()
	})

	settingsBtn := widget.NewButton("Settings", func() {
		showSettingsWindow(a, &prefs, func(port int, folder string) {
			// Called after clicking "Save"
			controller.Update(port, folder)

			url = fmt.Sprintf("http://%s:%d", utils.GetLocalIP(), port)
			urlLabel.SetText("Server running at: " + url)

			qrImg.Image = qrcode.GenerateQRImage(fmt.Sprintf("http://%s:%d", utils.GetLocalIP(), port))
			qrImg.Refresh()
			statusLabel.SetText("Settings saved. Server updated.")
			w.SetTitle("LAN Drop - Port: " + strconv.Itoa(port) + ", Folder: " + folder)
			dialog.ShowInformation("Settings Updated", fmt.Sprintf("Server is now running on port %d and uploads are saved to %s", port, folder), w)
		})
	})

	w.SetContent(container.NewVBox(
		urlLabel,
		qrContainer,
		statusLabel,
		openBtn,
		settingsBtn,
	))

	w.Resize(fyne.NewSize(400, 520))
	w.ShowAndRun()
}
