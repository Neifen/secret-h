package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"secret-h/view"
)

// e.POST("/closePopup", s.closePopupHandler)
func (s *Session) closePopupHandler(c echo.Context) error {
	return view.ClosePopup(c)
}

// e.GET("/lobby/:id/:player", s.lobbyHandler)
func (s *Session) lobbyHandler(c echo.Context) error {
	c.Response().Header().Set("HX-Refresh", "true")
	id := c.Param("id")
	pid := c.Param("player")

	g := s.gamePool.FindGame(id)
	if g == nil {
		return redirectHome(c)
	}

	p, err := s.gamePool.FindPlayer(id, pid)
	if err != nil {
		//todo maybe an observer mode?
		return redirectHome(c)
	}

	return view.RenderViewLobby(c, g, p)
}

// e.POST("/vote/:id/:originPid/:destPlayer", s.initVoteHandler)
func (s *Session) initVoteHandler(c echo.Context) error {
	gid := c.Param("id")
	destPid := c.Param("destPid")
	originPid := c.Param("originPid")

	destPlayer, err := s.gamePool.FindPlayer(gid, destPid)
	if err != nil {
		return view.RenderError(c, err)
	}
	originPlayer, err := s.gamePool.FindPlayer(gid, originPid)
	if err != nil {
		return view.RenderError(c, err)
	}

	g := s.gamePool.FindGame(gid)
	if g == nil {
		errMsg := fmt.Sprintf("game with id %v does not exist", gid)
		return view.RenderMessage(c, errMsg)
	}

	v, err := g.NewVote(originPlayer, destPlayer)
	if err != nil {
		if v != nil {
			// vote already exists
			return view.RenderVote(c, v.OriginPlayer == originPlayer, gid, v.Votes[originPid], originPid, v.DestPlayer)
		}

		return view.RenderError(c, err)
	}

	// brand new vote
	return view.RenderVote(c, true, gid, "", originPid, v.DestPlayer)
}

// e.POST("/cancel-vote/:id/", s.cancelVoteHandler)
func (s *Session) cancelVoteHandler(c echo.Context) error {
	gid := c.Param("id")

	g := s.gamePool.FindGame(gid)
	if g != nil {
		g.CancelVote()
	}

	return view.ClosePopup(c)
}

// e.POST("/make-vote/:id/:originPid/:destPid", s.makeVoteHandler)
func (s *Session) makeVoteHandler(c echo.Context) error {
	gid := c.Param("id")
	originPid := c.Param("originPid")
	destPid := c.Param("destPid")

	toggle := c.QueryParam("toggle")

	g := s.gamePool.FindGame(gid)
	if g == nil {
		errMsg := fmt.Sprintf("game with id %v does not exist", gid)
		return view.RenderMessage(c, errMsg)
	}

	destPlayer, err := s.gamePool.FindPlayer(gid, destPid)
	if err != nil {
		return view.RenderError(c, err)
	}
	err = g.MakeVote(destPlayer, originPid, toggle)

	return view.RenderVoteButton(c, gid, toggle, originPid, destPlayer)
}

// e.POST("/finish-vote/:id/:player", s.finishVoteHandler)
func (s *Session) finishVoteHandler(c echo.Context) error {
	gid := c.Param("id")
	originPid := c.Param("originPid")
	destPid := c.Param("destPid")

	g := s.gamePool.FindGame(gid)
	if g == nil {
		errMsg := fmt.Sprintf("game with id %v does not exist", gid)
		return view.RenderMessage(c, errMsg)
	}

	player, err := s.gamePool.FindPlayer(gid, destPid)
	if err != nil {
		return view.RenderError(c, err)
	}

	result, err := g.FinishVote(player)
	if err != nil {
		return view.RenderError(c, err)
	}

	if !result.Finished {
		// todo this could be more fluid with ws
		fmt.Printf("results: %+v", result)
		return view.RenderVoteWaitPopup(c, result.Empty, gid, originPid, destPid)
	}

	return view.RenderAfterVotePopup(c, result)
}

// e.POST("/kill/:id/:player", s.initKillHandler)
func (s *Session) initKillHandler(c echo.Context) error {
	gid := c.Param("id")
	pid := c.Param("player")
	p, err := s.gamePool.FindPlayer(gid, pid)
	if err != nil {
		return view.RenderError(c, err)
	}

	return view.RenderKillPopup(c, gid, p)
}

// e.POST("/kill-confirmed/:id/:player", s.killConfirmedHandler)
func (s *Session) killConfirmedHandler(c echo.Context) error {
	gid := c.Param("id")
	pid := c.Param("player")
	p, err := s.gamePool.FindPlayer(gid, pid)
	if err != nil {
		return view.RenderError(c, err)
	}

	pName := p.Name
	err = s.gamePool.RemoveFromGame(gid, pid)
	if err != nil {
		return view.RenderError(c, err)
	}
	return view.RenderKillConfirmPopup(c, pName)
}
