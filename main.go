package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type MessageObject struct {
	From    string
	Content string
}

var upgrader = websocket.Upgrader{}
var todoList []string
var connections []*websocket.Conn

func main() {

	r := mux.NewRouter()

	r.HandleFunc("/chat/{username}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		username := params["username"]

		// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}
		connections = append(connections, conn)
		defer conn.Close()

		messageObject := MessageObject{From: username, Content: "connected"}

		broadcast(messageObject, 1)

		// Continuosly read and write message
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("read failed:", err)
				break
			}

			var messageObject MessageObject
			messageObject.From = username

			if err := json.Unmarshal(message, &messageObject); err != nil {
				log.Println("error unmarshaling message", err)
				break
			}

			fmt.Printf("%v : %v", message, messageObject)

			broadcast(messageObject, mt)
		}
	})

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.ListenAndServe(":8090", r)
}

func broadcast(m MessageObject, mt int) {
	fmt.Println("Broadcasting", m, mt)
	messageString, err := json.Marshal(m)
	if err != nil {
		log.Println("error marshaling message", err)
	}

	message := []byte(messageString)
	for _, conn := range connections {
		err = conn.WriteMessage(mt, message)
		if err != nil {
			log.Println("write failed:", err)
		}
	}
}
