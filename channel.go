package main

type Channel struct {
	clientMap     map[string]Client
	PlayerManager PlayerManager
	join          chan Client
	leave         chan Client
	broadcast     chan Response
	attack        chan Action
}

func newChannel() Channel {
	return Channel{
		clientMap:     make(map[string]Client),
		PlayerManager: newPlayerManager(),
		join:          make(chan Client, CHAN_BUFFER_SIZE),
		leave:         make(chan Client, CHAN_BUFFER_SIZE),
		broadcast:     make(chan Response, CHAN_BUFFER_SIZE),
		attack:        make(chan Action, CHAN_BUFFER_SIZE),
	}
}

func (channel *Channel) run() {
	for {
		select {
		case act := <-channel.attack:
			if client, ok := channel.clientMap[act.targetId]; ok {
				client.action <- act
			}
		case client := <-channel.join:
			channel.clientMap[client.Sid] = client
		case client := <-channel.leave:
			delete(channel.clientMap, client.Sid)
			channel.PlayerManager.remove(client.Sid)
		case res := <-channel.broadcast:
			for _, client := range channel.clientMap {
				client.outbound <- res
			}
		}
	}
}
