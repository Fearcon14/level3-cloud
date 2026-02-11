package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

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
	return &Application{
		Store:  store,
		Logger: logger,
	}
}

// namespaceForUser maps a user identifier (e.g. from X-User header) to a Kubernetes namespace.
// Example: "alice" -> "tenant-alice".
func namespaceForUser(user string) string {
	user = strings.TrimSpace(strings.ToLower(user))
	if user == "" {
		return ""
	}
	return "tenant-" + user
}

// ListInstances returns a list of all Redis instances in the store's namespace.
func (a *Application) ListInstances(c *echo.Context) error {
	user := c.Request().Header.Get("X-User")
	ns := namespaceForUser(user)
	if ns == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing or empty X-User header"})
	}
	ctx := k8s.WithNamespace(c.Request().Context(), ns)

	instances, err := a.Store.ListInstances(ctx)
	if err != nil {
		a.Logger.Error("failed to list instances", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Errorf("failed to list instances").Error()})
	}
	return c.JSON(http.StatusOK, instances)
}

// GetInstance returns a single Redis instance by name (id). Returns 404 if not found.
func (a *Application) GetInstance(c *echo.Context) error {
	id := c.Param("id")
	user := c.Request().Header.Get("X-User")
	ns := namespaceForUser(user)
	if ns == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing or empty X-User header"})
	}
	ctx := k8s.WithNamespace(c.Request().Context(), ns)

	instance, err := a.Store.GetInstance(ctx, id)
	if err != nil {
		if errors.Is(err, k8s.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "instance not found"})
		}
		a.Logger.Error("get instance failed", "id", id, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get instance"})
	}
	return c.JSON(http.StatusOK, instance)
}

// CreateInstance creates a new Redis instance from the request.
func (a *Application) CreateInstance(c *echo.Context) error {
	var req models.CreateRedisRequest

	if err := c.Bind(&req); err != nil {
		a.Logger.Error("failed to bind request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Errorf("failed to bind request").Error()})
	}

	if req.Name == "" || req.Capacity == "" {
		a.Logger.Error("missing required fields")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name and capacity are required"})
	}

	user := c.Request().Header.Get("X-User")
	ns := namespaceForUser(user)
	if ns == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing or empty X-User header"})
	}
	ctx := k8s.WithNamespace(c.Request().Context(), ns)

	instance, err := a.Store.CreateInstance(ctx, req)
	if err != nil {
		a.Logger.Error("failed to create instance", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": fmt.Sprintf("failed to create instance: %v", err)})
	}
	return c.JSON(http.StatusCreated, instance)
}

// UpdateInstanceCapacity updates the capacity (and optionally StorageClass) of an existing Redis instance.
func (a *Application) UpdateInstanceCapacity(c *echo.Context) error {
	id := c.Param("id")
	var req models.UpdateInstanceCapacityRequest
	if err := c.Bind(&req); err != nil {
		a.Logger.Error("failed to bind request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Capacity == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "capacity is required"})
	}

	user := c.Request().Header.Get("X-User")
	ns := namespaceForUser(user)
	if ns == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing or empty X-User header"})
	}
	ctx := k8s.WithNamespace(c.Request().Context(), ns)
	updated, err := a.Store.UpdateInstanceCapacity(ctx, id, req)
	if err != nil {
		if errors.Is(err, k8s.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "instance not found"})
		}
		a.Logger.Error("failed to update instance", "id", id, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update instance"})
	}
	return c.JSON(http.StatusOK, updated)
}

// DeleteInstance deletes an existing Redis instance. Returns 404 if the instance does not exist.
func (a *Application) DeleteInstance(c *echo.Context) error {
	id := c.Param("id")
	user := c.Request().Header.Get("X-User")
	ns := namespaceForUser(user)
	if ns == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing or empty X-User header"})
	}
	ctx := k8s.WithNamespace(c.Request().Context(), ns)
	if err := a.Store.DeleteInstance(ctx, id); err != nil {
		if errors.Is(err, k8s.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "instance not found"})
		}
		a.Logger.Error("failed to delete instance", "id", id, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete instance"})
	}
	return c.NoContent(http.StatusNoContent)
}
