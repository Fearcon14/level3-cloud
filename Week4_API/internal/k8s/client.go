package k8s

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/models"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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
	PatchInstance(ctx context.Context, id string, req models.PatchInstanceRequest) (*models.RedisInstance, error)
	RegenerateInstancePassword(ctx context.Context, id string) (*models.RedisInstance, error)
	DeleteInstance(ctx context.Context, id string) error
}

// namespaceKey is used to store the target namespace in the context for multi-tenant operation.
type namespaceKey struct{}

// WithNamespace returns a derived context that carries the Kubernetes namespace to operate in.
func WithNamespace(ctx context.Context, ns string) context.Context {
	return context.WithValue(ctx, namespaceKey{}, ns)
}

// namespaceFromContext returns the namespace from context, falling back to the store's default or "default".
func namespaceFromContext(ctx context.Context, fallback string) string {
	if v, ok := ctx.Value(namespaceKey{}).(string); ok && v != "" {
		return v
	}
	if fallback != "" {
		return fallback
	}
	return "default"
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
	client              dynamic.Interface
	namespace           string
	templatePath        string
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

// EnsureNamespace creates the namespace if it does not exist (for per-user tenant namespaces).
func (s *RedisFailoverStore) EnsureNamespace(ctx context.Context, ns string) error {
	if ns == "" {
		return fmt.Errorf("namespace is required")
	}
	nsClient := s.client.Resource(gvrNamespaces)
	_, err := nsClient.Get(ctx, ns, metav1.GetOptions{})
	if err == nil {
		return nil
	}
	if !k8serrors.IsNotFound(err) {
		return fmt.Errorf("get namespace %q: %w", ns, err)
	}
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": ns,
			},
		},
	}
	_, err = nsClient.Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("create namespace %q: %w", ns, err)
	}
	return nil
}

// ListInstances returns all RedisFailover CRs in the store's namespace.
// If a CR has no status (e.g. Spotahome operator), status is inferred from that instance's pods.
func (s *RedisFailoverStore) ListInstances(ctx context.Context) ([]models.RedisInstance, error) {
	ns := namespaceFromContext(ctx, s.namespace)
	if err := s.EnsureNamespace(ctx, ns); err != nil {
		return nil, fmt.Errorf("ensure namespace: %w", err)
	}
	list, err := s.client.Resource(gvrRedisFailover).Namespace(ns).List(ctx, metav1.ListOptions{})
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
			s.attachConnectionInfo(ctx, inst)
			instances = append(instances, *inst)
		}
	}
	return instances, nil
}

// GetInstance returns a single RedisFailover by name (id). Returns ErrNotFound if the CR does not exist.
// If the CR has no status (e.g. Spotahome operator), status is inferred from the instance's pods.
func (s *RedisFailoverStore) GetInstance(ctx context.Context, id string) (*models.RedisInstance, error) {
	ns := namespaceFromContext(ctx, s.namespace)
	obj, err := s.client.Resource(gvrRedisFailover).Namespace(ns).Get(ctx, id, metav1.GetOptions{})
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
	s.attachConnectionInfo(ctx, inst)

	// Fetch password from secret
	secretName := id + "-auth"
	secretObj, err := s.client.Resource(gvrSecrets).Namespace(ns).Get(ctx, secretName, metav1.GetOptions{})
	if err == nil {
		if data, found, _ := unstructured.NestedStringMap(secretObj.Object, "data"); found {
			if encoded, ok := data["password"]; ok {
				decoded, _ := base64.StdEncoding.DecodeString(encoded)
				inst.Password = string(decoded)
			}
		} else if stringData, found, _ := unstructured.NestedStringMap(secretObj.Object, "stringData"); found {
			if pwd, ok := stringData["password"]; ok {
				inst.Password = pwd
			}
		}
	}

	return inst, nil
}

