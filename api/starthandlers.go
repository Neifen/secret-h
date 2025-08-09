package api

import (
	"fmt"
	"github.com/Neifen/secret-h/view"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
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

	setCookies(c, code, p.Uid)
	c.Response().Header().Set("HX-Redirect", url) //HX-Redirect to url
	return c.NoContent(http.StatusOK)
}

// e.GET("/join-qr/:id", s.joinQrHandler)
func (s *Session) joinQrHandler(c echo.Context) error {
	gid := c.Param("id")

	return view.RenderViewJoinQr(c, gid)
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

	setCookies(c, code, p.Uid)
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

	setCookies(c, gid, pid)
	c.Response().Header().Set("HX-Redirect", "/") //HX-Redirect to home
	return c.NoContent(http.StatusOK)
}

func setCookies(c echo.Context, gid, pid string) {
	c.SetCookie(&http.Cookie{Name: "gid", Value: gid, Path: "/"})
	c.SetCookie(&http.Cookie{Name: "pid", Value: pid, Path: "/"})
}

func deleteCookies(c echo.Context) {
	c.SetCookie(&http.Cookie{Name: "gid", Value: "", Path: "/", Expires: time.Unix(0, 0)})
	c.SetCookie(&http.Cookie{Name: "pid", Value: "", Path: "/", Expires: time.Unix(0, 0)})
}
