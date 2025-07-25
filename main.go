package main

import (
	"embed"
	"lan-drop/config"
	"lan-drop/gui"
	"lan-drop/server"

	"fyne.io/fyne/v2/app"
)

// read static files from embedded filesystem
//
//go:embed static/*
var embeddedFiles embed.FS

func main() {
	// Create the Fyne app first
	a := app.NewWithID("lan-drop")

	// read version from metadata
	version := a.Metadata().Version
	if version == "" {
		version = "unknown"
	}

	prefs := config.LoadPreferences(a)
	config.EnsureUploadDir(prefs)
	controller := server.NewServerController(prefs.Port, prefs.UploadDir, &prefs, embeddedFiles, version)
	controller.Start()
	gui.Start(a, prefs, controller, version)
}
