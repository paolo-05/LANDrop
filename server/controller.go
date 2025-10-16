package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"lan-drop/config"
	"lan-drop/p2p"
	"lan-drop/utils"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
)

type ServerController struct {
	mu            sync.Mutex
	server        *http.Server
	port          int
	folder        string
	prefs         *config.Preferences // Add preferences to the controller
	embeddedFiles embed.FS            // Embedded filesystem for static files
	version       string              // Version of the application
	OnStatus      func(string)        // GUI callback
}

func NewServerController(port int, folder string, prefs *config.Preferences, embeddedFiles embed.FS, version string) *ServerController {
	return &ServerController{
		port:          port,
		folder:        folder,
		prefs:         prefs,
		embeddedFiles: embeddedFiles,
		version:       version,
	}
}

func (sc *ServerController) Start() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.server != nil {
		sc.stopLocked()
	}

	// Set up the status reporter for P2P
	p2p.SetStatusReporter(sc)

	// Create a new HTTP server
	mux := http.NewServeMux()

	// Use the embedded filesystem for static files
	content, err := fs.Sub(sc.embeddedFiles, "static")
	if err != nil {
		fmt.Println("Error creating embedded filesystem:", err)
		return
	}

	// Update handlers to use embedded content
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data, err := fs.ReadFile(content, "index.html")
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(data)
	})

	// Replace the static file server with one that serves from embedded files
	mux.Handle("/static/", http.FileServer(http.FS(content)))

	// Upload endpoint
	mux.HandleFunc("/upload", sc.handleUpload)

	// Delete endpoint for file cleanup
	mux.HandleFunc("/delete", sc.handleDelete)

	// Signaling endpoint
	mux.HandleFunc("/signaling", func(w http.ResponseWriter, r *http.Request) {
		p2p.SignalingHandler(w, r, sc.prefs)
	})

	// Version endpoint
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		sc.handleVersion(w, r, sc.version)
	})

	// File browsing and download endpoints (for bidirectional transfers)
	mux.HandleFunc("/files", sc.handleFileBrowse)
	mux.HandleFunc("/download", sc.handleFileDownload)

	addr := fmt.Sprintf(":%d", sc.port)
	sc.server = &http.Server{Addr: addr, Handler: mux}

	go func() {
		if sc.OnStatus != nil {
			sc.OnStatus(fmt.Sprintf("Server listening on port %d", sc.port))
		}
		err := sc.server.ListenAndServe()
		if err != nil && sc.OnStatus != nil {
			sc.OnStatus(fmt.Sprintf("Server stopped: %s", err))
		}
	}()
}

func (sc *ServerController) Stop() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.stopLocked()
}

func (sc *ServerController) stopLocked() {
	if sc.server != nil {
		_ = sc.server.Shutdown(context.Background())
		sc.server = nil
	}
}

func (sc *ServerController) Update(port int, folder string) {
	sc.mu.Lock()
	sc.port = port
	sc.folder = folder
	// Update the preferences as well
	sc.prefs.Port = port
	sc.prefs.UploadDir = folder
	sc.mu.Unlock()
	sc.Start()
}

// ReportStatus implements the p2p.StatusReporter interface
func (sc *ServerController) ReportStatus(message string) {
	if sc.OnStatus != nil {
		sc.OnStatus(message)
	}
}

func (sc *ServerController) safeSavePath(filename string) string {
	dir := sc.folder
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	ext := filepath.Ext(filename)
	savePath := filepath.Join(dir, filename)

	i := 1
	for {
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			break
		}
		savePath = filepath.Join(dir, fmt.Sprintf("%s_%d%s", base, i, ext))
		i++
	}

	return savePath
}
func (sc *ServerController) handleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20) // 32 MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	noErrCount := 0
	var savedFiles []string

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		savePath := sc.safeSavePath(fileHeader.Filename)

		out, err := os.Create(savePath)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		io.Copy(out, file)

		if sc.OnStatus != nil {
			noErrCount++
			sc.OnStatus("Received: " + filepath.Base(savePath))
		}
		savedFiles = append(savedFiles, savePath)
	}
	sc.OnStatus(fmt.Sprintf("Received %d file(s)", noErrCount))

	if sc.prefs.ShowNotifications {
		if len(savedFiles) == 1 {
			// Single file - use enhanced notification with file action
			filePath := savedFiles[0]
			action := utils.GetBestActionForFile(filePath)

			utils.SendNotificationWithAction(fyne.CurrentApp(), utils.NotificationConfig{
				Title:    "LAN-Drop",
				Content:  fmt.Sprintf("Received file: %s", filepath.Base(filePath)),
				FilePath: filePath,
				Action:   action,
			})

			// Automatically perform the action only if enabled
			if sc.prefs.AutoOpenFiles {
				utils.HandleFileAction(filePath, action)
			}
		} else {
			// Multiple files - show notification and open upload folder
			utils.SendNotificationWithAction(fyne.CurrentApp(), utils.NotificationConfig{
				Title:    "LAN-Drop",
				Content:  fmt.Sprintf("Received %d files", len(savedFiles)),
				FilePath: sc.prefs.UploadDir,
				Action:   "show",
			})

			// Open the upload folder to show all files only if enabled
			if sc.prefs.AutoOpenFiles {
				if err := utils.OpenFolder(sc.prefs.UploadDir); err != nil {
					log.Printf("Failed to auto-open upload folder: %v", err)
				}
			}
		}
	}

	w.Write([]byte("Upload successful"))
}