// generatePassword generates a random hex string for the password.
func generatePassword() (string, error) {
	b := make([]byte, passwordLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// createSecret creates a Secret with the given password in stringData.
func (s *RedisFailoverStore) createSecret(ctx context.Context, ns, name, password string) error {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": ns,
			},
			"stringData": map[string]interface{}{
				"password": password,
			},
			"type": "Opaque",
		},
	}
	_, err := s.client.Resource(gvrSecrets).Namespace(ns).Create(ctx, obj, metav1.CreateOptions{})
	return err
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
	ns := namespaceFromContext(ctx, s.namespace)
	if err := s.EnsureNamespace(ctx, ns); err != nil {
		return nil, fmt.Errorf("ensure namespace: %w", err)
	}

	// Generate and create password secret
	password, err := generatePassword()
	if err != nil {
		return nil, fmt.Errorf("generate password: %w", err)
	}
	secretName := req.Name + "-auth"
	if err := s.createSecret(ctx, ns, secretName, password); err != nil {
		return nil, fmt.Errorf("create secret: %w", err)
	}

	data := BuildRedisFailoverTemplateData(req, ns, s.defaultStorageClass, secretName)
	yamlBytes, err := RenderRedisFailoverTemplate(s.templatePath, data)
	if err != nil {
		return nil, fmt.Errorf("render template: %w", err)
	}
	obj, err := DecodeYAMLToUnstructured(yamlBytes)
	if err != nil {
		return nil, fmt.Errorf("decode yaml: %w", err)
	}
	obj.SetNamespace(ns)
	gv := schema.GroupVersion{Group: gvrRedisFailover.Group, Version: gvrRedisFailover.Version}
	obj.SetAPIVersion(gv.String())
	obj.SetKind("RedisFailover")
	created, err := s.client.Resource(gvrRedisFailover).Namespace(ns).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create redisfailover: %w", err)
	}

	inst := redisfailoverToModel(created)
	s.attachConnectionInfo(ctx, inst)
	inst.Password = password
	return inst, nil
}

// jsonPatchOp represents one RFC 6902 JSON Patch operation.
type jsonPatchOp struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// PatchInstance performs a partial update on the RedisFailover CR using the Kubernetes API server's
// JSON Patch (RFC 6902). Only the requested fields are sent; the server applies the patch on the
// current resource version, avoiding read-modify-write races and accidental overwrite of other fields.
// It can update the display name (annotation), replicas, and capacity (PVC size).
// Returns ErrNotFound if the CR does not exist.
func (s *RedisFailoverStore) PatchInstance(ctx context.Context, id string, req models.PatchInstanceRequest) (*models.RedisInstance, error) {
	if err := ValidatePatchInstanceRequest(req); err != nil {
		return nil, fmt.Errorf("validation: %w", err)
	}
	ns := namespaceFromContext(ctx, s.namespace)

	var ops []jsonPatchOp

	// Display name: patch annotation. We may need current annotations so we don't wipe them when the map is missing.
	const displayNameKey = "app.kubernetes.io/display-name"
	if req.Name != nil {
		var existingAnnotations map[string]string
		existing, err := s.client.Resource(gvrRedisFailover).Namespace(ns).Get(ctx, id, metav1.GetOptions{})
		if err != nil {
			if k8serrors.IsNotFound(err) {
				return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
			}
			return nil, fmt.Errorf("get redisfailover for annotation patch: %w", err)
		}
		existingAnnotations, _, _ = unstructured.NestedStringMap(existing.Object, "metadata", "annotations")
		merged := map[string]string{}
		for k, v := range existingAnnotations {
			merged[k] = v
		}
		merged[displayNameKey] = *req.Name
		if existingAnnotations == nil {
			ops = append(ops, jsonPatchOp{Op: "add", Path: "/metadata/annotations", Value: merged})
		} else {
			// JSON Pointer: / in key must be encoded as ~1
			ops = append(ops, jsonPatchOp{Op: "add", Path: "/metadata/annotations/app.kubernetes.io~1display-name", Value: *req.Name})
		}
	}

	if req.Capacity != nil {
		ops = append(ops, jsonPatchOp{
			Op: "replace", Path: "/spec/redis/storage/persistentVolumeClaim/spec/resources/requests/storage", Value: *req.Capacity,
		})
	}
	if req.RedisReplicas != nil {
		ops = append(ops, jsonPatchOp{Op: "replace", Path: "/spec/redis/replicas", Value: int64(*req.RedisReplicas)})
	}
	if req.SentinelReplicas != nil {
		ops = append(ops, jsonPatchOp{Op: "replace", Path: "/spec/sentinel/replicas", Value: int64(*req.SentinelReplicas)})
	}

	if len(ops) == 0 {
		return nil, fmt.Errorf("validation: at least one field must be provided")
	}

	patchBytes, err := json.Marshal(ops)
	if err != nil {
		return nil, fmt.Errorf("build json patch: %w", err)
	}

	updated, err := s.client.Resource(gvrRedisFailover).Namespace(ns).Patch(ctx, id, types.JSONPatchType, patchBytes, metav1.PatchOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
		}
		return nil, fmt.Errorf("patch redisfailover %q: %w", id, err)
	}
	inst := redisfailoverToModel(updated)
	s.attachConnectionInfo(ctx, inst)
	return inst, nil
}

