package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"log/slog"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/k8s"
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/models"
	"github.com/labstack/echo/v5"
)

// mockStore is a test double for k8s.InstanceStore.
type mockStore struct {
	ListInstancesFn          func(ctx context.Context) ([]models.RedisInstance, error)
	GetInstanceFn            func(ctx context.Context, id string) (*models.RedisInstance, error)
	CreateInstanceFn         func(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error)
	UpdateInstanceCapacityFn func(ctx context.Context, id string, req models.UpdateInstanceCapacityRequest) (*models.RedisInstance, error)
	DeleteInstanceFn         func(ctx context.Context, id string) error
}

func (m *mockStore) ListInstances(ctx context.Context) ([]models.RedisInstance, error) {
	if m.ListInstancesFn == nil {
		return nil, nil
	}
	return m.ListInstancesFn(ctx)
}

func (m *mockStore) GetInstance(ctx context.Context, id string) (*models.RedisInstance, error) {
	if m.GetInstanceFn == nil {
		return nil, nil
	}
	return m.GetInstanceFn(ctx, id)
}

func (m *mockStore) CreateInstance(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error) {
	if m.CreateInstanceFn == nil {
		return nil, nil
	}
	return m.CreateInstanceFn(ctx, req)
}

func (m *mockStore) UpdateInstanceCapacity(ctx context.Context, id string, req models.UpdateInstanceCapacityRequest) (*models.RedisInstance, error) {
	if m.UpdateInstanceCapacityFn == nil {
		return nil, nil
	}
	return m.UpdateInstanceCapacityFn(ctx, id, req)
}

func (m *mockStore) DeleteInstance(ctx context.Context, id string) error {
	if m.DeleteInstanceFn == nil {
		return nil
	}
	return m.DeleteInstanceFn(ctx, id)
}

// newTestApp creates an Application with a mock store and a no-op logger.
func newTestApp(store k8s.InstanceStore) *Application {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	return NewApplication(store, logger)
}

func newEcho() *echo.Echo {
	e := echo.New()
	return e
}

// POST
func TestCreateInstance_Handler(t *testing.T) {
	tests := []struct {
		name           string
		body           any
		mockStore      *mockStore
		wantStatusCode int
	}{
		{
			name: "success",
			body: map[string]any{
				"name":     "test-redis",
				"capacity": "1Gi",
			},
			mockStore: &mockStore{
				CreateInstanceFn: func(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error) {
					return &models.RedisInstance{
						ID:                "test-redis",
						Name:              "test-redis",
						Namespace:         "test",
						Status:            "running",
						Capacity:          "1Gi",
						RedisReplicas:     3,
						SentinelReplicas:  3,
						PublicServiceName: "test-redis-public",
						PublicHostname:    "",
						PublicPort:        6379,
						PublicEndpoint:    "",
					}, nil
				},
			},
			wantStatusCode: http.StatusCreated,
		},
		{
			name: "missing required fields returns 400",
			body: map[string]any{
				"name": "",
			},
			mockStore:      &mockStore{},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name: "store error returns 500",
			body: map[string]any{
				"name":     "test-redis",
				"capacity": "1Gi",
			},
			mockStore: &mockStore{
				CreateInstanceFn: func(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error) {
					return nil, errors.New("backend failure")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEcho()
			app := newTestApp(tt.mockStore)

			bodyBytes, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("marshal body: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/instances", bytes.NewReader(bodyBytes))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			// Register route so Echo builds the context and invokes the handler as in real usage.
			e.POST("/api/v1/instances", app.CreateInstance)
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Fatalf("unexpected status code: got %d, want %d; body=%s", rec.Code, tt.wantStatusCode, rec.Body.String())
			}
		})
	}
}

