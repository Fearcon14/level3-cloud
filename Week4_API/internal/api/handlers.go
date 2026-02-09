package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/k8s"
	"github.com/labstack/echo/v5"
)

// Application holds dependencies for our handlers (e.g. K8s client, Logger etc.)
type Application struct {
	Store  k8s.InstanceStore
	Logger *slog.Logger
}

func NewApplication(store k8s.InstanceStore, logger *slog.Logger) *Application {
	return &Application {
		Store: store,
		Logger: logger,
	}
}

func (a *Application) ListInstances(c echo.Context) error {
	ctx := c.Request().Context()
	instances, err := a.Store.ListInstances(ctx)
	if err != nil {
		a.Logger.Error("failed to list instances", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Errorf("failed to list instances").Error()})
	}
	return c.JSON(http.StatusOK, instances)
}
