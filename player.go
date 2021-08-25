package main

import (
	"math"
	"math/rand"
	"sync"
)

type Player struct {
	Id  string `json:"id"`
	Pos [2]int `json:"pos"`
	Hp  [2]int `json:"hp"`
}

type PlayerManager struct {
	mu      sync.Mutex
	players map[string]Player
}

func newPlayer(sid string) *Player {
	return &Player{
		Id:  sid,
		Pos: [2]int{rand.Intn(MAX_SIZE), rand.Intn(MAX_SIZE)},
		Hp:  [2]int{10, 0},
	}
}

func newPlayerManager() PlayerManager {
	return PlayerManager{
		players: make(map[string]Player),
		mu:      sync.Mutex{},
	}
}

func (player *Player) attack(pos [2]int, power int, regTime uint64) {
	var x = player.Pos[0] - pos[0]
	var y = player.Pos[1] - pos[1]
	var dis = math.Sqrt(float64((x * x) + (y * y)))

	if dis < 2 {
		player.Hp[0] = player.Hp[0] - power
	}

	channel.broadcast <- Response{
		PayloadType: "ATTACK",
		SessionId:   player.Id,
		Player:      *player,
		RegTime:     regTime,
	}

	if player.Hp[0] < 1 {
		channel.broadcast <- Response{
			PayloadType: "DEAD",
			SessionId:   player.Id,
			RegTime:     regTime,
		}
	}
}

func (player *Player) move(dir [2]int, regTime uint64) {
	var x = player.Pos[0]
	var y = player.Pos[1]
	x += dir[0]
	y += dir[1]

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

	channel.broadcast <- Response{
		PayloadType: "MOVE",
		SessionId:   player.Id,
		Player:      *player,
		RegTime:     regTime,
	}
}

func (playerManager *PlayerManager) add(key string, player Player) {
	playerManager.mu.Lock()
	defer playerManager.mu.Unlock()
	playerManager.players[key] = player
}

func (playerManager *PlayerManager) remove(key string) {
	playerManager.mu.Lock()
	defer playerManager.mu.Unlock()
	delete(playerManager.players, key)
}

func (playerManager *PlayerManager) list() map[string]Player {
	playerManager.mu.Lock()
	defer playerManager.mu.Unlock()
	newMap := make(map[string]Player)

	for k, v := range playerManager.players {
		newMap[k] = v
	}

	return newMap
}
