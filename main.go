package main

import (
	"log"
	"net/http"
	"runtime"

	"github.com/gorilla/websocket"
)

const (
	POWER            = 1
	MAX_SIZE         = 100
	CHAN_BUFFER_SIZE = 50

	BROADCAST PayloadType = 0
	INIT      PayloadType = 1
	MOVE      PayloadType = 2
	ATTACK    PayloadType = 3
	DEAD      PayloadType = 4
	SPAWN     PayloadType = 5

	ACT_MOVE   ActionType = 0
	ACT_ATTACK ActionType = 1
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
	var req Request
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrader.Upgrade: %+v", err)
		return
	}

	go func() {
		for {
			if err := conn.ReadJSON(&req); err != nil {
				return
			}

			conn.WriteJSON(Response{
				Id:      "test",
				RegTime: req.RegTime,
			})
		}
	}()
}

type PayloadType int
type ActionType int

type Response struct {
	PayloadType string `json:"payloadType"`
	SessionId   string `json:"sessionId"`
	RegTime     uint64 `json:"regTime"`

	// init
	Id string `json:"id"`
}

type Request struct {
	PayloadType PayloadType
	Sid         string
	RegTime     uint64
	TargetId    string `json:"targetId,omitempty"`
	Dir         [2]int
}
