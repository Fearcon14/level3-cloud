package k8s

import (
	"context"
	"fmt"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/models"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// RedisFailover CRD (Spotahome Redis operator): GroupVersionResource for the dynamic client.
var gvrRedisFailover = schema.GroupVersionResource{
	Group:    "databases.spotahome.com",
	Version:  "v1",
	Resource: "redisfailovers",
}

// InstanceStore defines instance operations; implemented by RedisFailoverStore (dynamic client).
type InstanceStore interface {
	ListInstances(ctx context.Context) ([]models.RedisInstance, error)
	GetInstance(ctx context.Context, id string) (*models.RedisInstance, error)
	CreateInstance(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error)
	UpdateInstanceCapacity(ctx context.Context, id string, capacity string) (*models.RedisInstance, error)
	DeleteInstance(ctx context.Context, id string) error
}

// buildRestConfig returns *rest.Config for the given kubeconfig path or fallbacks.
func buildRestConfig(kubeConfigPath string) (*rest.Config, error) {
	if kubeConfigPath != "" {
		config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		if err != nil {
			return nil, fmt.Errorf("kubeconfig %q: %w", kubeConfigPath, err)
		}
		return config, nil
	}
	if config, err := rest.InClusterConfig(); err == nil {
		return config, nil
	}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("default kubeconfig: %w", err)
	}
	return config, nil
}

// NewDynamicClient builds a dynamic Kubernetes client from kubeconfig path (for CRDs like RedisFailover).
func NewDynamicClient(kubeConfigPath string) (dynamic.Interface, error) {
	config, err := buildRestConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(config)
}

// RedisFailoverStore implements InstanceStore using RedisFailover CRs (Spotahome Redis operator).
// All instance operations go through the dynamic client; the operator handles the actual Redis/Sentinel workloads.
type RedisFailoverStore struct {
	client       dynamic.Interface
	namespace    string
	templatePath string
}

// NewRedisFailoverStore returns a store that lists/creates/updates/deletes RedisFailover CRs.
func NewRedisFailoverStore(client dynamic.Interface, namespace, templatePath string) *RedisFailoverStore {
	return &RedisFailoverStore{
		client:       client,
		namespace:    namespace,
		templatePath: templatePath,
	}
}

// ListInstances returns all RedisFailover CRs in the store's namespace.
func (s *RedisFailoverStore) ListInstances(ctx context.Context) ([]models.RedisInstance, error) {
	list, err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list redisfailovers: %w", err)
	}
	var instances []models.RedisInstance
	for i := range list.Items {
		if inst := redisfailoverToModel(&list.Items[i]); inst != nil {
			instances = append(instances, *inst)
		}
	}
	return instances, nil
}

// GetInstance returns a single RedisFailover by name (id).
func (s *RedisFailoverStore) GetInstance(ctx context.Context, id string) (*models.RedisInstance, error) {
	obj, err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).Get(ctx, id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get redisfailover %q: %w", id, err)
	}
	return redisfailoverToModel(obj), nil
}

// CreateInstance creates a new RedisFailover CR from the request by rendering the template and applying it.
func (s *RedisFailoverStore) CreateInstance(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	data := BuildRedisFailoverTemplateData(req, s.namespace)
	yamlBytes, err := RenderRedisFailoverTemplate(s.templatePath, data)
	if err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	obj, err := DecodeYAMLToUnstructured(yamlBytes)
	if err != nil {
		return nil, fmt.Errorf("decode yaml: %w", err)
	}
	obj.SetNamespace(s.namespace)
	created, err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create redisfailover: %w", err)
	}
	return redisfailoverToModel(created), nil
}

// UpdateInstanceCapacity updates the storage size on the RedisFailover CR.
func (s *RedisFailoverStore) UpdateInstanceCapacity(ctx context.Context, id string, capacity string) (*models.RedisInstance, error) {
	obj, err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).Get(ctx, id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get redisfailover %q: %w", id, err)
	}
	if err := unstructured.SetNestedField(obj.Object, capacity, "spec", "redis", "storage", "persistentVolumeClaim", "spec", "resources", "requests", "storage"); err != nil {
		return nil, fmt.Errorf("set storage: %w", err)
	}
	updated, err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("update redisfailover %q: %w", id, err)
	}
	return redisfailoverToModel(updated), nil
}

// DeleteInstance deletes the RedisFailover CR; the operator cleans up the underlying resources.
func (s *RedisFailoverStore) DeleteInstance(ctx context.Context, id string) error {
	if err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).Delete(ctx, id, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("delete redisfailover %q: %w", id, err)
	}
	return nil
}

// redisfailoverToModel maps a RedisFailover CR to the API RedisInstance model.
func redisfailoverToModel(obj *unstructured.Unstructured) *models.RedisInstance {
	if obj == nil {
		return nil
	}
	name, _, _ := unstructured.NestedString(obj.Object, "metadata", "name")
	ns, _, _ := unstructured.NestedString(obj.Object, "metadata", "namespace")
	if ns == "" {
		ns = "default"
	}
	status := "unknown"
	if phase, ok, _ := unstructured.NestedString(obj.Object, "status", "phase"); ok && phase != "" {
		status = phase
	}
	capacity, _, _ := unstructured.NestedString(obj.Object, "spec", "redis", "storage", "persistentVolumeClaim", "spec", "resources", "requests", "storage")
	redisReplicas, _, _ := unstructured.NestedInt64(obj.Object, "spec", "redis", "replicas")
	sentinelReplicas, _, _ := unstructured.NestedInt64(obj.Object, "spec", "sentinel", "replicas")
	return &models.RedisInstance{
		ID:               name,
		Name:             name,
		Namespace:        ns,
		Status:           status,
		Capacity:         capacity,
		RedisReplicas:    int(redisReplicas),
		SentinelReplicas: int(sentinelReplicas),
	}
}
