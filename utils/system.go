package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

func OpenFolder(path string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", path).Start()
	case "darwin":
		return exec.Command("open", path).Start()
	case "linux":
		return exec.Command("xdg-open", path).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

// OpenFile opens a file with the default application
func OpenFile(filePath string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", filePath).Start()
	case "darwin":
		return exec.Command("open", filePath).Start()
	case "linux":
		return exec.Command("xdg-open", filePath).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

// ShowInFileManager shows the file in the file manager (Finder/Explorer/File Manager)
func ShowInFileManager(filePath string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", "/select,", filepath.Clean(filePath)).Start()
	case "darwin":
		return exec.Command("open", "-R", filePath).Start()
	case "linux":
		// Try to show in file manager, fallback to opening parent directory
		dir := filepath.Dir(filePath)
		return exec.Command("xdg-open", dir).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}
