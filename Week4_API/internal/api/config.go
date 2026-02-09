package api

import (
	"os"
)

// Config holds API and RedisFailover backend settings. Only RedisFailover is used as the instance backend.
type Config struct {
	KubeConfigPath            string
	PaaSNamespace             string
	APIListenAddr             string
	RedisFailoverTemplatePath string
	DefaultStorageClass       string
}

// GetConfig loads config from environment with centralized defaults.
func GetConfig() *Config {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	paasNamespace := os.Getenv("PAAS_NAMESPACE")
	apiListenAddr := os.Getenv("API_LISTEN_ADDR")
	templatePath := os.Getenv("REDIS_FAILOVER_TEMPLATE")
	defaultStorageClass := os.Getenv("PAAS_DEFAULT_STORAGE_CLASS")

	if paasNamespace == "" {
		paasNamespace = "default"
	}
	if apiListenAddr == "" {
		apiListenAddr = ":8080"
	}
	if templatePath == "" {
		templatePath = "internal/k8s/templates/redis-failover.yaml.tpl"
	}
	if defaultStorageClass == "" {
		defaultStorageClass = "premium-perf1-stackit"
	}

	return &Config{
		KubeConfigPath:           kubeConfigPath,
		PaaSNamespace:            paasNamespace,
		APIListenAddr:            apiListenAddr,
		RedisFailoverTemplatePath: templatePath,
		DefaultStorageClass:      defaultStorageClass,
	}
}
