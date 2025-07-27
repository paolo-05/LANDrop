package p2p

import (
	"lan-drop/config"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // allow all for now
	},
}

func SignalingHandler(w http.ResponseWriter, r *http.Request, prefs *config.Preferences) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		reportStatus("WebRTC connection failed")
		return
	}
	defer ws.Close()

	log.Println("WebSocket connection established")
	reportStatus("WebRTC client connected")

	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			// Check if it's a normal close (client disconnected)
			if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				reportStatus("WebRTC client disconnected")
			} else {
				reportStatus("WebRTC connection error")
			}
			break
		}
		log.Printf("Received signaling message: %s\n", msg)

		HandleSignalMessage(msg, ws, prefs)
	}
}
