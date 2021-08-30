package main

import (
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	CHAN_BUFFER_SIZE = 50
)

var channel Channel

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	channel = newChannel()
	go channel.run()

	http.HandleFunc("/channel", socketHandler)

	port := "22222"
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}

}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func socketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrader.Upgrade: %+v", err)
		return
	}

	sid := strings.Split(uuid.NewString(), "-")[0]
	client := newClient(sid, conn)
	channel.join <- &client
	go client.send()
	go client.receive()
}

type Response struct {
	Sid     string `json:"sid"`
	RegTime uint64 `json:"regTime"`
	Body    string `json:"body"`
}

type Request struct {
	Sid     string
	RegTime uint64
	body    string
}