// GET single Instance
func TestGetInstance_Handler(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockStore      *mockStore
		wantStatusCode int
	}{
		{
			name: "success",
			id:   "test-redis",
			mockStore: &mockStore{
				GetInstanceFn: func(ctx context.Context, id string) (*models.RedisInstance, error) {
					return &models.RedisInstance{
						ID:                "test-redis",
						Name:              "test-redis",
						Namespace:         "test",
						Status:            "running",
						Capacity:          "1Gi",
						RedisReplicas:     3,
						SentinelReplicas:  3,
						PublicServiceName: "test-redis-public",
						PublicHostname:    "test-redis-public.test.svc.cluster.local",
						PublicPort:        6379,
						PublicEndpoint:    "test-redis-public.test.svc.cluster.local:6379",
					}, nil
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "not found",
			id:   "missing",
			mockStore: &mockStore{
				GetInstanceFn: func(ctx context.Context, id string) (*models.RedisInstance, error) {
					return nil, k8s.ErrNotFound
				},
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "store error returns 500",
			id:   "test-redis",
			mockStore: &mockStore{
				GetInstanceFn: func(ctx context.Context, id string) (*models.RedisInstance, error) {
					return nil, errors.New("backend failure")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEcho()
			app := newTestApp(tt.mockStore)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/instances/"+tt.id, nil)
			rec := httptest.NewRecorder()

			e.GET("/api/v1/instances/:id", app.GetInstance)
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Fatalf("unexpected status code: got %d, want %d; body=%s", rec.Code, tt.wantStatusCode, rec.Body.String())
			}
		})
	}
}

// GET List of Instances
func TestListInstances_Handler(t *testing.T) {
	tests := []struct {
		name           string
		mockStore      *mockStore
		wantStatusCode int
	}{
		{
			name: "success with instances",
			mockStore: &mockStore{
				ListInstancesFn: func(ctx context.Context) ([]models.RedisInstance, error) {
					return []models.RedisInstance{
						{
							ID:                "redis-1",
							Name:              "redis-1",
							Namespace:         "test",
							Status:            "running",
							Capacity:          "1Gi",
							RedisReplicas:     3,
							SentinelReplicas:  3,
							PublicServiceName: "redis-1-public",
							PublicHostname:    "",
							PublicPort:        6379,
							PublicEndpoint:    "",
						},
					}, nil
				},
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name: "store error returns 500",
			mockStore: &mockStore{
				ListInstancesFn: func(ctx context.Context) ([]models.RedisInstance, error) {
					return nil, errors.New("backend failure")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEcho()
			app := newTestApp(tt.mockStore)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/instances", nil)
			rec := httptest.NewRecorder()

			e.GET("/api/v1/instances", app.ListInstances)
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Fatalf("unexpected status code: got %d, want %d; body=%s", rec.Code, tt.wantStatusCode, rec.Body.String())
			}
		})
	}
}

// DELETE single Instance
func TestDeleteInstance_Handler(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockStore      *mockStore
		wantStatusCode int
	}{
		{
			name: "success no content",
			id:   "redis-1",
			mockStore: &mockStore{
				DeleteInstanceFn: func(ctx context.Context, id string) error {
					return nil
				},
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name: "not found",
			id:   "missing",
			mockStore: &mockStore{
				DeleteInstanceFn: func(ctx context.Context, id string) error {
					return k8s.ErrNotFound
				},
			},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name: "store error returns 500",
			id:   "redis-1",
			mockStore: &mockStore{
				DeleteInstanceFn: func(ctx context.Context, id string) error {
					return errors.New("backend failure")
				},
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEcho()
			app := newTestApp(tt.mockStore)

			req := httptest.NewRequest(http.MethodDelete, "/api/v1/instances/"+tt.id, nil)
			rec := httptest.NewRecorder()

			e.DELETE("/api/v1/instances/:id", app.DeleteInstance)
			e.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatusCode {
				t.Fatalf("unexpected status code: got %d, want %d; body=%s", rec.Code, tt.wantStatusCode, rec.Body.String())
			}
		})
	}
}
