package main

type Channel struct {
	clientMap map[string]*Client
	join      chan *Client
	leave     chan *Client
	broadcast chan *Response
}

func newChannel() Channel {
	return Channel{
		clientMap: make(map[string]*Client),
		join:      make(chan *Client, CHAN_BUFFER_SIZE),
		leave:     make(chan *Client, CHAN_BUFFER_SIZE),
		broadcast: make(chan *Response, CHAN_BUFFER_SIZE),
	}
}

func (channel *Channel) run() {
	for {
		select {
		case client := <-channel.join:
			channel.clientMap[client.Sid] = client
		case client := <-channel.leave:
			delete(channel.clientMap, client.Sid)
		case res := <-channel.broadcast:
			for _, client := range channel.clientMap {
				client.outbound <- *res
			}
		}
	}
}
