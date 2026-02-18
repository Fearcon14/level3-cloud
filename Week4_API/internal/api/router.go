package api

import (
	"github.com/labstack/echo/v5"
)

func RegisterRoutes(e *echo.Echo, app *Application) {
	// Public Routes
	e.POST("/api/login", app.Login)

	// Protected Routes
	v1 := e.Group("/api/v1")
	v1.Use(JWTMiddleware)

	v1.GET("/instances", app.ListInstances)
	v1.GET("/instances/:id", app.GetInstance)
	v1.POST("/instances", app.CreateInstance)
	v1.PATCH("/instances/:id", app.PatchInstance)
	v1.DELETE("/instances/:id", app.DeleteInstance)

	// Cache: more specific route first so :id does not capture "cache"
	v1.GET("/instances/:id/cache/:key", app.GetCache)
	v1.POST("/instances/:id/cache", app.SetCache)
}
