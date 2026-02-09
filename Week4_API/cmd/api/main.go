package main

import (
	"log"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/api"
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/k8s"
)

func main() {
	cfg := api.GetConfig()

	clientset, err := k8s.NewClientset(cfg.KubeConfigPath)
	if err != nil {
		log.Fatalf("failed to create clientset: %v", err)
	}

	store := k8s.NewK8sInstanceStore(clientset, cfg.PaaSNamespace)
	e := api.NewServer(cfg, store)

	if err := e.Start(cfg.APIListenAddr); err != nil {
		log.Fatal(err)
	}
	log.Printf("API server started on %s", cfg.APIListenAddr)
}
