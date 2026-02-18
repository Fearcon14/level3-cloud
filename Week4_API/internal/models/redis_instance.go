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

	// Connection / access data (in-cluster DNS; use from pods in the same cluster or via port-forward).
	PublicServiceName string `json:"publicServiceName"` // e.g. "<name>-redis"
	PublicHostname    string `json:"publicHostname"`    // e.g. "<name>-redis.default.svc.cluster.local"
	PublicPort        int    `json:"publicPort"`        // 6379
	PublicEndpoint    string `json:"publicEndpoint"`    // host:port for Redis clients
	Password          string `json:"password,omitempty"` // Password for Redis authentication
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

// PatchInstanceRequest is the body for PATCH /instances/:id (partial update).
// All fields are optional; only the provided fields will be updated.
type PatchInstanceRequest struct {
	// New display name for the instance (stored as an annotation on the RedisFailover resource).
	Name *string `json:"name,omitempty"`

	// New storage capacity (PVC size), e.g. "20Gi".
	Capacity *string `json:"capacity,omitempty"`

	// New number of Redis replicas.
	RedisReplicas *int `json:"redisReplicas,omitempty"`

	// New number of Sentinel replicas.
	SentinelReplicas *int `json:"sentinelReplicas,omitempty"`
}
