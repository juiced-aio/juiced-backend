package ws

import (
	"juiced/juiced.infrastructure/common/events"
	"os"

	"flag"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

// Message is any WebSocket message
type Message struct {
	Route string                   `json:"route"`
	Data  map[string](interface{}) `json:"data"`
}

// ErrorResponse is sent when something goes wrong
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"errorMessage"`
}

// StartWebsocketServer launches the local server that hosts the WebSocket connection for two-way communication between the app and the backend
func StartWebsocketServer(eventBus *events.EventBus) {
	go ManageEvents(eventBus)
	addr := flag.String("addr", "localhost:8080", "http service address")
	flag.Parse()
	http.HandleFunc("/", HandleConnections)
	http.ListenAndServe(*addr, nil)
}

// HandleConnections handles new WebSocket connections
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	clients[conn] = true

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			delete(clients, conn)
			break
		}
		err = conn.WriteMessage(mt, message)
		if err != nil {
			delete(clients, conn)
			break
		}
	}
}

// ManageEvents manages the events for the EventBus
func ManageEvents(eventBus *events.EventBus) {
	channel := make(chan events.Event)
	eventBus.Subscribe(channel)
	for {
		event := <-channel
		// log.Println("Event received: " + event.EventType)
		// if event.EventType == events.MonitorEventType {
		// 	log.Println("Event info: " + string(event.MonitorEvent.EventType) + ", " + string(event.MonitorEvent.Status))
		// }
		// if event.EventType == events.TaskEventType {
		// 	log.Println("Event info: " + string(event.TaskEvent.EventType) + ", " + string(event.TaskEvent.Status))
		// }
		for client, connected := range clients {
			if connected {
				client.WriteJSON(event)
			}
		}
		if event.EventType == events.CloseEventType {
			os.Exit(0)
		}
	}
}
