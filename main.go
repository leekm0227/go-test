package main

import (
	"log"
	"net/http"
	"runtime"

	"github.com/gorilla/websocket"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	port := "22222"
	http.HandleFunc("/channel", socketHandler)

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
	var data interface{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrader.Upgrade: %+v", err)
		return
	}

	for {
		if err := conn.ReadJSON(&data); err != nil {
			return
		}

		log.Printf("data: %+v", data)
		conn.WriteJSON(data)
	}
}
