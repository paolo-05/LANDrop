package main

import (
	"lan-drop/config"
	"lan-drop/gui"
	"lan-drop/server"
)

func main() {
	prefs := config.LoadPreferences()
	config.EnsureUploadDir(prefs)
	controller := server.NewServerController(prefs.Port, prefs.UploadDir)
	controller.Start()
	gui.Start(prefs, controller)

}
