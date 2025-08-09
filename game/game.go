package game

import (
	"fmt"
	"github.com/Neifen/secret-h/entities"
	"github.com/Neifen/secret-h/view"
	"github.com/gorilla/websocket"
	"math/rand"
	"strconv"
	"sync"
)

func (gp *GamePool) NewVote(gid string, origin *entities.Player, dest *entities.Player) (*entities.Vote, error) {
	g, err := gp.FindGame(gid)
	if err != nil {
		return nil, err
	}

	if g.Vote != nil {
		return g.Vote, fmt.Errorf("vote already exists")
	}

	votes := &sync.Map{}
	for key := range g.Players {
		votes.Store(key, "")
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

	g.Vote.Votes.Store(fromId, vote)

	// notify
	fmt.Println("voting, waiting?", g.Vote.Waiting)
	if g.Vote.Waiting {
		p, err := gp.FindPlayer(gid, fromId)
		if err != nil {
			return err
		}

		if vote == "" {
			view.WSRenderAddPlayerWait(g.Vote.OriginPlayer.Ws, p)
			view.WSRenderRemoveTryAgainWait(g.Vote.OriginPlayer.Ws, gid, g.Vote.OriginPlayer.Uid, g.Vote.DestPlayer.Uid)
		} else {
			view.WSRenderRemovePlayerWait(g.Vote.OriginPlayer.Ws, p)
			count := 0
			g.Vote.Votes.Range(func(k, v interface{}) bool {
				if vote == "" {
					count++
					return false
				}
				return true
			})
			if count == 0 {
				view.WSRenderAddTryAgainWait(g.Vote.OriginPlayer.Ws, gid, g.Vote.OriginPlayer.Uid, g.Vote.DestPlayer.Uid)
			}

		}
	}
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
	var empty []*entities.Player

	g.Vote.Votes.Range(func(k, voteRes interface{}) bool {
		pui := k.(string)
		switch voteRes {
		case "yes":
			yes = append(yes, g.Players[pui].Name)
		case "no":
			no = append(no, g.Players[pui].Name)
		case "":
			empty = append(empty, g.Players[pui])
		}
		return true
	})

	// tie is a fail
	success := len(yes) > len(no)
	finished := len(empty) == 0
	result := &entities.VoteResult{Empty: empty, Yes: yes, No: no, Finished: finished, Success: success, PlayerName: dest.Name}

	if finished {
		// finish vote
		g.Vote = nil

		// inform websockets
		// todo countdown?
		// president gets this double, oh well
		for _, wsPlayer := range g.Players {
			if wsPlayer.Ws != nil {
				view.WsRenderAfterVote(wsPlayer.Ws, result)
			}
		}
	} else {
		g.Vote.Waiting = true
	}

	return result, nil
}

func (gp *GamePool) CancelWait(gid string) {
	g, _ := gp.FindGame(gid)
	if g != nil && g.Vote != nil {
		g.Vote.Waiting = false
	}
}

func (gp *GamePool) CancelVote(gid string) {

	g, _ := gp.FindGame(gid)
	if g != nil {
		g.Vote = nil
		// inform websockets
		for _, wsPlayer := range g.Players {
			if wsPlayer.Ws != nil {
				view.WsRenderCancelVote(wsPlayer.Ws)
			}
		}
	}
}

type GamePool struct {
	Games *sync.Map
}

func NewGamePool() *GamePool {
	return &GamePool{&sync.Map{}}
}

func (gp *GamePool) FindGame(gid string) (*entities.Game, error) {
	g, _ := gp.Games.Load(gid)
	if g == nil {
		errMsg := fmt.Errorf("game with id %v does not exist", gid)
		return nil, errMsg
	}

	return g.(*entities.Game), nil
}

func (gp *GamePool) SetPlayerWS(conn *websocket.Conn, gid, pid string) error {
	g, err := gp.FindGame(gid)
	if err != nil {
		return err
	}
	p := g.Players[pid]
	if p == nil {
		return fmt.Errorf("player with id %v does not exist in game %v", pid, gid)
	}

	p.Ws = conn
	return nil
}

func (gp *GamePool) FindPlayer(code, playerId string) (*entities.Player, error) {
	g, err := gp.FindGame(code)
	if err != nil {
		return nil, err
	}

	p := g.Players[playerId]
	if p == nil {
		return nil, fmt.Errorf("player with id %v does not exist in game %v", playerId, code)
	}

	return p, nil
}

func (gp *GamePool) VoteForPlayer(code, playerId string) (*entities.Player, error) {
	g, err := gp.FindGame(code)
	if err != nil {
		return nil, err
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

		_, contains := gp.Games.Load(code)
		if !contains {
			g := entities.NewGame(code)
			p, err := g.AddPlayer(playerName)
			if err != nil {
				return "", nil, err
			}
			gp.Games.Store(code, g)
			fmt.Printf("Starting game %v with DestPlayer %v\n", code, playerName)
			return code, p, nil
		}
		fmt.Printf("trying to create game, code already existed: %v\n", code)
	}
}

func (gp *GamePool) JoinGame(gid string, playerName string) (*entities.Player, error) {
	fmt.Printf("%v trying to join game %v\n", playerName, gid)

	g, _ := gp.FindGame(gid)
	if g == nil {
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
	g, err := gp.FindGame(code)
	if err != nil {
		return err
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
	if len(g.Players) == 0 {
		gp.Games.Delete(code)
	}
	return nil
}
