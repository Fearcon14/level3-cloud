package k8s

import (
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/models"
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
)

// BuildRedisFailoverTemplateData builds template data from the API request and default namespace.
// Optional fields use defaults when not set (pointers nil or strings empty).
func BuildRedisFailoverTemplateData(req models.CreateRedisRequest, defaultNamespace string) RedisFailoverTemplateData {
	data := RedisFailoverTemplateData{
		Name:             req.Name,
		Namespace:        defaultNamespace,
		SentinelReplicas:  defaultSentinelReplicas,
		RedisReplicas:     defaultRedisReplicas,
		CPURequest:       defaultCPURequest,
		MemoryRequest:    defaultMemoryRequest,
		CPULimit:         defaultCPULimit,
		MemoryLimit:      defaultMemoryLimit,
		StorageClass:     defaultStorageClass,
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
	if req.StorageClass != "" {
		data.StorageClass = req.StorageClass
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
