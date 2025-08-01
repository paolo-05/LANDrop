package config

import (
	"os"

	"fyne.io/fyne/v2"
)

type Preferences struct {
	UploadDir         string
	Port              int
	ShowNotifications bool
	AutoUpdateCheck   bool
	AutoOpenFiles     bool
}

// LoadPreferences loads preferences using Fyne's preferences API
func LoadPreferences(app fyne.App) Preferences {
	// Default values
	defaultUploadDir := "./uploads"
	defaultPort := 8080
	defaultShowNotifications := true
	defaultAutoUpdateCheck := true
	defaultAutoOpenFiles := true

	// Load from Fyne preferences
	p := Preferences{
		UploadDir:         app.Preferences().StringWithFallback("upload_dir", defaultUploadDir),
		Port:              app.Preferences().IntWithFallback("port", defaultPort),
		ShowNotifications: app.Preferences().BoolWithFallback("show_notifications", defaultShowNotifications),
		AutoUpdateCheck:   app.Preferences().BoolWithFallback("auto_update_check", defaultAutoUpdateCheck),
		AutoOpenFiles:     app.Preferences().BoolWithFallback("auto_open_files", defaultAutoOpenFiles),
	}

	return p
}

// SavePreferences saves preferences using Fyne's preferences API
func SavePreferences(app fyne.App, p Preferences) {
	app.Preferences().SetString("upload_dir", p.UploadDir)
	app.Preferences().SetInt("port", p.Port)
	app.Preferences().SetBool("show_notifications", p.ShowNotifications)
	app.Preferences().SetBool("auto_update_check", p.AutoUpdateCheck)
	app.Preferences().SetBool("auto_open_files", p.AutoOpenFiles)
}

func EnsureUploadDir(p Preferences) {
	os.MkdirAll(p.UploadDir, os.ModePerm)
}
