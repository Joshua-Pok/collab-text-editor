package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"strings"
	"sync"
)

type UpdateMessage struct {
	ClientID  string `json:"clientId"`
	UpdateB64 string `json:"updateB64"` //CDRT binary
	TS        int64  `json:"ts"`
}

func getEnv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}

	return v
}

func parseDocID(topic string) (string, bool) {
	//expected: doc/{docID}/updates

	parts := strings.Split(topic, "/")

	if len(parts) != 3 {
		return "", false
	}

	if parts[0] != "doc" || parts[2] != "updates" {
		return "", false
	}

	if parts[1] == "" {
		return "", false
	}

	return parts[1], true
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]bool) //client: connected?
var broadcast = make(chan []byte)            //broadcast channel
var mutex = &sync.Mutex{}                    // protect clients map

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil) // upgrade http request to ws
	if err != nil {
		fmt.Println("Unable to upgrade http connection", err)
		return

	}

	defer conn.Close()

	mutex.Lock() //any other go routine that uses this mutex must wait here
	clients[conn] = true
	mutex.Unlock()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			mutex.Lock()
			delete(clients, conn)
			mutex.Unlock()
			break
		}
		broadcast <- message // put message in the broadcast channel
	}
}

func handleBroadcast() {
	for {
		message := <-broadcast //grab next message from broadcast

		mutex.Lock()
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
		mutex.Unlock()
	}
}
func main() {

	http.HandleFunc("/ws", wsHandler)
	go handleBroadcast()
	fmt.Println("Websocket server started on port:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting websocket server")
	}

}
