package models

// RedisInstance represents a Redis instance (backed by a RedisFailover CR).
// Read-only fields (e.g. RedisReplicas, SentinelReplicas) are filled from the CR when listing/getting.
type RedisInstance struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	Status           string `json:"status"`
	Capacity         string `json:"capacity"`
	RedisReplicas    int    `json:"redisReplicas,omitempty"`
	SentinelReplicas int    `json:"sentinelReplicas,omitempty"`

	// Connection / access data for public connectivity.
	PublicServiceName string `json:"publicServiceName"`
	PublicHostname    string `json:"publicHostname"`
	PublicPort        int    `json:"publicPort"`
	PublicEndpoint    string `json:"publicEndpoint"`
}

// CreateRedisRequest is the body for POST /instances
// Optional fields use pointers or empty string; the backend applies defaults when not set.
type CreateRedisRequest struct {
	Name     string `json:"name" validate:"required"`
	Capacity string `json:"capacity" validate:"required"`

	// Optional: replicas (defaults applied by backend if not set)
	RedisReplicas    *int `json:"redisReplicas,omitempty"`
	SentinelReplicas *int `json:"sentinelReplicas,omitempty"`

	// Optional: storage
	StorageClass string `json:"storageClass,omitempty"` // (e.g. "premium-perf1-stackit")

	// Optional: redis resources (defaults applied by backend if not set)
	CPURequest    string `json:"cpuRequest,omitempty"`    // e.g. "100m"
	MemoryRequest string `json:"memoryRequest,omitempty"` // e.g. "128Mi"
	CPULimit      string `json:"cpuLimit,omitempty"`      // e.g. "500m"
	MemoryLimit   string `json:"memoryLimit,omitempty"`    // e.g. "512Mi"
}

// UpdateInstanceCapacityRequest is the body for PUT /instances/:id (update capacity/storage).
type UpdateInstanceCapacityRequest struct {
	Capacity      string `json:"capacity"`
	StorageClass string `json:"storageClass,omitempty"`
}
