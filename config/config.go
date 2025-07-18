package config

import (
	"os"

	"fyne.io/fyne/v2"
)

type Preferences struct {
	UploadDir         string
	Port              int
	ShowNotifications bool
}

// LoadPreferences loads preferences using Fyne's preferences API
func LoadPreferences(app fyne.App) Preferences {
	// Default values
	defaultUploadDir := "./uploads"
	defaultPort := 8080
	defaultShowNotifications := true

	// Load from Fyne preferences
	p := Preferences{
		UploadDir:         app.Preferences().StringWithFallback("upload_dir", defaultUploadDir),
		Port:              app.Preferences().IntWithFallback("port", defaultPort),
		ShowNotifications: app.Preferences().BoolWithFallback("show_notifications", defaultShowNotifications),
	}

	return p
}

// SavePreferences saves preferences using Fyne's preferences API
func SavePreferences(app fyne.App, p Preferences) {
	app.Preferences().SetString("upload_dir", p.UploadDir)
	app.Preferences().SetInt("port", p.Port)
	app.Preferences().SetBool("show_notifications", p.ShowNotifications)
}

func EnsureUploadDir(p Preferences) {
	os.MkdirAll(p.UploadDir, os.ModePerm)
}
