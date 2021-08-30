package main

import (
	"math"
	"math/rand"
	"sync"
)

type Player struct {
	mu  *sync.Mutex
	Id  string `json:"id"`
	Pos [2]int `json:"pos"`
	Hp  [2]int `json:"hp"`
}

type PlayerManager struct {
	mu      sync.Mutex
	players map[string]*Player
}

func newPlayer(sid string) *Player {
	return &Player{
		mu:  &sync.Mutex{},
		Id:  sid,
		Pos: [2]int{rand.Intn(MAX_SIZE), rand.Intn(MAX_SIZE)},
		Hp:  [2]int{10, 0},
	}
}

func newPlayerManager() PlayerManager {
	return PlayerManager{
		players: make(map[string]*Player),
		mu:      sync.Mutex{},
	}
}

func (player *Player) attack(enemyPos [2]int, regTime uint64) {
	x, y := player.getPos()
	x = x - enemyPos[0]
	y = y - enemyPos[1]
	var dis = math.Sqrt(float64((x * x) + (y * y)))

	if dis < 2 {
		player.Hp[0] = player.Hp[0] - POWER
	}

	result := player
	channel.broadcast <- Response{
		PayloadType: "ATTACK",
		SessionId:   result.Id,
		Player:      *result,
		RegTime:     regTime,
	}
}

func (player *Player) getPos() (int, int) {
	pos := player.Pos
	return pos[0], pos[1]
}

func (player *Player) setPos(x int, y int) {
	player.mu.Lock()
	defer player.mu.Unlock()
	player.Pos[0] = x
	player.Pos[1] = y
}

func (player *Player) move(dir [2]int, regTime uint64) {
	x, y := player.getPos()
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

	player.setPos(x, y)
	channel.broadcast <- Response{
		PayloadType: "MOVE",
		SessionId:   player.Id,
		Player:      *player,
		RegTime:     regTime,
	}
}

func (playerManager *PlayerManager) add(key string, player *Player) {
	playerManager.mu.Lock()
	defer playerManager.mu.Unlock()
	playerManager.players[key] = player
}

func (playerManager *PlayerManager) remove(key string) {
	playerManager.mu.Lock()
	defer playerManager.mu.Unlock()
	delete(playerManager.players, key)
}
