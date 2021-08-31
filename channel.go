package main

type Channel struct {
	clients   []*Client
	join      chan *Client
	leave     chan *Client
	broadcast chan Response
}

func newChannel() Channel {
	return Channel{
		clients:   make([]*Client, 0),
		join:      make(chan *Client, CHAN_BUFFER_SIZE),
		leave:     make(chan *Client, CHAN_BUFFER_SIZE),
		broadcast: make(chan Response, CHAN_BUFFER_SIZE),
	}
}

func (channel *Channel) run() {
	for {
		select {
		case client := <-channel.join:
			channel.clients = append(channel.clients, client)
		case client := <-channel.leave:
			for i, v := range channel.clients {
				if v == client {
					channel.clients = append(channel.clients[:i], channel.clients[i+1:]...)
					break
				}
			}
		case res := <-channel.broadcast:
			for _, v := range channel.clients {
				v.outbound <- res
			}
		}
	}
}
