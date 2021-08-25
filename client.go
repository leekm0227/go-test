package main

import (
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Sid      string
	Conn     *websocket.Conn
	Player   *Player
	action   chan Action
	outbound chan Response
}

func newClient(sid string, conn *websocket.Conn) Client {
	return Client{
		Sid:      sid,
		Conn:     conn,
		action:   make(chan Action, CHAN_BUFFER_SIZE),
		outbound: make(chan Response, CHAN_BUFFER_SIZE),
	}
}

func (client *Client) receive() {
	for {
		select {
		case act := <-client.action:
			switch act.actionType {
			case ACT_ATTACK:
				client.Player.attack(act.pos, act.power, act.regTime)
			case ACT_MOVE:
				client.Player.move(act.dir, act.regTime)
			}
		case res := <-client.outbound:
			client.Conn.WriteJSON(res)
		}
	}
}

func (client *Client) send() {
	defer func() {
		// log.Printf("client leave")
		channel.leave <- *client
		channel.broadcast <- Response{
			PayloadType: "DEAD",
			SessionId:   client.Sid,
			RegTime:     uint64(time.Now().Unix() * 1000),
		}
		client.Conn.Close()
	}()

	for {
		var req Request
		if err := client.Conn.ReadJSON(&req); err != nil {
			// log.Printf("conn.ReadJSON: %+v", err)
			return
		}

		switch req.PayloadType {
		case INIT:
			client.Player = newPlayer(client.Sid)
			channel.PlayerManager.add(client.Sid, *client.Player)
			client.outbound <- Response{
				PayloadType: "INIT",
				Id:          client.Sid,
				Players:     channel.PlayerManager.list(),
				RegTime:     req.RegTime,
			}

			channel.broadcast <- Response{
				PayloadType: "SPAWN",
				SessionId:   client.Sid,
				Player:      *client.Player,
				RegTime:     req.RegTime,
			}
		case MOVE:
			client.action <- newActionMove(req.Dir, req.RegTime)
		case ATTACK:
			channel.attack <- newActionAttack(req.TargetId, client.Player.Pos, POWER, req.RegTime)
		}
	}
}
