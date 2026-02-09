package api

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/k8s"
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/models"
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

// ListInstances returns a list of all Redis instances in the store's namespace.
func (a *Application) ListInstances(c echo.Context) error {
	ctx := c.Request().Context()

	instances, err := a.Store.ListInstances(ctx)
	if err != nil {
		a.Logger.Error("failed to list instances", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Errorf("failed to list instances").Error()})
	}
	return c.JSON(http.StatusOK, instances)
}

// GetInstance returns a single Redis instance by name (id). Returns an error if not found.
func (a *Application) GetInstance(c echo.Context) error {
	id := c.Param("id")
	ctx := c.Request().Context()

	instance, err := a.Store.GetInstance(ctx, id)
	if err != nil {
		a.Logger.Error("intance not found", "error", err)
		return c.JSON(http.StatusNotFound, map[string]string{"error": fmt.Errorf("instance not found").Error()})
	}
	return c.JSON(http.StatusOK, instance)
}

// CreateInstance creates a new Redis instance from the request.
func (a *Application) CreateInstance(c echo.Context) error {
	var req models.CreateRedisRequest

	if err := c.Bind(&req); err != nil {
		a.Logger.Error("failed to bind request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Errorf("failed to bind request").Error()})
	}

	if req.Name == "" || req.Capacity == "" {
		a.Logger.Error("missing required fields")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name and capacity are required"})
	}

	ctx := c.Request().Context()

	instance, err := a.Store.CreateInstance(ctx, req)
	if err != nil {
		a.Logger.Error("failed to create instance", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Errorf("failed to create instance").Error()})
	}
	return c.JSON(http.StatusCreated, instance)
}

// UpdateInstanceCapacity updates the capacity of an existing Redis instance.
func (a *Application) UpdateInstanceCapacity(c echo.Context) error {
	id := c.Param("id")
	var req models.UpdateInstanceCapacityRequest
	if err := c.Bind(&req); err != nil {
		a.Logger.Error("failed to bind request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	ctx := c.Request().Context()
	updated, err := a.Store.UpdateInstanceCapacity(ctx, id, req.Capacity)
	if err != nil {
		a.Logger.Error("failed to update instance", "id", id, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Errorf("failed to update instance").Error()})
	}
	return c.JSON(http.StatusOK, updated)
}


