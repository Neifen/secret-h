package entities

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Game struct {
	Code      string
	Players   *sync.Map // string - *Player
	Vote      *Vote
	CreatedAt time.Time
}

func NewGame(code string) *Game {
	players := &sync.Map{}

	fmt.Printf("New game created with code %v\n", code)
	return &Game{Code: code, Players: players, CreatedAt: time.Now()}
}

type VoteResult struct {
	Yes        []string  // name, not uid
	No         []string  // name, not uid
	Empty      []*Player //player
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
	Votes        *sync.Map //playerid - vote (ja, nein, empty)
	Waiting      bool      // origin player is on "wait" screen
}

func (g *Game) AddPlayer(name string) (*Player, error) {
	p, err := NewPlayer(name)
	if err != nil {
		return nil, err
	}

	g.Players.Store(p.Uid, p)
	fmt.Printf("%v added to game %v\n", p.Name, g.Code)
	return p, nil
}
