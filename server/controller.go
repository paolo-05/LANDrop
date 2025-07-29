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

	// Signaling endpoint
	mux.HandleFunc("/signaling", func(w http.ResponseWriter, r *http.Request) {
		p2p.SignalingHandler(w, r, sc.prefs)
	})

	// Version endpoint
	mux.HandleFunc("/version", func(w http.ResponseWriter, r *http.Request) {
		sc.handleVersion(w, r, sc.version)
	})

	addr := fmt.Sprintf(":%d", sc.port)
	sc.server = &http.Server{Addr: addr, Handler: mux}

	go func() {
		if sc.OnStatus != nil {
			sc.OnStatus(fmt.Sprintf("Server listening on %s", addr))
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

			// Automatically perform the action
			utils.HandleFileAction(filePath, action)
		} else {
			// Multiple files - show notification and open upload folder
			utils.SendNotificationWithAction(fyne.CurrentApp(), utils.NotificationConfig{
				Title:    "LAN-Drop",
				Content:  fmt.Sprintf("Received %d files", len(savedFiles)),
				FilePath: sc.prefs.UploadDir,
				Action:   "show",
			})

			// Open the upload folder to show all files
			utils.OpenFolder(sc.prefs.UploadDir)
		}
	}

	w.Write([]byte("Upload successful"))
}

func (c *ServerController) handleVersion(w http.ResponseWriter, _ *http.Request, version string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"version": version,
	})
}
