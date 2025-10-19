package config

import (
	"os"

	"fyne.io/fyne/v2"
)

type Preferences struct {
	UploadDir           string
	Port                int
	ShowNotifications   bool
	AutoUpdateCheck     bool
	AutoOpenFiles       bool
	EnableDownloads     bool
	SharedDir           string
	OnboardingCompleted bool
}

// LoadPreferences loads preferences using Fyne's preferences API
func LoadPreferences(app fyne.App) Preferences {
	// Default values
	defaultUploadDir := "./uploads"
	defaultPort := 8080
	defaultShowNotifications := true
	defaultAutoUpdateCheck := true
	defaultAutoOpenFiles := true
	defaultEnableDownloads := true
	defaultSharedDir := "./shared"

	// Load from Fyne preferences
	p := Preferences{
		UploadDir:           app.Preferences().StringWithFallback("upload_dir", defaultUploadDir),
		Port:                app.Preferences().IntWithFallback("port", defaultPort),
		ShowNotifications:   app.Preferences().BoolWithFallback("show_notifications", defaultShowNotifications),
		AutoUpdateCheck:     app.Preferences().BoolWithFallback("auto_update_check", defaultAutoUpdateCheck),
		AutoOpenFiles:       app.Preferences().BoolWithFallback("auto_open_files", defaultAutoOpenFiles),
		EnableDownloads:     app.Preferences().BoolWithFallback("enable_downloads", defaultEnableDownloads),
		SharedDir:           app.Preferences().StringWithFallback("shared_dir", defaultSharedDir),
		OnboardingCompleted: app.Preferences().BoolWithFallback("onboarding_completed", false),
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
	app.Preferences().SetBool("enable_downloads", p.EnableDownloads)
	app.Preferences().SetString("shared_dir", p.SharedDir)
	app.Preferences().SetBool("onboarding_completed", p.OnboardingCompleted)
}

// MarkOnboardingCompleted marks the onboarding as completed
func MarkOnboardingCompleted(app fyne.App) {
	app.Preferences().SetBool("onboarding_completed", true)
}

func EnsureUploadDir(p Preferences) {
	os.MkdirAll(p.UploadDir, os.ModePerm)
}

func EnsureSharedDir(p Preferences) {
	os.MkdirAll(p.SharedDir, os.ModePerm)
}
