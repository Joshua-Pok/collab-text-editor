package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

type UpdateMessage struct {
	ClientID  string `json:"clientId"`
	UpdateB64 string `json:"updateB64"` //CDRT binary
	TS        int64  `json:"ts"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	clients   = make(map[string]*Client)
	mutex     = &sync.Mutex{}
	broadcast = make(chan *UpdateMessage)
)

type Client struct {
	conn     *websocket.Conn
	clientID string
	send     chan []byte //each client needs a channel for fan out pattenr, if not each message would only go to one client
}

func (c *Client) readPump() {
	defer func() { //cleanup function
		mutex.Lock()
		delete(clients, c.clientID)
		mutex.Unlock()
		c.conn.Close()
		fmt.Printf("Client %s disconnected!", c.clientID)
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("Websocket error: %v\n", err)
			}
			break
		}

		var updateMsg UpdateMessage
		if err := json.Unmarshal(message, &updateMsg); err != nil {
			fmt.Printf("Error unmarshalling message: %v\n", err)
			continue
		}

		broadcast <- &updateMsg

	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for message := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Printf("Error writing message: %v\n", err)
			return
		}
	}

}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil) // upgrade http request to ws
	if err != nil {
		fmt.Println("Unable to upgrade http connection", err)
		return

	}

	defer conn.Close()

	clientID := r.URL.Query().Get("clientId") //request still comes in as http first before being upgraded so we can still get stuff from the url
	if clientID == "" {
		fmt.Println("Client has no clientid")
		conn.Close()
		return
	}

	client := &Client{
		conn:     conn,
		clientID: clientID,
		send:     make(chan []byte, 256),
	}

	mutex.Lock()
	clients[clientID] = client
	mutex.Unlock()

	go client.readPump()
	go client.writePump()

}

func handleBroadcast() { //broadcasts messages for clients in their send channel
	for {
		UpdateMessage := <-broadcast //grab next message from broadcast

		message, err := json.Marshal(UpdateMessage)
		if err != nil {
			fmt.Printf("Error marshalling broadcast msg: %s", err)
			continue //skip this message
		}

		mutex.Lock()
		for id, client := range clients {
			if id != UpdateMessage.ClientID {
				select { //waits on multiple channel operations
				case client.send <- message: //puts the message in clients send channels
				default:
					//client's send buffer is full, close the connection
					close(client.send)
					delete(clients, id)
				}
			}
		}
		mutex.Unlock()
	}
}
func main() {

	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	})
	go handleBroadcast()
	fmt.Println("Websocket server started on port:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting websocket server")
	}

}
