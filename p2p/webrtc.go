package p2p

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"lan-drop/config"
	"lan-drop/utils"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var peerConnection *webrtc.PeerConnection

// StatusReporter interface for decentralized status updates
type StatusReporter interface {
	ReportStatus(message string)
}

// Global status reporter instance
var statusReporter StatusReporter

// SetStatusReporter sets the global status reporter
func SetStatusReporter(reporter StatusReporter) {
	statusReporter = reporter
}

// reportStatus safely reports status if a reporter is set
func reportStatus(message string) {
	if statusReporter != nil {
		statusReporter.ReportStatus(message)
	}
}

type SignalMessage struct {
	Type      string `json:"type"`
	SDP       string `json:"sdp,omitempty"`
	Candidate string `json:"candidate,omitempty"`
}

// Called when we get an offer from the browser
func HandleSignalMessage(msg []byte, conn *websocket.Conn, prefs *config.Preferences) {
	var signal SignalMessage
	if err := json.Unmarshal(msg, &signal); err != nil {
		dialog.ShowError(errors.New("invalid signaling message"), nil)
		// log.Println("Invalid signaling message:", err)
		return
	}

	switch signal.Type {
	case "offer":
		handleOffer(signal.SDP, conn, prefs)
	case "candidate":
		handleRemoteCandidate(signal.Candidate)
	}
}

func handleOffer(sdp string, conn *websocket.Conn, prefs *config.Preferences) {
	// Create the WebRTC config
	config := webrtc.Configuration{}

	var err error
	peerConnection, err = webrtc.NewPeerConnection(config)
	if err != nil {
		dialog.ShowError(errors.New("failed to create PeerConnection"), nil)
		// log.Println("Failed to create PeerConnection:", err)
		return
	}

	// Setup ICE candidate callback
	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}
		candidateJSON, _ := json.Marshal(SignalMessage{
			Type:      "candidate",
			Candidate: c.ToJSON().Candidate,
		})
		conn.WriteMessage(websocket.TextMessage, candidateJSON)
	})

	// Monitor connection state changes
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("Peer connection state changed: %s\n", s.String())
		switch s {
		case webrtc.PeerConnectionStateConnected:
			reportStatus("WebRTC peer connected")
		case webrtc.PeerConnectionStateDisconnected:
			reportStatus("WebRTC peer disconnected")
		case webrtc.PeerConnectionStateFailed:
			reportStatus("WebRTC connection failed")
		case webrtc.PeerConnectionStateClosed:
			reportStatus("WebRTC connection closed")
		}
	})

	// Create a data channel
	peerConnection.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnOpen(func() {
			reportStatus("Data channel opened")
			log.Println("Data channel opened")
			dc.SendText("Data channel established")
		})

		dc.OnClose(func() {
			reportStatus("Data channel closed")
			log.Println("Data channel closed")
		})

		dc.OnMessage(dcOnMessage(prefs))
	})

	// Set the remote offer
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp,
	}
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		dialog.ShowError(errors.New("failed to set remote description"), nil)
		// log.Println("Failed to set remote description:", err)
		return
	}

	// Create and send the answer
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("Failed to create answer:", err)
		return
	}

	if err := peerConnection.SetLocalDescription(answer); err != nil {
		log.Println("Failed to set local description:", err)
		return
	}

	answerJSON, _ := json.Marshal(SignalMessage{
		Type: "answer",
		SDP:  answer.SDP,
	})
	conn.WriteMessage(websocket.TextMessage, answerJSON)
}

func handleRemoteCandidate(candidateStr string) {
	if peerConnection == nil {
		return
	}
	candidate := webrtc.ICECandidateInit{Candidate: candidateStr}
	peerConnection.AddICECandidate(candidate)
}

var (
	currentFile      *os.File
	currentFileName  string
	expectedFileSize int64
	receivedBytes    int64
	transferSession  *TransferSession
)

// TransferSession tracks a batch of file transfers
type TransferSession struct {
	ID            string
	TotalFiles    int
	ReceivedFiles int
	Files         []string // Track received file paths
	StartTime     time.Time
}

// safeSavePath generates a unique file path to avoid overwriting existing files
func safeSavePath(folder, filename string) string {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	ext := filepath.Ext(filename)
	savePath := filepath.Join(folder, filename)

	i := 1
	for {
		if _, err := os.Stat(savePath); os.IsNotExist(err) {
			break
		}
		savePath = filepath.Join(folder, fmt.Sprintf("%s_%d%s", base, i, ext))
		i++
	}

	return savePath
}

