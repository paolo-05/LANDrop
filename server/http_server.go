package server

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"lan-drop/config"
	"lan-drop/utils"
)

func Start(prefs config.Preferences) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body><form action="/upload" method="POST" enctype="multipart/form-data">
<input type="file" name="file" />
<input type="submit" value="Upload" /></form></body></html>`)
	})

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Error reading file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		filename := header.Filename
		dstPath := filepath.Join(prefs.UploadDir, filename)
		base, ext := filename, ""
		if dot := strings.LastIndex(filename, "."); dot != -1 {
			base = filename[:dot]
			ext = filename[dot:]
		}
		i := 1
		for {
			if _, err := os.Stat(dstPath); os.IsNotExist(err) {
				break
			}
			filename = fmt.Sprintf("%s(%d)%s", base, i, ext)
			dstPath = filepath.Join(prefs.UploadDir, filename)
			i++
		}

		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		io.Copy(dst, file)
		fmt.Fprint(w, "Upload successful")
	})

	addr := fmt.Sprintf(":%d", prefs.Port)
	fmt.Println("Server started at http://" + utils.GetLocalIP() + addr)
	http.ListenAndServe(addr, nil)
}
