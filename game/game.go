package game

import (
	"fmt"
	"github.com/google/uuid"
	"golang.org/x/net/websocket"
	"math/rand"
	"strconv"
)

type Player struct {
	Uid  string
	Name string
	ws   *websocket.Conn
}

func newPlayer(name string) (*Player, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("cannot create new DestPlayer with empty name\n")
	}

	uid := uuid.NewString()

	fmt.Printf("New DestPlayer created with uid %v and name %v\n", uid, name)
	return &Player{Uid: uid, Name: name, ws: nil}, nil
}

type Game struct {
	Code    string
	Players map[string]*Player
	Vote    *Vote
}

func newGame(code string) *Game {
	players := make(map[string]*Player)

	fmt.Printf("New game created with code %v\n", code)
	return &Game{Code: code, Players: players}
}

type Vote struct {
	DestPlayer   *Player
	OriginPlayer *Player
	Votes        map[string]string //playerid - vote (ja, nein, empty)
}

func (g *Game) NewVote(origin *Player, dest *Player) (*Vote, error) {
	if g.Vote != nil {
		return g.Vote, fmt.Errorf("vote already exists")
	}

	votes := make(map[string]string)
	for key, _ := range g.Players {
		votes[key] = ""
	}

	g.Vote = &Vote{OriginPlayer: origin, DestPlayer: dest, Votes: votes}
	return g.Vote, nil
}

func (g *Game) MakeVote(dest *Player, fromId, vote string) error {
	if g.Vote == nil {
		return fmt.Errorf("no votes ongoing in this game")
	}

	if g.Vote.DestPlayer.Uid != dest.Uid {
		return fmt.Errorf("you are trying to vote for %v, while ongoing vote is against %v", dest.Name, g.Vote.DestPlayer.Name)
	}

	g.Vote.Votes[fromId] = vote
	return nil
}

type VoteResult struct {
	Yes        []string // name, not uid
	No         []string // name, not uid
	Empty      []string //name, not uid
	Finished   bool
	Success    bool
	PlayerName string
}

func (g *Game) FinishVote(dest *Player) (*VoteResult, error) {
	if g.Vote == nil {
		return nil, fmt.Errorf("no votes ongoing in this game")
	}

	if g.Vote.DestPlayer.Uid != dest.Uid {
		return nil, fmt.Errorf("you are trying to vote for %v, while ongoing vote is against %v", dest.Name, g.Vote.DestPlayer.Name)
	}

	var yes []string
	var no []string
	var empty []string

	for pui, voteRes := range g.Vote.Votes {
		switch voteRes {
		case "yes":
			yes = append(yes, g.Players[pui].Name)
		case "no":
			no = append(no, g.Players[pui].Name)
		case "":
			empty = append(empty, g.Players[pui].Name)
		}

	}

	// tie is a fail
	fmt.Printf("length %v", len(empty))
	success := len(yes) > len(no)
	finished := len(empty) == 0

	if finished {
		// finish vote
		// todo notify players
		g.Vote = nil
	}

	return &VoteResult{Empty: empty, Yes: yes, No: no, Finished: finished, Success: success, PlayerName: dest.Name}, nil
}

func (g *Game) CancelVote() {

	// finish vote
	// todo notify players
	g.Vote = nil

}

func (g *Game) addPlayer(name string) (*Player, error) {
	p, err := newPlayer(name)
	if err != nil {
		return nil, err
	}

	g.Players[p.Uid] = p
	fmt.Printf("%v added to game %v\n", p.Name, g.Code)
	return p, nil
}

type GamePool struct {
	Games map[string]*Game
}

func NewGamePool() *GamePool {
	return &GamePool{Games: make(map[string]*Game)}
}

func (gp *GamePool) FindGame(code string) *Game {
	return gp.Games[code]
}

func (gp *GamePool) FindPlayer(code, playerId string) (*Player, error) {
	g := gp.Games[code]
	if g == nil {
		return nil, fmt.Errorf("game with code %v does not exist", code)
	}

	p := g.Players[playerId]
	if p == nil {
		return nil, fmt.Errorf("player with id %v does not exist in game %v", playerId, code)
	}

	return p, nil
}

func (gp *GamePool) VoteForPlayer(code, playerId string) (*Player, error) {
	g := gp.Games[code]
	if g == nil {
		return nil, fmt.Errorf("game with code %v does not exist", code)
	}

	p := g.Players[playerId]
	if p == nil {
		return nil, fmt.Errorf("player with id %v does not exist in game %v", playerId, code)
	}

	return p, nil
}

func (gp *GamePool) StartGame(playerName string) (string, *Player, error) {
	iCode := 0
	const minCode = 11111

	for {
		iCode = minCode + rand.Intn(99999-minCode)
		code := strconv.Itoa(iCode)

		_, contains := gp.Games[code]
		if !contains {
			g := newGame(code)
			p, err := g.addPlayer(playerName)
			if err != nil {
				return "", nil, err
			}
			gp.Games[code] = g
			fmt.Printf("Starting game %v with DestPlayer %v\n", code, playerName)
			return code, p, nil
		}
		fmt.Printf("trying to create game, code already existed: %v\n", code)
	}
}

func (gp *GamePool) JoinGame(code string, playerName string) (*Player, error) {
	fmt.Printf("%v trying to join game %v\n", playerName, code)

	g, contains := gp.Games[code]
	if !contains {
		fmt.Printf("%v failed to join game %v, code didn't exist\n", playerName, code)
		return nil, fmt.Errorf("could not find a game with code %v", code)
	}

	p, err := g.addPlayer(playerName)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%v successfully joined game %v\n", playerName, code)
	return p, nil
}

func (gp *GamePool) RemoveFromGame(code string, playerId string) error {
	g := gp.Games[code]
	if g == nil {
		return fmt.Errorf("game with code %v does not exist", code)
	}

	p := g.Players[playerId]
	if p == nil {
		return fmt.Errorf("DestPlayer with id %v does not exist in game %v", playerId, code)
	}

	fmt.Printf("remove %v from game %v\n", p.Name, code)
	delete(g.Players, playerId)
	return nil
}