func dcOnMessage(prefs *config.Preferences) func(msg webrtc.DataChannelMessage) {
	return func(msg webrtc.DataChannelMessage) {
		if msg.IsString {
			// Parse message type first
			var msgType struct {
				Type string `json:"type,omitempty"`
			}
			if err := json.Unmarshal(msg.Data, &msgType); err == nil && msgType.Type != "" {
				// Handle session messages
				switch msgType.Type {
				case "session_start":
					var sessionMsg struct {
						Type       string `json:"type"`
						SessionID  string `json:"session_id"`
						TotalFiles int    `json:"total_files"`
					}
					if err := json.Unmarshal(msg.Data, &sessionMsg); err == nil {
						transferSession = &TransferSession{
							ID:            sessionMsg.SessionID,
							TotalFiles:    sessionMsg.TotalFiles,
							ReceivedFiles: 0,
							Files:         make([]string, 0, sessionMsg.TotalFiles),
							StartTime:     time.Now(),
						}
						// Silent session start - no status reporting during auto-upload
					}
					return
				case "session_end":
					if transferSession != nil {
						fileCount := transferSession.TotalFiles

						// Show notification when user confirms upload (clicks Upload button)
						if prefs.ShowNotifications {
							if fileCount == 1 && len(transferSession.Files) > 0 {
								// Single file - show specific file notification and open file
								filePath := transferSession.Files[0]
								action := utils.GetBestActionForFile(filePath)
								utils.SendNotificationWithAction(fyne.CurrentApp(), utils.NotificationConfig{
									Title:    "LAN-Drop",
									Content:  fmt.Sprintf("Received file: %s", filepath.Base(filePath)),
									FilePath: filePath,
									Action:   action,
								})

								// Auto-open single file if enabled
								if prefs.AutoOpenFiles {
									utils.HandleFileAction(filePath, action)
								}
							} else {
								// Multiple files or no files tracked yet - show generic notification and open folder
								utils.SendNotificationWithAction(fyne.CurrentApp(), utils.NotificationConfig{
									Title:    "LAN-Drop",
									Content:  fmt.Sprintf("Received %d files", fileCount),
									FilePath: prefs.UploadDir,
									Action:   "show",
								})

								// Auto-open upload folder if enabled
								if prefs.AutoOpenFiles {
									utils.OpenFolder(prefs.UploadDir)
								}
							}
						}

						transferSession = nil
						if fileCount == 1 {
							reportStatus("File received")
						} else {
							reportStatus(fmt.Sprintf("Received %d files", fileCount))
						}
					}
					return
				}
			}

			// Handle file metadata (existing logic, but track in session)
			var meta struct {
				Name string `json:"name"`
				Size int64  `json:"size"`
			}
			if err := json.Unmarshal(msg.Data, &meta); err != nil {
				log.Println("Failed to parse file metadata:", err)
				return
			}

			// Create file
			savePath := safeSavePath(prefs.UploadDir, meta.Name)
			file, err := os.Create(savePath)
			if err != nil {
				dialog.ShowError(errors.New("failed to create file"), nil)
				// log.Println("Failed to create file:", err)
				return
			}
			currentFile = file
			currentFileName = meta.Name
			expectedFileSize = meta.Size
			receivedBytes = 0

			// Trim filename if longer than 10 characters
			displayName := meta.Name
			if len(displayName) > 10 {
				displayName = displayName[:7] + "..."
			}

			reportStatus(fmt.Sprintf("Receiving: %s", displayName))
		} else {
			// Append chunk to file
			if currentFile == nil {
				dialog.ShowError(errors.New("error retrieving file"), nil)
				// log.Println("Received data before metadata!")
				return
			}

			_, err := currentFile.Write(msg.Data)
			if err != nil {
				dialog.ShowError(errors.New("error retrieving file"), nil)
				// log.Println("Error writing chunk:", err)
				return
			}
			receivedBytes += int64(len(msg.Data))

			if receivedBytes >= expectedFileSize {
				// Trim filename if longer than 10 characters
				displayName := currentFileName
				if len(displayName) > 10 {
					displayName = displayName[:7] + "..."
				}

				filePath := filepath.Join(prefs.UploadDir, currentFileName)

				// log.Printf("✅ File %s received completely (%d bytes)\n", currentFileName, receivedBytes)
				reportStatus(fmt.Sprintf("Received: %s", displayName))

				// Track file in transfer session
				if transferSession != nil {
					transferSession.Files = append(transferSession.Files, filePath)
					transferSession.ReceivedFiles++

					// NO individual file notifications during auto-upload
					// Notifications only happen on session_end when user clicks upload
				} else {
					// Legacy mode - single file without session (also no notification during auto-upload)
					// User will get notification only when they click upload button
				}

				currentFile.Close()
				currentFile = nil
			}
		}
	}
}
