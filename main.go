package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by returning true
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) // Track connected clients
var broadcast = make(chan []byte)            // Broadcast channel

// handleConnections handles incoming WebSocket connections.
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	// Register new client
	clients[ws] = true

	for {
		// Read message from client
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			delete(clients, ws) // Unregister client on error
			break
		}
		// Send message to the broadcast channel
		broadcast <- message
	}
}

// handleMessages listens for messages on the broadcast channel and sends them to all clients.
func handleMessages() {
	for {
		// Get the message from the broadcast channel
		message := <-broadcast

		// Send the message to every client
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println("Error writing message:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {
	http.HandleFunc("/ws", handleConnections)

	// Start the message handling goroutine
	go handleMessages()

	fmt.Println("WebSocket server starting on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
