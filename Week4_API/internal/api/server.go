package api

import (
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/k8s"
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/logstore"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func NewServer(cfg *Config, store k8s.InstanceStore, logStore logstore.Store) *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	app := NewApplication(store, nil, logStore, e.Logger)
	RegisterRoutes(e, app)
	return e
}
