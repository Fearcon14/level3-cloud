package k8s

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/models"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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

// ErrNotFound is returned when a Redis instance (RedisFailover CR) does not exist.
// Handlers should respond with HTTP 404 when errors.Is(err, ErrNotFound).
var ErrNotFound = errors.New("redis instance not found")

// InstanceStore defines instance operations; implemented by RedisFailoverStore (dynamic client).
type InstanceStore interface {
	ListInstances(ctx context.Context) ([]models.RedisInstance, error)
	GetInstance(ctx context.Context, id string) (*models.RedisInstance, error)
	CreateInstance(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error)
	UpdateInstanceCapacity(ctx context.Context, id string, req models.UpdateInstanceCapacityRequest) (*models.RedisInstance, error)
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
	client             dynamic.Interface
	namespace          string
	templatePath       string
	defaultStorageClass string
}

// NewRedisFailoverStore returns a store that lists/creates/updates/deletes RedisFailover CRs.
// defaultStorageClass is used when CreateRedisRequest.StorageClass is empty (e.g. from config).
func NewRedisFailoverStore(client dynamic.Interface, namespace, templatePath, defaultStorageClass string) *RedisFailoverStore {
	return &RedisFailoverStore{
		client:              client,
		namespace:           namespace,
		templatePath:        templatePath,
		defaultStorageClass: defaultStorageClass,
	}
}

// ListInstances returns all RedisFailover CRs in the store's namespace.
// If a CR has no status (e.g. Spotahome operator), status is inferred from that instance's pods.
func (s *RedisFailoverStore) ListInstances(ctx context.Context) ([]models.RedisInstance, error) {
	list, err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list redisfailovers: %w", err)
	}
	var instances []models.RedisInstance
	for i := range list.Items {
		inst := redisfailoverToModel(&list.Items[i])
		if inst != nil {
			if inst.Status == "unknown" {
				inst.Status = s.inferStatusFromPods(ctx, inst.Name)
			}
			instances = append(instances, *inst)
		}
	}
	return instances, nil
}

// GetInstance returns a single RedisFailover by name (id). Returns ErrNotFound if the CR does not exist.
// If the CR has no status (e.g. Spotahome operator), status is inferred from the instance's pods.
func (s *RedisFailoverStore) GetInstance(ctx context.Context, id string) (*models.RedisInstance, error) {
	obj, err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).Get(ctx, id, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
		}
		return nil, fmt.Errorf("get redisfailover %q: %w", id, err)
	}
	inst := redisfailoverToModel(obj)
	if inst != nil && inst.Status == "unknown" {
		inst.Status = s.inferStatusFromPods(ctx, id)
	}
	return inst, nil
}

// CreateInstance implements template → create: validate request, render RedisFailover from template, decode YAML, then create CR via dynamic client.
func (s *RedisFailoverStore) CreateInstance(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Capacity == "" {
		return nil, fmt.Errorf("capacity is required")
	}
	if err := ValidateCreateRedisRequest(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	data := BuildRedisFailoverTemplateData(req, s.namespace, s.defaultStorageClass)
	yamlBytes, err := RenderRedisFailoverTemplate(s.templatePath, data)
	if err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	obj, err := DecodeYAMLToUnstructured(yamlBytes)
	if err != nil {
		return nil, fmt.Errorf("decode yaml: %w", err)
	}
	obj.SetNamespace(s.namespace)
	gv := schema.GroupVersion{Group: gvrRedisFailover.Group, Version: gvrRedisFailover.Version}
	obj.SetAPIVersion(gv.String())
	obj.SetKind("RedisFailover")
	created, err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create redisfailover: %w", err)
	}
	return redisfailoverToModel(created), nil
}

// UpdateInstanceCapacity updates storage size (and optionally StorageClass) on the RedisFailover CR.
// Returns ErrNotFound if the CR does not exist.
func (s *RedisFailoverStore) UpdateInstanceCapacity(ctx context.Context, id string, req models.UpdateInstanceCapacityRequest) (*models.RedisInstance, error) {
	if err := ValidateUpdateInstanceCapacityRequest(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	obj, err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).Get(ctx, id, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
		}
		return nil, fmt.Errorf("get redisfailover %q: %w", id, err)
	}
	pvcPath := []string{"spec", "redis", "storage", "persistentVolumeClaim", "spec"}
	if err := unstructured.SetNestedField(obj.Object, req.Capacity, append(pvcPath, "resources", "requests", "storage")...); err != nil {
		return nil, fmt.Errorf("set storage: %w", err)
	}
	if req.StorageClass != "" {
		if err := unstructured.SetNestedField(obj.Object, req.StorageClass, append(pvcPath, "storageClassName")...); err != nil {
			return nil, fmt.Errorf("set storageClassName: %w", err)
		}
	}
	updated, err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("update redisfailover %q: %w", id, err)
	}
	return redisfailoverToModel(updated), nil
}

// DeleteInstance deletes the RedisFailover CR; the operator cleans up the underlying resources.
// Returns ErrNotFound if the CR does not exist.
func (s *RedisFailoverStore) DeleteInstance(ctx context.Context, id string) error {
	if err := s.client.Resource(gvrRedisFailover).Namespace(s.namespace).Delete(ctx, id, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return fmt.Errorf("%w: %s", ErrNotFound, id)
		}
		return fmt.Errorf("delete redisfailover %q: %w", id, err)
	}
	return nil
}

// gvrPods is used to infer instance status from pod phases when the CR has no status (e.g. Spotahome operator).
var gvrPods = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

// redisfailoverToModel maps a RedisFailover CR to the API RedisInstance model.
// Status is read from status.phase, status.state, or status.status; if none set, remains "unknown".
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
	for _, path := range []string{"status.phase", "status.state", "status.status"} {
		if v, ok, _ := unstructured.NestedString(obj.Object, pathToSlice(path)...); ok && v != "" {
			status = v
			break
		}
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

func pathToSlice(path string) []string {
	var out []string
	for _, s := range splitPath(path) {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func splitPath(path string) []string {
	return strings.Split(path, ".")
}

// inferStatusFromPods infers instance status from pod phases when the RedisFailover CR has no status.
// Lists pods with label app.kubernetes.io/instance=<name> (Spotahome operator convention).
func (s *RedisFailoverStore) inferStatusFromPods(ctx context.Context, name string) string {
	list, err := s.client.Resource(gvrPods).Namespace(s.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app.kubernetes.io/instance=" + name,
	})
	if err != nil || len(list.Items) == 0 {
		return "unknown"
	}
	hasPending := false
	hasRunning := false
	hasFailed := false
	for i := range list.Items {
		phase, _, _ := unstructured.NestedString(list.Items[i].Object, "status", "phase")
		switch phase {
		case "Running":
			hasRunning = true
		case "Pending":
			hasPending = true
		case "Failed":
			hasFailed = true
		}
	}
	if hasFailed {
		return "failed"
	}
	if hasPending {
		return "pending"
	}
	if hasRunning {
		return "running"
	}
	return "unknown"
}
