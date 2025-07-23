package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"secret-h/view"
)

// e.GET("/", s.homeHandler)
func (s *Session) homeHandler(c echo.Context) error {
	c.Response().Header().Set("HX-Refresh", "true")
	return view.RenderView(c, view.ViewHome())
}

func redirectHome(c echo.Context) error {
	if c.Request().Header.Get("Hx-Request") != "true" {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	c.Response().Header().Set("HX-Redirect", "/") //HX-Redirect to url
	return c.NoContent(http.StatusOK)
}

// e.POST("/closePopup", s.closePopupHandler)
func (s *Session) closePopupHandler(c echo.Context) error {
	return view.ClosePopup(c)
}

// e.POST("/start", s.startHandler)
func (s *Session) startHandler(c echo.Context) error {
	playerName := c.FormValue("name")
	if playerName == "" {
		return view.RenderMessage(c, "Field \"Name\" is a required field")
	}
	
	code, p, err := s.gamePool.StartGame(playerName)
	if err != nil {
		return view.RenderView(c, view.ViewError(err.Error())) // todo add error to site
	}

	url := fmt.Sprintf("/lobby/%v/%s", code, p.Uid)
	c.Response().Header().Set("HX-Redirect", url) //HX-Redirect to url
	return c.NoContent(http.StatusOK)
}

// e.POST("/join", s.joinHandler)
func (s *Session) joinHandler(c echo.Context) error {
	playerName := c.FormValue("name")
	if playerName == "" {
		return view.RenderMessage(c, "Field \"Name\" is a required field")
	}
	code := c.FormValue("code")

	p, err := s.gamePool.JoinGame(code, playerName)
	if err != nil {
		return view.RenderView(c, view.ViewError(err.Error())) // todo add error to site
	}

	url := fmt.Sprintf("/lobby/%v/%s", code, p.Uid)
	c.Response().Header().Set("HX-Redirect", url) //HX-Redirect to url
	return c.NoContent(http.StatusOK)
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

	return view.RenderView(c, view.ViewLobby(g, p))
}

// e.POST("/leave", s.leaveHandler)
func (s *Session) leaveHandler(c echo.Context) error {
	gid := c.Param("id")
	pid := c.Param("player")

	return view.RenderView(c, view.LeavePopup(gid, pid))
}

// e.POST("/leave-confirmed/:id/:player", s.leaveConfirmedHandler)
func (s *Session) leaveConfirmedHandler(c echo.Context) error {
	gid := c.Param("id")
	pid := c.Param("player")

	err := s.gamePool.RemoveFromGame(gid, pid)
	if err != nil {
		return view.RenderView(c, view.ViewError(err.Error())) // todo add error to site
	}

	c.Response().Header().Set("HX-Redirect", "/") //HX-Redirect to home
	return c.NoContent(http.StatusOK)
}

// e.POST("/vote/:id/:originPid/:destPlayer", s.initVoteHandler)
func (s *Session) initVoteHandler(c echo.Context) error {
	gid := c.Param("id")
	destPid := c.Param("destPid")
	originPid := c.Param("originPid")

	destPlayer, err := s.gamePool.FindPlayer(gid, destPid)
	if err != nil {
		return view.RenderView(c, view.ViewError(err.Error())) // todo add error to site
	}
	originPlayer, err := s.gamePool.FindPlayer(gid, originPid)
	if err != nil {
		return view.RenderView(c, view.ViewError(err.Error())) // todo add error to site
	}

	g := s.gamePool.FindGame(gid)
	if g == nil {
		errMsg := fmt.Sprintf("game with id %v does not exist", gid)
		return view.RenderView(c, view.ViewError(errMsg)) // todo add error to site
	}

	v, err := g.NewVote(originPlayer, destPlayer)
	if err != nil {
		if v != nil {
			// vote already exists
			return view.RenderVote(c, v.OriginPlayer == originPlayer, gid, v.Votes[originPid], originPid, v.DestPlayer)
		}

		return view.RenderView(c, view.ViewError(err.Error()))
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
		return view.RenderView(c, view.ViewError(err.Error())) // todo add error to site
	}

	return view.RenderView(c, view.KillPopup(gid, p))
}

// e.POST("/kill-confirmed/:id/:player", s.killConfirmedHandler)
func (s *Session) killConfirmedHandler(c echo.Context) error {
	gid := c.Param("id")
	pid := c.Param("player")
	p, err := s.gamePool.FindPlayer(gid, pid)
	if err != nil {
		return view.RenderView(c, view.ViewError(err.Error())) // todo add error to site
	}

	pName := p.Name
	err = s.gamePool.RemoveFromGame(gid, pid)
	if err != nil {
		return view.RenderView(c, view.ViewError(err.Error())) // todo add error to site
	}
	return view.RenderView(c, view.KillConfirmPopup(pName))
}
