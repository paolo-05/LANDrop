package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func OpenFolder(path string) error {
	// Validate that the directory exists
	if stat, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("directory does not exist: %s", path)
		}
		return fmt.Errorf("cannot access directory %s: %v", path, err)
	} else if !stat.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Convert to absolute path for better platform compatibility
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("cannot resolve absolute path for %s: %v", path, err)
	}

	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer", absPath).Start()
	case "darwin":
		return exec.Command("open", absPath).Start()
	case "linux":
		return exec.Command("xdg-open", absPath).Start()
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
