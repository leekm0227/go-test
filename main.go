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
	channel.join <- client
	go client.send()
	go client.receive()
	// log.Printf("client start")
}

type PayloadType int
type ActionType int

type Response struct {
	PayloadType string `json:"payloadType"`
	SessionId   string `json:"sessionId"`
	RegTime     uint64 `json:"regTime"`
	Player      Player `json:"player"`

	// init
	Id      string             `json:"id"`
	Players map[string]*Player `json:"players"`
}

type Request struct {
	PayloadType PayloadType
	Sid         string
	RegTime     uint64
	TargetId    string `json:"targetId,omitempty"`
	Dir         [2]int

	// attack
	Player Player `json:"player,omitempty"`
}

type Action struct {
	actionType ActionType
	regTime    uint64

	// attack
	targetId string
	enemyPos [2]int

	// move
	dir [2]int
}

func newActionAttack(targetId string, pos [2]int, regTime uint64) Action {
	return Action{
		actionType: ACT_ATTACK,
		regTime:    regTime,
		targetId:   targetId,
		enemyPos:   pos,
	}
}

func newActionMove(dir [2]int, regTime uint64) Action {
	return Action{
		actionType: ACT_MOVE,
		regTime:    regTime,
		dir:        dir,
	}
}
