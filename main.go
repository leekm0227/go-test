package main

import (
	"log"
	"math"
	"math/rand"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var channel Channel = Channel{
	clients: make(map[string]*Client),
	players: make(map[string]*Player),
}

const MAX_SIZE = 100

type PayloadType int

const (
	BROADCAST PayloadType = 0
	INIT      PayloadType = 1
	MOVE      PayloadType = 2
	ATTACK    PayloadType = 3
	DEAD      PayloadType = 4
	SPAWN     PayloadType = 5
)

func main() {
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
	defer conn.Close()

	// regist client
	client := Client{
		Sid:  strings.Split(uuid.NewString(), "-")[0],
		Conn: conn,
	}
	channel.clients[client.Sid] = &client

	for {
		var req Request
		if err := conn.ReadJSON(&req); err != nil {
			log.Printf("conn.ReadJSON: %+v", err)
			return
		}

		req.Sid = client.Sid
		eventHandler(req)
	}
}

func eventHandler(req Request) {
	var client = channel.clients[req.Sid]
	if client == nil {
		return
	}

	switch req.PayloadType {
	case INIT:
		Player := Player{
			Id:  client.Sid,
			Pos: [2]int{rand.Intn(MAX_SIZE), rand.Intn(MAX_SIZE)},
			Hp:  [2]int{10, 0},
		}
		channel.players[client.Sid] = &Player

		client.Conn.WriteJSON(Response{
			PayloadType: "INIT",
			Id:          client.Sid,
			Players:     channel.players,
			RegTime:     req.RegTime,
		})

		channel.Broadcast(Response{
			PayloadType: "SPAWN",
			SessionId:   client.Sid,
			Player:      Player,
			RegTime:     req.RegTime,
		})
	case MOVE:
		var player = channel.players[req.Sid]
		if player == nil {
			return
		}

		var x = player.Pos[0]
		var y = player.Pos[1]
		x += req.Dir[0]
		y += req.Dir[1]

		if x < 0 {
			x = 0
		} else if x > MAX_SIZE {
			x = MAX_SIZE
		}

		if y < 0 {
			y = 0
		} else if y > MAX_SIZE {
			y = MAX_SIZE
		}

		player.Pos[0] = x
		player.Pos[1] = y

		channel.Broadcast(Response{
			PayloadType: "MOVE",
			SessionId:   client.Sid,
			Player:      *player,
			RegTime:     req.RegTime,
		})
	case ATTACK:
		var target = channel.players[req.TargetId]
		if target == nil {
			return
		}

		var x = channel.players[req.Sid].Pos[0] - target.Pos[0]
		var y = channel.players[req.Sid].Pos[1] - target.Pos[1]
		var dis = math.Sqrt(float64((x * x) + (y * y)))
		if dis < 2 {
			target.Hp[0] = target.Hp[0] - 1
		}

		channel.Broadcast(Response{
			PayloadType: "ATTACK",
			SessionId:   req.TargetId,
			Player:      *target,
			RegTime:     req.RegTime,
		})

		if target.Hp[0] < 1 {
			delete(channel.players, req.TargetId)
			channel.Broadcast(Response{
				PayloadType: "DEAD",
				SessionId:   req.TargetId,
				RegTime:     req.RegTime,
			})
		}
	}
}

type Channel struct {
	clients map[string]*Client
	players map[string]*Player
}

type Client struct {
	Sid  string
	Conn *websocket.Conn
}

type Player struct {
	Id  string `json:"id"`
	Pos [2]int `json:"pos"`
	Hp  [2]int `json:"hp"`
}

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
	TargetId    string
	Dir         [2]int
}

func (client *Client) Send(response Response) {
	client.Conn.WriteJSON(response)
}

func (channel *Channel) Broadcast(response Response) {
	for _, client := range channel.clients {
		client.Conn.WriteJSON(response)
	}
}
