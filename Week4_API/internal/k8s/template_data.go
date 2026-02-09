package k8s

import (
	"fmt"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/models"
	"k8s.io/apimachinery/pkg/api/resource"
)

// RedisFailoverTemplateData holds the values passed into the RedisFailover YAML template.
type RedisFailoverTemplateData struct {
	Name             string
	Namespace        string
	SentinelReplicas int
	RedisReplicas    int
	CPURequest       string
	MemoryRequest    string
	CPULimit         string
	MemoryLimit      string
	StorageClass     string
	StorageSize      string
}

const (
	defaultSentinelReplicas = 3
	defaultRedisReplicas    = 3
	defaultCPURequest      = "100m"
	defaultMemoryRequest   = "128Mi"
	defaultCPULimit        = "500m"
	defaultMemoryLimit     = "512Mi"
	defaultStorageClass    = "premium-perf1-stackit"
	defaultStorageSize     = "1Gi"

	minReplicas = 1
	maxReplicas = 9
)

// ValidateCreateRedisRequest validates replicas, storage, and resource fields before rendering or calling the API.
func ValidateCreateRedisRequest(req models.CreateRedisRequest) error {
	if req.RedisReplicas != nil {
		if n := *req.RedisReplicas; n < minReplicas || n > maxReplicas {
			return fmt.Errorf("redisReplicas must be between %d and %d, got %d", minReplicas, maxReplicas, n)
		}
	}
	if req.SentinelReplicas != nil {
		if n := *req.SentinelReplicas; n < minReplicas || n > maxReplicas {
			return fmt.Errorf("sentinelReplicas must be between %d and %d, got %d", minReplicas, maxReplicas, n)
		}
	}
	if req.Capacity != "" {
		if _, err := resource.ParseQuantity(req.Capacity); err != nil {
			return fmt.Errorf("capacity: invalid quantity %q: %w", req.Capacity, err)
		}
	}
	if req.CPURequest != "" {
		if _, err := resource.ParseQuantity(req.CPURequest); err != nil {
			return fmt.Errorf("cpuRequest: invalid quantity %q: %w", req.CPURequest, err)
		}
	}
	if req.MemoryRequest != "" {
		if _, err := resource.ParseQuantity(req.MemoryRequest); err != nil {
			return fmt.Errorf("memoryRequest: invalid quantity %q: %w", req.MemoryRequest, err)
		}
	}
	if req.CPULimit != "" {
		if _, err := resource.ParseQuantity(req.CPULimit); err != nil {
			return fmt.Errorf("cpuLimit: invalid quantity %q: %w", req.CPULimit, err)
		}
	}
	if req.MemoryLimit != "" {
		if _, err := resource.ParseQuantity(req.MemoryLimit); err != nil {
			return fmt.Errorf("memoryLimit: invalid quantity %q: %w", req.MemoryLimit, err)
		}
	}
	return nil
}

// ValidateUpdateInstanceCapacityRequest validates capacity (and optionally storageClass) before calling the API.
func ValidateUpdateInstanceCapacityRequest(req models.UpdateInstanceCapacityRequest) error {
	if req.Capacity == "" {
		return fmt.Errorf("capacity is required")
	}
	if _, err := resource.ParseQuantity(req.Capacity); err != nil {
		return fmt.Errorf("capacity: invalid quantity %q: %w", req.Capacity, err)
	}
	return nil
}

// BuildRedisFailoverTemplateData builds template data from the API request, default namespace, and optional default storage class from config.
// Optional request fields use defaults when not set. cfgDefaultStorageClass is used when req.StorageClass is empty; if empty, package default is used.
func BuildRedisFailoverTemplateData(req models.CreateRedisRequest, defaultNamespace, cfgDefaultStorageClass string) RedisFailoverTemplateData {
	storageClass := cfgDefaultStorageClass
	if storageClass == "" {
		storageClass = defaultStorageClass
	}
	if req.StorageClass != "" {
		storageClass = req.StorageClass
	}

	data := RedisFailoverTemplateData{
		Name:             req.Name,
		Namespace:        defaultNamespace,
		SentinelReplicas: defaultSentinelReplicas,
		RedisReplicas:    defaultRedisReplicas,
		CPURequest:      defaultCPURequest,
		MemoryRequest:   defaultMemoryRequest,
		CPULimit:         defaultCPULimit,
		MemoryLimit:      defaultMemoryLimit,
		StorageClass:     storageClass,
		StorageSize:      defaultStorageSize,
	}
	if defaultNamespace == "" {
		data.Namespace = "default"
	}
	if req.Capacity != "" {
		data.StorageSize = req.Capacity
	}
	if req.SentinelReplicas != nil {
		data.SentinelReplicas = *req.SentinelReplicas
	}
	if req.RedisReplicas != nil {
		data.RedisReplicas = *req.RedisReplicas
	}
	if req.CPURequest != "" {
		data.CPURequest = req.CPURequest
	}
	if req.MemoryRequest != "" {
		data.MemoryRequest = req.MemoryRequest
	}
	if req.CPULimit != "" {
		data.CPULimit = req.CPULimit
	}
	if req.MemoryLimit != "" {
		data.MemoryLimit = req.MemoryLimit
	}
	return data
}