// RegenerateInstancePassword generates a new password and updates the instance's auth Secret.
// Returns the instance with the new password populated. Returns ErrNotFound if the instance or secret does not exist.
func (s *RedisFailoverStore) RegenerateInstancePassword(ctx context.Context, id string) (*models.RedisInstance, error) {
	ns := namespaceFromContext(ctx, s.namespace)
	secretName := id + "-auth"
	obj, err := s.client.Resource(gvrSecrets).Namespace(ns).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, id)
		}
		return nil, fmt.Errorf("get secret %q: %w", secretName, err)
	}
	newPassword, err := generatePassword()
	if err != nil {
		return nil, fmt.Errorf("generate password: %w", err)
	}
	if err := unstructured.SetNestedStringMap(obj.Object, map[string]string{"password": newPassword}, "stringData"); err != nil {
		return nil, fmt.Errorf("set secret stringData: %w", err)
	}
	if _, err := s.client.Resource(gvrSecrets).Namespace(ns).Update(ctx, obj, metav1.UpdateOptions{}); err != nil {
		return nil, fmt.Errorf("update secret %q: %w", secretName, err)
	}
	// Give the secret update time to propagate to the kubelet cache so new pods mount the new password.
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(5 * time.Second):
	}
	// Restart Redis and Sentinel workloads so they pick up the new password from the secret.
	s.rolloutRestartRedisWorkloads(ctx, ns, id)
	inst, err := s.GetInstance(ctx, id)
	if err != nil {
		return nil, err
	}
	inst.Password = newPassword
	return inst, nil
}

// rolloutRestartRedisWorkloads restarts Redis and Sentinel so they pick up the new password.
// - Redis: we delete the Redis pods (rfr-<id>-*) so the StatefulSet recreates them; the operator
//   often reconciles the StatefulSet spec and would revert an annotation-based rollout.
// - Sentinel: we patch the Deployment with restartedAt so it rolls out normally.
func (s *RedisFailoverStore) rolloutRestartRedisWorkloads(ctx context.Context, ns, id string) {
	// Restart Redis by deleting its pods (StatefulSet will recreate them with the new secret).
	s.deleteRedisPodsForRestart(ctx, ns, id)
	// Restart Sentinel by patching the Deployment template (rollout restart).
	s.rolloutRestartDeployment(ctx, ns, "rfs-"+id)
}

// deleteRedisPodsForRestart deletes pods that belong to the Redis StatefulSet (rfr-<id>-*) so they are recreated with the updated secret.
func (s *RedisFailoverStore) deleteRedisPodsForRestart(ctx context.Context, ns, id string) {
	list, err := s.client.Resource(gvrPods).Namespace(ns).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("redisfailovers.databases.spotahome.com/name=%s", id),
	})
	if err != nil {
		return
	}
	prefix := "rfr-" + id + "-"
	for i := range list.Items {
		name, _, _ := unstructured.NestedString(list.Items[i].Object, "metadata", "name")
		if name == "" || !strings.HasPrefix(name, prefix) {
			continue
		}
		_ = s.client.Resource(gvrPods).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
	}
}

// rolloutRestartDeployment patches a Deployment's pod template with restartedAt to trigger a rollout.
func (s *RedisFailoverStore) rolloutRestartDeployment(ctx context.Context, ns, name string) {
	obj, err := s.client.Resource(gvrDeployments).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return
	}
	annotations, _, _ := unstructured.NestedStringMap(obj.Object, "spec", "template", "metadata", "annotations")
	if annotations == nil {
		annotations = map[string]string{}
	}
	annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().UTC().Format(time.RFC3339)
	_ = unstructured.SetNestedStringMap(obj.Object, annotations, "spec", "template", "metadata", "annotations")
	_, _ = s.client.Resource(gvrDeployments).Namespace(ns).Update(ctx, obj, metav1.UpdateOptions{})
}

