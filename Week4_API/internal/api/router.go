package api

import (
	"github.com/labstack/echo/v5"
)

func RegisterRoutes(e *echo.Echo, app *Application) {
	v1 := e.Group("/api/v1")

	v1.GET("/instances", app.ListInstances)
	v1.GET("/instances/:id", app.GetInstance)
	v1.POST("/instances", app.CreateInstance)
	v1.PUT("/instances/:id/capacity", app.UpdateInstanceCapacity)
	v1.DELETE("/instances/:id", app.DeleteInstance)
}
