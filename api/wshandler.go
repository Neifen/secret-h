package api

import (
	"context"
	"fmt"
	"github.com/Neifen/secret-h/view"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"net/http"
)

// e.GET("/ws/:id/:player", s.wsHandler)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		err := r.Context().Value("error")
		if err != nil {
			return false
		}
		return true
	},
} // use default options
func (s *Session) wsHandler(c echo.Context) error {
	pid := c.Param("player")
	gid := c.Param("id")
	_, err := s.gamePool.FindPlayer(gid, pid)
	if err != nil {
		r := c.Request().WithContext(context.WithValue(c.Request().Context(), "error", err))
		c.SetRequest(r)
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return view.RenderError(c, err)
	}

	err = s.gamePool.SetPlayerWS(conn, gid, pid)
	if err != nil {
		return view.RenderError(c, err)
	}

	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("Websocket error: ", err.Error())
		}
	}(conn)

	for {
		_, mess, err := conn.ReadMessage()
		if err != nil {
			return view.RenderError(c, err)
		}

		fmt.Println("ws message ", string(mess))
	}
}
