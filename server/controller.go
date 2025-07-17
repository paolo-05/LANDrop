package server

import (
	"context"
	"fmt"
	"io"
	"lan-drop/config"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
)

type ServerController struct {
	mu       sync.Mutex
	server   *http.Server
	port     int
	folder   string
	OnStatus func(string) // GUI callback
}

func NewServerController(port int, folder string) *ServerController {
	return &ServerController{
		port:   port,
		folder: folder,
	}
}

func (sc *ServerController) Start() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.server != nil {
		sc.stopLocked()
	}

	mux := http.NewServeMux()

	// Serve the landing page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "index.html"))
	})

	// Serve all static files (CSS, JS, etc.)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Upload endpoint
	mux.HandleFunc("/upload", sc.handleUpload)

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
	sc.mu.Unlock()
	sc.Start()
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
			sc.OnStatus("Received: " + filepath.Base(savePath))
		}
	}
	if config.LoadPreferences().ShowNotifications {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "LAN-Drop",
			Content: fmt.Sprintf("Received %d file(s)", len(files)),
		})
	}

	w.Write([]byte("Upload successful"))
}
