package api

import (
	"os"
)

type Config struct {
	KubeConfigPath string
	PaaSNamespace   string
	APIListenAddr     string
}

func GetConfig() *Config {
	kubeConfigPath := os.Getenv("KUBECONFIG")
	paasNamespace := os.Getenv("PAAS_NAMESPACE")
	apiListenAddr := os.Getenv("API_LISTEN_ADDR")

	if paasNamespace == "" {
		paasNamespace = "default"
	}

	if apiListenAddr == "" {
		apiListenAddr = ":8080"
	}

	return &Config{
		KubeConfigPath: kubeConfigPath,
		PaaSNamespace:   paasNamespace,
		APIListenAddr:     apiListenAddr,
	}
}
