# Week 4 – PaaS REST API

RESTful API for managing PaaS product instances (e.g. Redis Failover). See repo root `README.md` for track context.

## Container image (Podman)

The API is containerized with a multi-stage Dockerfile. **Podman is Docker-compatible**, so you use the same Dockerfile with `podman` instead of `docker`.

### Build

From this directory (`Week4_API/`).

For **STACKIT SKE** (and most cloud clusters), build for **linux/amd64** (required on ARM Macs):

```bash
podman build --platform linux/amd64 -t paas-api:latest .
```

Or with a tag for STACKIT Container Registry:

```bash
podman build --platform linux/amd64 -t registry.onstackit.cloud/kevin-sinn/paas-api:latest .
```

For local run-only on ARM, you can omit `--platform linux/amd64` to use the native architecture.

### Run locally

The API needs a Kubernetes config and (optionally) env overrides:

```bash
podman run --rm -p 8080:8080 \
  -e KUBECONFIG=/kube/config \
  -v "$HOME/.kube/config:/kube/config:ro" \
  paas-api:latest
```

For a cluster that uses a service account (e.g. in SKE), you typically don’t mount `KUBECONFIG` and rely on in-cluster config when the container runs inside the cluster.

Optional env vars:

| Variable | Default | Description |
|----------|---------|-------------|
| `API_LISTEN_ADDR` | `:8080` | Listen address |
| `PAAS_NAMESPACE` | `default` | Namespace for RedisFailover resources |
| `REDIS_FAILOVER_TEMPLATE` | `internal/k8s/templates/redis-failover.yaml.tpl` | Path to template in container |
| `PAAS_DEFAULT_STORAGE_CLASS` | `premium-perf1-stackit` | Default StorageClass for PVCs |

### Push to STACKIT Container Registry

1. Log in (use a registry token or robot account):

   ```bash
   podman login registry.onstackit.cloud
   ```

2. Build for **linux/amd64** and push:

   ```bash
   export REGISTRY=registry.onstackit.cloud/kevin-sinn
   podman build --platform linux/amd64 -t $REGISTRY/paas-api:latest .
   podman push $REGISTRY/paas-api:latest
   ```

Then reference `registry.onstackit.cloud/kevin-sinn/paas-api:latest` in your Kubernetes Deployment (e.g. `Week4_API/API/deployment.yaml` or GitOps manifests).

### Accessing the API (Ingress)

The cluster uses **one** LoadBalancer for the ingress controller (ingress-nginx). The PaaS API is exposed via an Ingress (`Week4_API/API/ingress.yaml`) with host `paas-api`. After the ingress controller has an external IP:

```bash
# Get the LoadBalancer IP
kubectl get svc -n ingress-nginx ingress-nginx-controller -o jsonpath='{.status.loadBalancer.ingress[0].ip}'
# Add to /etc/hosts: <that-ip> paas-api
# Then: curl http://paas-api/instances
```

Redis instances use **in-cluster** connection info only (no per-instance LoadBalancer): `publicEndpoint` is the internal DNS (e.g. `<name>-redis.default.svc.cluster.local:6379`). Use from pods in the same cluster or via `kubectl port-forward`.

### Image pull secret (ErrImagePull on SKE)

STACKIT Container Registry is private. So the cluster needs credentials to pull the image. Create a `docker-registry` secret in the same namespace as the deployment (e.g. `default`), then the deployment’s `imagePullSecrets` will use it:

```bash
kubectl create secret docker-registry stackit-registry \
  --docker-server=registry.onstackit.cloud \
  --docker-username=<your-registry-username> \
  --docker-password=<your-registry-token-or-password> \
  -n default
```

Use the same username/token you use for `podman login registry.onstackit.cloud`. Then (re)apply the deployment:

```bash
kubectl apply -f API/deployment.yaml
```

If the image doesn’t exist yet, fix the ErrImagePull by pushing first: `podman push registry.onstackit.cloud/kevin-sinn/paas-api:latest`.

## Development

- Run tests: `go test ./...`
- API spec: `docs/openapi.yaml`
