package main

import (
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Channel struct {
	clients   []Client
	playerMap map[string]Player
	addPlayer chan Player
	send      chan Response
	broadcast chan Response
	action    chan Action
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

func newChannel() Channel {
	return Channel{
		clients:   make([]Client, 0),
		playerMap: make(map[string]Player),
		addPlayer: make(chan Player, CHAN_BUFFER_SIZE),
		send:      make(chan Response, CHAN_BUFFER_SIZE),
		broadcast: make(chan Response, CHAN_BUFFER_SIZE),
		action:    make(chan Action, CHAN_BUFFER_SIZE),
	}
}

func (channel *Channel) run() {
	go func() {
		for {
			select {
			case action := <-channel.action:
				switch action.actionType {
				case ACT_MOVE:
					if player, ok := channel.playerMap[action.sid]; ok {
						player := movePlayer(player, action.dir)
						channel.playerMap[action.sid] = player
						channel.broadcast <- Response{
							PayloadType: "MOVE",
							SessionId:   player.Id,
							Player:      player,
							RegTime:     action.regTime,
						}
					}
				case ACT_ATTACK:
					if target, ok := channel.playerMap[action.targetId]; ok {
						if enemy, ok := channel.playerMap[action.enemyId]; ok {
							target := attackPlayer(target, enemy)
							channel.broadcast <- Response{
								PayloadType: "ATTACK",
								SessionId:   target.Id,
								Player:      target,
								RegTime:     action.regTime,
							}

							if target.Hp[0] < 1 {
								player := newPlayer(target.Id)
								channel.playerMap[target.Id] = player
								channel.broadcast <- Response{
									PayloadType: "SPAWN",
									SessionId:   target.Id,
									Player:      target,
									RegTime:     uint64(time.Now().Unix() * 1000),
								}
							}
						}
					}
				}
			case player := <-channel.addPlayer:
				channel.playerMap[player.Id] = player
			case res := <-channel.broadcast:
				for i := range channel.clients {
					channel.clients[i].Conn.WriteJSON(res)
				}
			case res := <-channel.send:
				for i := range channel.clients {
					if channel.clients[i].Sid == res.SessionId {
						channel.clients[i].Conn.WriteJSON(res)
						break
					}
				}
			}
		}
	}()
}

func newPlayer(sid string) Player {
	return Player{
		Id:  sid,
		Pos: [2]int{rand.Intn(MAX_SIZE), rand.Intn(MAX_SIZE)},
		Hp:  [2]int{10, 0},
	}
}

func (channel *Channel) handle(conn *websocket.Conn) {
	sid := strings.Split(uuid.NewString(), "-")[0]
	client := Client{
		Sid:  sid,
		Conn: conn,
	}

	channel.clients = append(channel.clients, client)
	go func() {
		var req Request
		defer func() {
			channel.leave(sid)
			channel.broadcast <- Response{
				PayloadType: "DEAD",
				SessionId:   sid,
				RegTime:     uint64(time.Now().Unix() * 1000),
			}
			client.Conn.Close()
		}()

		for {
			if err := client.Conn.ReadJSON(&req); err != nil {
				return
			}

			switch req.PayloadType {
			case INIT:
				player := newPlayer(sid)
				channel.addPlayer <- player
				channel.send <- Response{
					PayloadType: "SPAWN",
					SessionId:   sid,
					Players:     channel.playerMap,
					RegTime:     req.RegTime,
				}

				channel.broadcast <- Response{
					PayloadType: "SPAWN",
					SessionId:   sid,
					Player:      player,
					RegTime:     req.RegTime,
				}
			case MOVE:
				channel.action <- newActionMove(sid, req.Dir, req.RegTime)
			case ATTACK:
				channel.action <- newActionAttack(req.TargetId, sid, req.RegTime)
			}
		}
	}()
}

func (channel *Channel) leave(sid string) {
	for i := range channel.clients {
		if channel.clients[i].Sid == sid {
			channel.clients = append(channel.clients[:i], channel.clients[i+1:]...)
			break
		}
	}
}

func movePlayer(player Player, dir [2]int) Player {
	x := player.Pos[0] + dir[0]
	y := player.Pos[1] + dir[1]

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
	return player
}

func attackPlayer(target Player, enemy Player) Player {
	x := target.Pos[0] - enemy.Pos[0]
	y := target.Pos[1] - enemy.Pos[1]
	var dis = math.Sqrt(float64((x * x) + (y * y)))

	if dis < 2 {
		target.Hp[0] = target.Hp[0] - POWER
	}

	return target
}
