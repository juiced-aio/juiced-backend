package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"backend.juicedbot.io/juiced.infrastructure/events"

	"flag"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool)
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
var timer *time.Timer

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
	go func() {
		timer = time.NewTimer(999999 * time.Hour)
		<-timer.C
		http.DefaultClient.Get("http://localhost:9999/close")
		fmt.Println("Close")
		os.Exit(0)
	}()
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
	events.GetEventBus().PublishConnectEvent()
	timer.Reset(999999 * time.Hour)
	conn.SetCloseHandler(func(code int, text string) error {
		clients[conn] = false
		// @silent: Here is where you can change the amount of time before it exits
		timer.Reset(5 * time.Minute)
		return nil
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if err.Error() != "websocket: close 1001 (going away)" {
				log.Println("Error receiving message from frontend: " + err.Error())
			}
			// delete(clients, conn)
			break
		} else {
			incomingMessage := IncomingMessage{}
			err = json.Unmarshal(message, &incomingMessage)
			if err != nil {
				log.Println("Error reading message from frontend: " + err.Error())
			}
			// if incomingMessage.EventType == "WalmartEncryptionEvent" {
			// 	taskStore := stores.GetTaskStore()
			// 	taskStore.SetWalmartCardDetails(incomingMessage.TaskID, incomingMessage.CardDetails)
			// }
		}
		// err = conn.WriteMessage(mt, message)
		// if err != nil {
		// 	delete(clients, conn)
		// 	break
		// }
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
