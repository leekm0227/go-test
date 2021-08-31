package main

import (
	"github.com/gorilla/websocket"
)

type Client struct {
	Sid      string
	Conn     *websocket.Conn
	outbound chan Response
}

func newClient(sid string, conn *websocket.Conn) Client {
	return Client{
		Sid:      sid,
		Conn:     conn,
		outbound: make(chan Response, CHAN_BUFFER_SIZE),
	}
}

func (client *Client) receive() {
	for {
		res := <-client.outbound
		client.Conn.WriteJSON(res)
	}
}

func (client *Client) send() {
	defer func() {
		channel.leave <- client
		client.Conn.Close()
	}()

	for {
		var req Request
		if err := client.Conn.ReadJSON(&req); err != nil {
			return
		}

		channel.broadcast <- Response{
			Sid:     client.Sid,
			Body:    req.body,
			RegTime: req.RegTime,
		}
	}
}
