package k8s

import (
	"context"
	"fmt"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/models"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type InstanceStore interface {
	ListInstances(ctx context.Context) ([]models.RedisInstance, error)
	GetInstance(ctx context.Context, id string) (*models.RedisInstance, error)
	CreateInstance(ctx context.Context, req models.CreateRedisRequest) (*models.RedisInstance, error)
	UpdateInstanceCapacity(ctx context.Context, id string, capacity string) (*models.RedisInstance, error)
	DeleteInstance(ctx context.Context, id string) error
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
