package api

import (
	"fmt"
	"github.com/Neifen/secret-h/view"
	"github.com/labstack/echo/v4"
	"net/http"
)

// e.POST("/start", s.startHandler)
func (s *Session) startHandler(c echo.Context) error {
	playerName := c.FormValue("name")
	if playerName == "" {
		return view.RenderMessage(c, "Field \"Name\" is a required field")
	}

	code, p, err := s.gamePool.StartGame(playerName)
	if err != nil {
		return view.RenderError(c, err)
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
		return view.RenderError(c, err)
	}

	url := fmt.Sprintf("/lobby/%v/%s", code, p.Uid)
	c.Response().Header().Set("HX-Redirect", url) //HX-Redirect to url
	return c.NoContent(http.StatusOK)
}

// e.POST("/leave", s.leaveHandler)
func (s *Session) leaveHandler(c echo.Context) error {
	gid := c.Param("id")
	pid := c.Param("player")

	return view.RenderLeavePopup(c, gid, pid)
}

// e.POST("/leave-confirmed/:id/:player", s.leaveConfirmedHandler)
func (s *Session) leaveConfirmedHandler(c echo.Context) error {
	gid := c.Param("id")
	pid := c.Param("player")

	err := s.gamePool.RemoveFromGame(gid, pid, false)
	if err != nil {
		return view.RenderError(c, err)
	}

	c.Response().Header().Set("HX-Redirect", "/") //HX-Redirect to home
	return c.NoContent(http.StatusOK)
}
