package game

import (
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"secret-h/entities"
	"secret-h/view"
	"strconv"
)

func (gp *GamePool) NewVote(gid string, origin *entities.Player, dest *entities.Player) (*entities.Vote, error) {
	g, err := gp.FindGame(gid)
	if err != nil {
		return nil, err
	}

	if g.Vote != nil {
		return g.Vote, fmt.Errorf("vote already exists")
	}

	votes := make(map[string]string)
	for key := range g.Players {
		votes[key] = ""
	}

	g.Vote = &entities.Vote{OriginPlayer: origin, DestPlayer: dest, Votes: votes}

	// inform websockets
	for _, p := range g.Players {
		if p.Ws != nil && p.Uid != origin.Uid {
			view.WsRenderVote(p.Ws, gid, p.Uid, dest)
		}
	}

	return g.Vote, nil
}

func (gp *GamePool) MakeVote(gid string, dest *entities.Player, fromId, vote string) error {
	g, err := gp.FindGame(gid)
	if err != nil {
		return err
	}

	if g.Vote == nil {
		return fmt.Errorf("no votes ongoing in this game")
	}

	if g.Vote.DestPlayer.Uid != dest.Uid {
		return fmt.Errorf("you are trying to vote for %v, while ongoing vote is against %v", dest.Name, g.Vote.DestPlayer.Name)
	}

	g.Vote.Votes[fromId] = vote
	return nil
}

func (gp *GamePool) FinishVote(gid string, dest *entities.Player) (*entities.VoteResult, error) {
	g, err := gp.FindGame(gid)
	if err != nil {
		return nil, err
	}

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

	return &entities.VoteResult{Empty: empty, Yes: yes, No: no, Finished: finished, Success: success, PlayerName: dest.Name}, nil
}

func (gp *GamePool) CancelVote(gid string) {

	g, _ := gp.FindGame(gid)
	if g != nil {
		// todo notify players
		g.Vote = nil
	}
}

type GamePool struct {
	Games map[string]*entities.Game
}

func NewGamePool() *GamePool {
	return &GamePool{Games: make(map[string]*entities.Game)}
}

func (gp *GamePool) FindGame(gid string) (*entities.Game, error) {
	g := gp.Games[gid]
	if g == nil {
		errMsg := fmt.Errorf("game with id %v does not exist", gid)
		return nil, errMsg
	}

	return g, nil
}

func (gp *GamePool) SetPlayerWS(conn *websocket.Conn, gid, pid string) error {
	g := gp.Games[gid]
	if g == nil {
		return fmt.Errorf("game with code %v does not exist", gid)
	}

	p := g.Players[pid]
	if p == nil {
		return fmt.Errorf("player with id %v does not exist in game %v", pid, gid)
	}

	p.Ws = conn
	return nil
}

func (gp *GamePool) FindPlayer(code, playerId string) (*entities.Player, error) {
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

func (gp *GamePool) VoteForPlayer(code, playerId string) (*entities.Player, error) {
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

func (gp *GamePool) StartGame(playerName string) (string, *entities.Player, error) {
	iCode := 0
	const minCode = 11111

	for {
		iCode = minCode + rand.Intn(99999-minCode)
		code := strconv.Itoa(iCode)

		_, contains := gp.Games[code]
		if !contains {
			g := entities.NewGame(code)
			p, err := g.AddPlayer(playerName)
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

func (gp *GamePool) JoinGame(gid string, playerName string) (*entities.Player, error) {
	fmt.Printf("%v trying to join game %v\n", playerName, gid)

	g, contains := gp.Games[gid]
	if !contains {
		fmt.Printf("%v failed to join game %v, code didn't exist\n", playerName, gid)
		return nil, fmt.Errorf("could not find a game with code %v", gid)
	}

	p, err := g.AddPlayer(playerName)
	if err != nil {
		return nil, err
	}

	// inform websockets
	for _, wsPlayer := range g.Players {
		if wsPlayer.Ws != nil && wsPlayer.Uid != p.Uid {
			view.WSRenderNewPlayer(wsPlayer.Ws, gid, wsPlayer.Uid, p)
		}
	}

	fmt.Printf("%v successfully joined game %v\n", playerName, gid)
	return p, nil
}

func (gp *GamePool) RemoveFromGame(code string, playerId string, kill bool) error {
	g := gp.Games[code]
	if g == nil {
		return fmt.Errorf("game with code %v does not exist", code)
	}

	p := g.Players[playerId]
	if p == nil {
		return fmt.Errorf("DestPlayer with id %v does not exist in game %v", playerId, code)
	}

	fmt.Printf("remove %v from game %v\n", p.Name, code)
	if p.Ws != nil && kill {
		// notify the killed player
		view.WSRenderRemovedPopup(p.Ws)
	}

	// notify all other players
	for _, wsPlayer := range g.Players {
		if wsPlayer.Ws != nil && wsPlayer.Uid != p.Uid {
			view.WSRenderRemovePlayer(wsPlayer.Ws, p.Uid)
		}
	}

	delete(g.Players, playerId)
	return nil
}
