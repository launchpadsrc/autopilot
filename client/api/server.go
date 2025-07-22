package api

import (
	"github.com/labstack/echo/v4"

	"launchpad.icu/autopilot/internal/database"
)

type Config struct {
	Addr string
	DB   *database.DB
}

type Server struct {
	addr string
	e    *echo.Echo
	db   *database.DB
}

func New(c Config) Server {
	if c.Addr == "" {
		c.Addr = ":8080"
	}

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	return Server{
		addr: c.Addr,
		e:    e,
		db:   c.DB,
	}
}

func (s Server) Start() error {
	s.addRoutes()

	return s.e.Start(s.addr)
}

func (s Server) addRoutes() {
	s.e.GET("/users/:id/profile", s.usersProfile)
}
