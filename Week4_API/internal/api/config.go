package api

import (
	"os"
)

type Config struct {
	KubeConfigPath           string
	PaaSNamespace            string
	APIListenAddr            string
	RedisFailoverTemplatePath string
}

func GetConfig() *Config {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	paasNamespace := os.Getenv("PAAS_NAMESPACE")
	apiListenAddr := os.Getenv("API_LISTEN_ADDR")
	templatePath := os.Getenv("REDIS_FAILOVER_TEMPLATE")

	if paasNamespace == "" {
		paasNamespace = "default"
	}
	if apiListenAddr == "" {
		apiListenAddr = ":8080"
	}
	if templatePath == "" {
		templatePath = "internal/k8s/templates/redis-failover.yaml.tpl"
	}

	return &Config{
		KubeConfigPath:           kubeConfigPath,
		PaaSNamespace:            paasNamespace,
		APIListenAddr:            apiListenAddr,
		RedisFailoverTemplatePath: templatePath,
	}
}
