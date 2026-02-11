# RedisFailover CRD (install once per cluster)

This CRD is **not** synced by Argo CD because it exceeds Kubernetes' metadata annotation size limit (256KB), which causes sync to fail.

**Apply once** (e.g. after creating the cluster or before the Argo CD app syncs). Use **server-side apply** so the full spec is not stored in client-side annotations (which would exceed the 256KB limit):

```bash
kubectl apply --server-side -f crd.yaml
```

To update an existing CRD in place, add `--force-conflicts`:

```bash
kubectl apply --server-side --force-conflicts -f crd.yaml
```

The Argo CD application `redis-operator` syncs only the operator deployment and instance from `../spotahome_redis_operator/` (operator.yaml, instance.yaml). Ensure this CRD is installed before creating any `RedisFailover` resources.
