package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/cache"
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/k8s"
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/logstore"
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
	LogStore    logstore.Store // nil if DATABASE_URL unset; audit/service log writes and ListLogs no-op or skip
	Logger      *slog.Logger
	// lastInstanceStatus is used to detect status changes for service log (key: instanceID, value: status).
	lastInstanceStatus   map[string]string
	lastInstanceStatusMu sync.RWMutex
}

func NewApplication(store k8s.InstanceStore, cacheClient cache.ClientInterface, logStore logstore.Store, logger *slog.Logger) *Application {
	if cacheClient == nil {
		cacheClient = cache.NewClient()
	}
	return &Application{
		Store:                store,
		CacheClient:          cacheClient,
		LogStore:             logStore,
		Logger:               logger,
		lastInstanceStatus:   make(map[string]string),
	}
}

// writeAuditLog appends an audit log entry in the background. No-op if LogStore is nil. Errors are logged only.
func (a *Application) writeAuditLog(ctx context.Context, tenantUser, instanceID, action string, details map[string]any) {
	if a.LogStore == nil {
		return
	}
	go func() {
		if err := a.LogStore.AppendAuditLog(context.Background(), tenantUser, instanceID, action, details); err != nil {
			a.Logger.Error("audit log write failed", "instanceId", instanceID, "action", action, "error", err)
		}
	}()
}

// writeServiceLog appends a service log entry in the background. No-op if LogStore is nil. Errors are logged only.
func (a *Application) writeServiceLog(ctx context.Context, tenantUser, instanceID, eventType, message string, metadata map[string]any) {
	if a.LogStore == nil {
		return
	}
	go func() {
		if err := a.LogStore.AppendServiceLog(context.Background(), tenantUser, instanceID, eventType, message, metadata); err != nil {
			a.Logger.Error("service log write failed", "instanceId", instanceID, "eventType", eventType, "error", err)
		}
	}()
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

	// Service log: record status change (best-effort)
	a.lastInstanceStatusMu.Lock()
	key := user + "/" + id
	prev := a.lastInstanceStatus[key]
	if prev != instance.Status {
		a.lastInstanceStatus[key] = instance.Status
		a.lastInstanceStatusMu.Unlock()
		if prev != "" {
			a.writeServiceLog(ctx, user, id, "status_change", "Instance status changed to "+instance.Status, map[string]any{"previous": prev})
		}
	} else {
		a.lastInstanceStatusMu.Unlock()
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
	a.writeAuditLog(ctx, user, instance.ID, "create", map[string]any{
		"name":             req.Name,
		"capacity":         req.Capacity,
		"redisReplicas":    instance.RedisReplicas,
		"sentinelReplicas": instance.SentinelReplicas,
	})
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
	details := map[string]any{}
	if req.Name != nil {
		details["name"] = *req.Name
	}
	if req.Capacity != nil {
		details["capacity"] = *req.Capacity
	}
	if req.RedisReplicas != nil {
		details["redisReplicas"] = *req.RedisReplicas
	}
	if req.SentinelReplicas != nil {
		details["sentinelReplicas"] = *req.SentinelReplicas
	}
	a.writeAuditLog(ctx, user, id, "update", details)
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
	a.writeAuditLog(ctx, user, id, "delete", map[string]any{"instanceId": id})
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
	details := map[string]any{"key": req.Key}
	if req.TTLSeconds > 0 {
		details["ttlSeconds"] = req.TTLSeconds
	}
	a.writeAuditLog(ctx, user, id, "cache_set", details)
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
	a.writeAuditLog(ctx, user, id, "cache_get", map[string]any{"key": key})
	return c.JSON(http.StatusOK, models.GetCacheResponse{Key: key, Value: value})
}

// ListLogs returns audit and/or service logs for the instance. Instance must belong to the tenant (X-User).
func (a *Application) ListLogs(c *echo.Context) error {
	id := c.Param("id")
	user := c.Request().Header.Get("X-User")
	ns := namespaceForUser(user)
	if ns == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "missing or empty X-User header"})
	}
	if a.LogStore == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "log store not configured"})
	}
	ctx := k8s.WithNamespace(c.Request().Context(), ns)

	// Ensure instance exists and belongs to tenant
	_, err := a.Store.GetInstance(ctx, id)
	if err != nil {
		if errors.Is(err, k8s.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "instance not found"})
		}
		a.Logger.Error("get instance for logs failed", "id", id, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get instance"})
	}

	opts := logstore.ListOpts{}
	if t := c.QueryParam("type"); t != "" {
		opts.Type = t
	}
	if sinceStr := c.QueryParam("since"); sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			opts.Since = t
		}
	}
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if n, err := parseInt(limitStr); err == nil && n > 0 {
			opts.Limit = n
		}
	}

	entries, err := a.LogStore.ListLogs(ctx, user, id, opts)
	if err != nil {
		a.Logger.Error("list logs failed", "id", id, "error", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list logs"})
	}
	// Map to API model for consistent response shape
	out := make([]models.LogEntry, len(entries))
	for i := range entries {
		out[i] = models.LogEntry{
			ID:         entries[i].ID,
			Type:       entries[i].Type,
			Timestamp:  entries[i].Timestamp,
			Action:     entries[i].Action,
			Message:    entries[i].Message,
			Details:    entries[i].Details,
			Metadata:   entries[i].Metadata,
			TenantUser: entries[i].TenantUser,
			InstanceID: entries[i].InstanceID,
		}
	}
	return c.JSON(http.StatusOK, out)
}

// parseInt parses a decimal integer; returns 0 and non-nil error on failure.
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
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
