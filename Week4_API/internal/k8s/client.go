package k8s

import (
	"context"
	"fmt"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/models"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	labelAppKey   = "app"
	labelAppValue = "redis-paas"

	annotationInstanceID = "paas.example.com/instance-id"
	annotationCapacity   = "paas.example.com/capacity"

	redisContainerName = "redis"
	redisImage         = "redis:7"
	redisPort          = 6379
)

type InstanceStore interface {
	ListInstances(ctx context.Context) ([]models.RedisInstance, error)
	GetInstance(ctx context.Context, id string) (*models.RedisInstance, error)
	CreateInstance(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error)
	UpdateInstanceCapacity(ctx context.Context, id string, capacity string) (*models.RedisInstance, error)
	DeleteInstance(ctx context.Context, id string) error
}

type K8sInstanceStore struct {
	clientset *kubernetes.Clientset
	namespace string
}

// NewClientset builds a Kubernetes clientset using, in order:
// 1. kubeConfigPath if non-empty (e.g. from KUBECONFIG),
// 2. in-cluster config if the process is running inside a cluster,
// 3. default kubeconfig loading rules (e.g. ~/.kube/config).
func NewClientset(kubeConfigPath string) (*kubernetes.Clientset, error) {
	config, err := buildRestConfig(kubeConfigPath)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
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

func NewK8sInstanceStore(clientset *kubernetes.Clientset, namespace string) *K8sInstanceStore {
	return &K8sInstanceStore{
		clientset: clientset,
		namespace: namespace,
	}
}

// deploymentToModel converts a Deployment into the API-facing RedisInstance model.
// Returns nil if d is nil (callers can skip nil without treating it as an error).
func deploymentToModel(d *appsv1.Deployment) *models.RedisInstance {
	if d == nil {
		return nil
	}

	ns := d.Namespace
	if ns == "" {
		ns = "default"
	}

	capacity := ""
	if d.Annotations != nil {
		capacity = d.Annotations[annotationCapacity]
	}

	status := "unknown"
	if d.Status.ReadyReplicas >= 1 {
		status = "running"
	} else if d.Status.Replicas > 0 {
		status = "starting"
	} else {
		status = "stopped"
	}

	id := d.Name
	if d.Annotations != nil {
		if v := d.Annotations[annotationInstanceID]; v != "" {
			id = v
		}
	}

	return &models.RedisInstance{
		ID:        id,
		Name:      d.Name,
		Namespace: ns,
		Status:    status,
		Capacity:  capacity,
	}
}

// buildDeployment constructs a Deployment that runs a single Redis container.
// Labels and annotations match deploymentToModel so list/get work correctly.
func buildDeployment(name, capacity string) *appsv1.Deployment {
	labelMap := map[string]string{
		labelAppKey: labelAppValue,
	}
	annotations := map[string]string{
		annotationInstanceID: name,
		annotationCapacity:   capacity,
	}
	replicas := int32(1)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labelMap,
			Annotations: annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labelMap,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labelMap,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  redisContainerName,
							Image: redisImage,
							Ports: []corev1.ContainerPort{
								{
									Name:          "redis",
									ContainerPort: redisPort,
								},
							},
						},
					},
				},
			},
		},
	}
}

// buildService constructs a ClusterIP Service that selects the Deployment's Pods
// by the same labels, exposing the Redis port.
func buildService(name string) *corev1.Service {
	labelMap := map[string]string{
		labelAppKey: labelAppValue,
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labelMap,
		},
		Spec: corev1.ServiceSpec{
			Selector: labelMap,
			Ports: []corev1.ServicePort{
				{
					Name:       "redis",
					Port:       redisPort,
					TargetPort: intstr.FromInt(redisPort),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
}

// ListInstances returns all Redis instances (Deployments with our label) in the store's namespace.
func (s *K8sInstanceStore) ListInstances(ctx context.Context) ([]models.RedisInstance, error) {
	deploymentsClient := s.clientset.AppsV1().Deployments(s.namespace)
	selector := labels.Set{labelAppKey: labelAppValue}.AsSelector().String()

	list, err := deploymentsClient.List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, fmt.Errorf("list deployments: %w", err)
	}

	var instances []models.RedisInstance
	for i := range list.Items {
		if inst := deploymentToModel(&list.Items[i]); inst != nil {
			instances = append(instances, *inst)
		}
	}
	return instances, nil
}

// GetInstance returns a single Redis instance by name (id). Returns an error if not found.
func (s *K8sInstanceStore) GetInstance(ctx context.Context, id string) (*models.RedisInstance, error) {
	deploy, err := s.clientset.AppsV1().Deployments(s.namespace).Get(ctx, id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get deployment %q: %w", id, err)
	}
	return deploymentToModel(deploy), nil
}

// CreateInstance creates a new Redis instance (Deployment + Service) from the request.
func (s *K8sInstanceStore) CreateInstance(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	deploymentsClient := s.clientset.AppsV1().Deployments(s.namespace)
	servicesClient := s.clientset.CoreV1().Services(s.namespace)

	deploy := buildDeployment(req.Name, req.Capacity)
	created, err := deploymentsClient.Create(ctx, deploy, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create deployment: %w", err)
	}
	svc := buildService(req.Name)
	if _, err := servicesClient.Create(ctx, svc, metav1.CreateOptions{}); err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}
	return deploymentToModel(created), nil
}

// UpdateInstanceCapacity updates the capacity annotation on the Deployment.
func (s *K8sInstanceStore) UpdateInstanceCapacity(ctx context.Context, id string, capacity string) (*models.RedisInstance, error) {
	deploymentsClient := s.clientset.AppsV1().Deployments(s.namespace)
	deploy, err := deploymentsClient.Get(ctx, id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get deployment %q: %w", id, err)
	}
	if deploy.Annotations == nil {
		deploy.Annotations = make(map[string]string)
	}
	deploy.Annotations[annotationCapacity] = capacity
	updated, err := deploymentsClient.Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("update deployment %q: %w", id, err)
	}
	return deploymentToModel(updated), nil
}

// DeleteInstance deletes the Deployment and its Service.
func (s *K8sInstanceStore) DeleteInstance(ctx context.Context, id string) error {
	deploymentsClient := s.clientset.AppsV1().Deployments(s.namespace)
	servicesClient := s.clientset.CoreV1().Services(s.namespace)
	propagation := metav1.DeletePropagationForeground

	if err := deploymentsClient.Delete(ctx, id, metav1.DeleteOptions{PropagationPolicy: &propagation}); err != nil {
		return fmt.Errorf("delete deployment %q: %w", id, err)
	}
	_ = servicesClient.Delete(ctx, id, metav1.DeleteOptions{})
	return nil
}