// DeleteInstance deletes the RedisFailover CR; the operator cleans up the underlying resources.
// Returns ErrNotFound if the CR does not exist.
func (s *RedisFailoverStore) DeleteInstance(ctx context.Context, id string) error {
	ns := namespaceFromContext(ctx, s.namespace)
	if err := s.client.Resource(gvrRedisFailover).Namespace(ns).Delete(ctx, id, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			return fmt.Errorf("%w: %s", ErrNotFound, id)
		}
		return fmt.Errorf("delete redisfailover %q: %w", id, err)
	}

	_ = s.client.Resource(gvrSecrets).Namespace(ns).Delete(ctx, id+"-auth", metav1.DeleteOptions{})
	return nil
}

// gvrPods is used to infer instance status from pod phases when the CR has no status (e.g. Spotahome operator).
var gvrPods = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}

// gvrNamespaces is used to ensure a tenant namespace exists before use.
var gvrNamespaces = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}

// gvrSecrets is used to create and delete the secret containing the Redis password.
var gvrSecrets = schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}

// gvrDeployments is used to trigger Sentinel rollout restart after password regeneration (rfs-<name>). Redis is restarted by deleting its pods (rfr-<name>-*).
var gvrDeployments = schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}

const (
	passwordLength = 16
)

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
	// Use display-name annotation as the human-friendly name when present.
	displayName := name
	if dn, ok, _ := unstructured.NestedString(obj.Object, "metadata", "annotations", "app.kubernetes.io/display-name"); ok && dn != "" {
		displayName = dn
	}

	return &models.RedisInstance{
		ID:               name,
		Name:             displayName,
		Namespace:        ns,
		Status:           status,
		Capacity:         capacity,
		RedisReplicas:    int(redisReplicas),
		SentinelReplicas: int(sentinelReplicas),
	}
}

// attachConnectionInfo enriches a RedisInstance with in-cluster connection data.
// The Spotahome Redis operator creates a ClusterIP Service "rfrm-<name>-redis" for the Redis master (port 6379).
func (s *RedisFailoverStore) attachConnectionInfo(ctx context.Context, inst *models.RedisInstance) {
	if inst == nil {
		return
	}
	ns := inst.Namespace
	if ns == "" {
		ns = "default"
	}
	svcName := "rfrm-" + inst.Name
	inst.PublicServiceName = svcName
	inst.PublicPort = 6379
	inst.PublicHostname = fmt.Sprintf("%s.%s.svc.cluster.local", svcName, ns)
	inst.PublicEndpoint = fmt.Sprintf("%s:%d", inst.PublicHostname, inst.PublicPort)
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
	ns := namespaceFromContext(ctx, s.namespace)

	list, err := s.client.Resource(gvrPods).Namespace(ns).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("redisfailovers.databases.spotahome.com/name=%s", name),
	})

	// If no pods found
	if err != nil || len(list.Items) == 0 {
		list, err = s.client.Resource(gvrPods).Namespace(ns).List(ctx, metav1.ListOptions{
			LabelSelector: "app.kubernetes.io/instance=" + name,
		})
	}

	if err != nil || len(list.Items) == 0 {
		return "unknown"
	}
	hasPending := false
	hasRunning := false
	hasFailed := false
	for i := range list.Items {
		// status.phase is top-level for pods
		phase, _, _ := unstructured.NestedString(list.Items[i].Object, "status", "phase")

		// If phase is empty, check container statuses
		if phase == "" {
			containerStatuses, found, _ := unstructured.NestedSlice(list.Items[i].Object, "status", "containerStatuses")
			if found && len(containerStatuses) > 0 {
				// Assume running if at least one container is running
				for _, cs := range containerStatuses {
					if statusMap, ok := cs.(map[string]interface{}); ok {
						if _, running := statusMap["state"].(map[string]interface{})["running"]; running {
							phase = "Running"
							break
						}
					}
				}
			}
		}

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
