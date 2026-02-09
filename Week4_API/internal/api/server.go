package api

import (
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/k8s"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

func NewServer(cfg *Config, store k8s.InstanceStore) *echo.Echo {
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	app := NewApplication(store, e.Logger)
	RegisterRoutes(e, app)
	return e
}
