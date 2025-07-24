package api

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"secret-h/view"
)

// e.GET("/", s.homeHandler)
func (s *Session) homeHandler(c echo.Context) error {
	c.Response().Header().Set("HX-Refresh", "true")
	return view.RenderViewHome(c)
}

func redirectHome(c echo.Context) error {
	if c.Request().Header.Get("Hx-Request") != "true" {
		return c.Redirect(http.StatusSeeOther, "/")
	}

	c.Response().Header().Set("HX-Redirect", "/") //HX-Redirect to url
	return c.NoContent(http.StatusOK)
}
