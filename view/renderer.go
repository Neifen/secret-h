package view

import (
	"bytes"
	"context"
	"github.com/a-h/templ"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

func ClosePopup(c echo.Context) error {
	return renderView(c, closePopup())
}

func renderView(c echo.Context, cmp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	if c.Request().Header.Get("HX-Request") != "true" {
		return cmp.Render(c.Request().Context(), c.Response().Writer)
		//// whole page
		//return BuildBase(nil, nil, cmp).Render(c.Request().Context(), c.Response().Writer)
	}

	return cmp.Render(c.Request().Context(), c.Response().Writer)
}

func renderWebsocket(ws *websocket.Conn, cmp templ.Component) error {
	//c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)

	var buf bytes.Buffer
	err := cmp.Render(context.Background(), &buf)
	if err != nil {
		return err
	}
	err = ws.WriteMessage(websocket.TextMessage, buf.Bytes())
	if err != nil {
		return err
	}
	return nil
}
