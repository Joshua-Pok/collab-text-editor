package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
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

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pgURL := getEnv("DATABASE_URL", "postgres://collab:collabpass@localhost:5432/collabdb")
	mqttURL := getEnv("MQTT_URL", "tcp://localhost:1883")
	topic := getEnv("MQTT_TOPIC", "doc/+/updates") //wildcard subscribe

	pool, err := pgxpool.New(ctx, pgURL) //pool of postgres connections, important for concurrent backends because multiple writes may happen simultaenously
	if err != nil {
		log.Fatalf("Error creating pg pool: %v", err)
	}

	defer pool.Close()

	opts := mqtt.NewClientOptions().
		AddBroker(mqttURL).
		SetClientID(fmt.Sprintf("persist-%d", time.Now().UnixNano())).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(2 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt connect: %v", token.Error())
	}

	fmt.Print("MQTT connection established")

	subscribeHandler := func(_ mqtt.Client, msg mqtt.Message) {
		docID, ok := parseDocID(string(msg.Topic()))
		if !ok {
			log.Printf("Skipping topic %s: cannot parse doc id", msg.Topic())
			return
		}

		var m UpdateMessage
		if err := json.Unmarshal(msg.Payload(), &m); err != nil {
			log.Printf("bad topic")
		}

		updateBytes, err := base64.StdEncoding.DecodeString(m.UpdateB64) //convert to bytes

		if err != nil {
			log.Printf("bad topic")
		}

		_, err = pool.Exec(context.Background(), `INSERT into doc_updates (doc_id, client_id, update_bytes) VALUES ($1, $2, $3)`, docID, m.ClientID, updateBytes)

		if err != nil {
			log.Printf("db insert failed doc=%s, err=%v", docID, err)
		}

	}

	if token := client.Subscribe(topic, 1, subscribeHandler); token.Wait() && token.Error() != nil {
		log.Fatalf("MQTT subscribe: %v", token.Error())
	}
	log.Printf("Subscribe to %s (qos=1)\n", topic)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	<-sigc
	log.Println("Shutting Down...")

}