// handleDelete removes uploaded files that were not confirmed
func (sc *ServerController) handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := r.FormValue("filename")
	if filename == "" {
		http.Error(w, "Filename required", http.StatusBadRequest)
		return
	}

	// Construct file path
	filePath := filepath.Join(sc.prefs.UploadDir, filename)

	// Security check: ensure the file is within the upload directory
	uploadDir, err := filepath.Abs(sc.prefs.UploadDir)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	targetFile, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if !strings.HasPrefix(targetFile, uploadDir) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Delete the file
	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		}
		return
	}

	w.Write([]byte("File deleted successfully"))
}

func (c *ServerController) handleVersion(w http.ResponseWriter, _ *http.Request, version string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"version": version,
	})
}

// FileInfo represents a file available for download
type FileInfo struct {
	Name         string `json:"name"`
	Size         int64  `json:"size"`
	ModTime      string `json:"modTime"`
	IsDirectory  bool   `json:"isDirectory"`
	RelativePath string `json:"relativePath"`
}

// handleFileBrowse lists files available for download
func (sc *ServerController) handleFileBrowse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if downloads are enabled
	if !sc.prefs.EnableDownloads {
		http.Error(w, "Downloads not enabled", http.StatusForbidden)
		return
	}

	// Get the requested path (subdirectory within shared folder)
	requestedPath := r.URL.Query().Get("path")
	if requestedPath == "" {
		requestedPath = "."
	}

	// Ensure shared directory exists
	config.EnsureSharedDir(*sc.prefs)

	// Construct the full path and validate it's within shared directory
	fullPath := filepath.Join(sc.prefs.SharedDir, requestedPath)
	sharedDirAbs, err := filepath.Abs(sc.prefs.SharedDir)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fullPathAbs, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	// Security check: ensure we're within the shared directory
	if !strings.HasPrefix(fullPathAbs, sharedDirAbs) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check if path exists and is a directory
	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Path not found", http.StatusNotFound)
		} else {
			http.Error(w, "Cannot access path", http.StatusInternalServerError)
		}
		return
	}

	var files []FileInfo

	if stat.IsDir() {
		// List directory contents
		entries, err := os.ReadDir(fullPath)
		if err != nil {
			http.Error(w, "Cannot read directory", http.StatusInternalServerError)
			return
		}

		for _, entry := range entries {
			// Skip .DS_Store files
			if entry.Name() == ".DS_Store" {
				continue
			}

			info, err := entry.Info()
			if err != nil {
				continue // Skip files we can't read
			}

			relativePath := requestedPath
			if relativePath != "." {
				relativePath = filepath.Join(requestedPath, entry.Name())
			} else {
				relativePath = entry.Name()
			}

			files = append(files, FileInfo{
				Name:         entry.Name(),
				Size:         info.Size(),
				ModTime:      info.ModTime().Format("2006-01-02 15:04:05"),
				IsDirectory:  entry.IsDir(),
				RelativePath: relativePath,
			})
		}
	} else {
		// Single file info
		files = append(files, FileInfo{
			Name:         stat.Name(),
			Size:         stat.Size(),
			ModTime:      stat.ModTime().Format("2006-01-02 15:04:05"),
			IsDirectory:  false,
			RelativePath: requestedPath,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"path":  requestedPath,
		"files": files,
	})
}

// handleFileDownload serves files for download
func (sc *ServerController) handleFileDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if downloads are enabled
	if !sc.prefs.EnableDownloads {
		http.Error(w, "Downloads not enabled", http.StatusForbidden)
		return
	}

	// Get the requested file path
	filePath := r.URL.Query().Get("file")
	if filePath == "" {
		http.Error(w, "File parameter required", http.StatusBadRequest)
		return
	}

	// Construct the full path and validate it's within shared directory
	fullPath := filepath.Join(sc.prefs.SharedDir, filePath)
	sharedDirAbs, err := filepath.Abs(sc.prefs.SharedDir)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	fullPathAbs, err := filepath.Abs(fullPath)
	if err != nil {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Security check: ensure we're within the shared directory
	if !strings.HasPrefix(fullPathAbs, sharedDirAbs) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	// Check if file exists and is not a directory
	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Cannot access file", http.StatusInternalServerError)
		}
		return
	}

	if stat.IsDir() {
		http.Error(w, "Cannot download directory", http.StatusBadRequest)
		return
	}

	// Open the file
	file, err := os.Open(fullPath)
	if err != nil {
		http.Error(w, "Cannot open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set appropriate headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(fullPath)))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))

	// Report status
	if sc.OnStatus != nil {
		sc.OnStatus(fmt.Sprintf("Downloading: %s", filepath.Base(fullPath)))
	}

	// Stream the file
	io.Copy(w, file)

	// Report completion
	if sc.OnStatus != nil {
		sc.OnStatus(fmt.Sprintf("Downloaded: %s", filepath.Base(fullPath)))
	}
}
