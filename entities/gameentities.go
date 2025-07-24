package entities

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Game struct {
	Code    string
	Players map[string]*Player
	Vote    *Vote
}

func NewGame(code string) *Game {
	players := make(map[string]*Player)

	fmt.Printf("New game created with code %v\n", code)
	return &Game{Code: code, Players: players}
}

type VoteResult struct {
	Yes        []string // name, not uid
	No         []string // name, not uid
	Empty      []string //name, not uid
	Finished   bool
	Success    bool
	PlayerName string
}

type Player struct {
	Uid  string
	Name string
	Ws   *websocket.Conn
}

func NewPlayer(name string) (*Player, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("cannot create new DestPlayer with empty name\n")
	}

	uid := uuid.NewString()

	fmt.Printf("New DestPlayer created with uid %v and name %v\n", uid, name)
	return &Player{Uid: uid, Name: name, Ws: nil}, nil
}

type Vote struct {
	DestPlayer   *Player
	OriginPlayer *Player
	Votes        map[string]string //playerid - vote (ja, nein, empty)
}

func (g *Game) AddPlayer(name string) (*Player, error) {
	p, err := NewPlayer(name)
	if err != nil {
		return nil, err
	}

	g.Players[p.Uid] = p
	fmt.Printf("%v added to game %v\n", p.Name, g.Code)
	return p, nil
}
