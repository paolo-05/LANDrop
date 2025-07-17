package config

import (
	"encoding/json"
	"os"
)

type Preferences struct {
	UploadDir         string `json:"upload_dir"`
	Port              int    `json:"port"`
	ShowNotifications bool   `json:"show_notifications"`
}

var ConfigFile = "config.json"

func LoadPreferences() Preferences {
	p := Preferences{
		UploadDir:         "./uploads",
		Port:              8080,
		ShowNotifications: true,
	}
	if _, err := os.Stat(ConfigFile); err == nil {
		data, err := os.ReadFile(ConfigFile)
		if err == nil {
			json.Unmarshal(data, &p)
		}
	}
	return p
}

func SavePreferences(p Preferences) error {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFile, data, 0644)
}

func EnsureUploadDir(p Preferences) {
	os.MkdirAll(p.UploadDir, os.ModePerm)
}
