package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/cache"
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/k8s"
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

// MaxCacheKeyLength is the maximum allowed length for a cache key (Redis best practice).
const MaxCacheKeyLength = 512

// Application holds dependencies for our handlers (e.g. K8s client, Logger etc.)
type Application struct {
	Store       k8s.InstanceStore
	CacheClient cache.ClientInterface
	Logger      *slog.Logger
}

func NewApplication(store k8s.InstanceStore, cacheClient cache.ClientInterface, logger *slog.Logger) *Application {
	if cacheClient == nil {
		cacheClient = cache.NewClient()
	}
	return &Application{
		Store:       store,
		CacheClient: cacheClient,
		Logger:      logger,
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

// PatchInstance applies a partial update to an existing Redis instance.
// It can update the display name, Redis replicas, Sentinel replicas, and capacity (PVC size).
func (a *Application) PatchInstance(c *echo.Context) error {
	id := c.Param("id")
	var req models.PatchInstanceRequest
	if err := c.Bind(&req); err != nil {
		a.Logger.Error("failed to bind request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Basic guard: ensure at least one field is provided.
	if req.Name == nil && req.Capacity == nil && req.RedisReplicas == nil && req.SentinelReplicas == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "at least one field must be provided"})
	}

	user := c.Request().Header.Get("X-User")
	ns := namespaceForUser(user)
	if ns == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing or empty X-User header"})
	}
	ctx := k8s.WithNamespace(c.Request().Context(), ns)
	updated, err := a.Store.PatchInstance(ctx, id, req)
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

// SetCache stores a key-value pair in the Redis instance's cache (POST /instances/:id/cache).
// Request body: { "key": "...", "value": "...", "ttlSeconds": 0 (optional) }.
// Returns 400 if key/value are missing or key is too long; 404 if instance not found; 503 if Redis is unreachable.
func (a *Application) SetCache(c *echo.Context) error {
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
		a.Logger.Error("get instance for cache set failed", "id", id, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get instance"})
	}
	if instance.PublicEndpoint == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "instance has no public endpoint (not ready)"})
	}

	var req models.SetCacheRequest
	if err := c.Bind(&req); err != nil {
		a.Logger.Error("failed to bind set cache request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}
	if req.Key == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "key is required"})
	}
	if len(req.Key) > MaxCacheKeyLength {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("key must be at most %d bytes", MaxCacheKeyLength)})
	}
	if req.TTLSeconds < 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ttlSeconds must be non-negative"})
	}

	if err := a.CacheClient.Set(ctx, instance.PublicEndpoint, instance.Password, cache.SetOptions{
		Key:        req.Key,
		Value:      req.Value,
		TTLSeconds: req.TTLSeconds,
	}); err != nil {
		a.Logger.Error("cache set failed", "id", id, "key", req.Key, "error", err)
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "failed to store value in cache"})
	}
	return c.JSON(http.StatusOK, models.GetCacheResponse{Key: req.Key, Value: req.Value})
}

// GetCache returns the value for a key from the Redis instance's cache (GET /instances/:id/cache/:key).
// Returns 404 if the instance or the key does not exist; 503 if Redis is unreachable.
func (a *Application) GetCache(c *echo.Context) error {
	id := c.Param("id")
	key := c.Param("key")
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
		a.Logger.Error("get instance for cache get failed", "id", id, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get instance"})
	}
	if instance.PublicEndpoint == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "instance has no public endpoint (not ready)"})
	}
	if key == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "key is required"})
	}

	value, err := a.CacheClient.Get(ctx, instance.PublicEndpoint, instance.Password, key)
	if err != nil {
		if errors.Is(err, cache.ErrKeyNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "cache key not found"})
		}
		a.Logger.Error("cache get failed", "id", id, "key", key, "error", err)
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "failed to get value from cache"})
	}
	return c.JSON(http.StatusOK, models.GetCacheResponse{Key: key, Value: value})
}

// Login handles user authentication and issues a JWT
func (a *Application) Login(c *echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		a.Logger.Error("failed to bind request", "error", err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Hardcoded auth check
	// TODO: Implement proper auth check
	if req.Username != "kevin" || req.Password != "KevinsPassword" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid credentials"})
	}

	// Create claims
	claims := jwt.MapClaims{
		"sub": req.Username,
		"exp": time.Now().Add(time.Hour * 72).Unix(), // Token expires after 72 hours
	}

	// Create Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	// TODO: Load secret from env
	t, err := token.SignedString([]byte("kevins-super-secret-key"))
	if err != nil {
		a.Logger.Error("failed to sign token", "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to sign token"})
	}

	return c.JSON(http.StatusOK, models.LoginResponse{Token: t})
}
