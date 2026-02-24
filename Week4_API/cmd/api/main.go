package main

import (
	"context"
	"log"

	"github.com/Fearcon14/level3-cloud/Week4_API/internal/api"
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/k8s"
	"github.com/Fearcon14/level3-cloud/Week4_API/internal/logstore"
)

func main() {
	cfg := api.GetConfig()

	dynamicClient, err := k8s.NewDynamicClient(cfg.KubeConfigPath)
	if err != nil {
		log.Fatalf("failed to create dynamic client: %v", err)
	}
	store := k8s.NewRedisFailoverStore(dynamicClient, cfg.PaaSNamespace, cfg.RedisFailoverTemplatePath, cfg.DefaultStorageClass)

	var logStore logstore.Store
	if cfg.DatabaseURL != "" {
		ps, err := logstore.NewPostgresStore(context.Background(), cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("failed to connect to log database: %v", err)
		}
		defer ps.Close()
		logStore = ps
		log.Printf("log store connected (audit and service logs enabled)")
	}

	e := api.NewServer(cfg, store, logStore)

	log.Printf("API server started on %s", cfg.APIListenAddr)
	if err := e.Start(cfg.APIListenAddr); err != nil {
		log.Fatal(err)
	}
}
