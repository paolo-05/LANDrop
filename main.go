package main

import (
	"embed"
	"lan-drop/config"
	"lan-drop/gui"
	"lan-drop/server"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

// read static files from embedded filesystem
//
//go:embed static/*
var embeddedFiles embed.FS

// embed app icon
//
//go:embed assets/logo.png
var iconResource []byte

func main() {
	// f, err := os.OpenFile("landrop.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatalf("error opening file: %v", err)
	// }
	// log.SetOutput(f)

	// Create the Fyne app first
	a := app.NewWithID("works.bianchessipaolo.landrop")

	// Set app icon using embedded resource
	iconRes := fyne.NewStaticResource("icon.png", iconResource)
	a.SetIcon(iconRes)

	// read version from metadata
	version := a.Metadata().Version
	if version == "" {
		version = "unknown"
	}

	prefs := config.LoadPreferences(a)
	config.EnsureUploadDir(prefs)
	controller := server.NewServerController(prefs.Port, prefs.UploadDir, &prefs, embeddedFiles, version)
	controller.Start()
	gui.Start(a, &prefs, controller, version)
}
