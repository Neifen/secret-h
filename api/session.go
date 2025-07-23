package api

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"secret-h/game"
)

type Session struct {
	gamePool *game.GamePool
}

func NewSession() *Session {
	return &Session{
		gamePool: game.NewGamePool(),
	}
}

func (s *Session) Start() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Static("static", "assets")

	e.GET("/", s.homeHandler)
	e.POST("/start", s.startHandler)
	e.POST("/join", s.joinHandler)
	e.POST("/leave/:id/:player", s.leaveHandler)
	e.POST("/leave-confirmed/:id/:player", s.leaveConfirmedHandler)

	e.GET("/lobby/:id/:player", s.lobbyHandler)
	e.POST("/kill/:id/:player", s.initKillHandler)
	e.POST("/kill-confirmed/:id/:player", s.killConfirmedHandler)
	e.POST("/vote/:id/:originPid/:destPid", s.initVoteHandler)
	e.POST("/cancel-vote/:id", s.cancelVoteHandler)
	e.POST("/finish-vote/:id/:originPid/:destPid", s.finishVoteHandler)
	e.POST("/make-vote/:id/:originPid/:destPid", s.makeVoteHandler)
	e.POST("/closePopup", s.closePopupHandler)

	err := e.Start(":3000")
	if err != nil {
		log.Fatalln("could not start to listen to to error ", err)
	}
}
