package api

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"secret-h/view"
)

// e.GET("/ws/:id/:player", s.wsHandler)
var upgrader = websocket.Upgrader{} // use default options
func (s *Session) wsHandler(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return view.RenderError(c, err)
	}

	pid := c.Param("player")
	gid := c.Param("id")
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
