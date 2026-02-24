# PostgreSQL for PaaS User-Centric Logs

PostgreSQL runs in the `paas-logs` namespace and stores **audit logs** (user actions) and **service logs** (async status/compliance) for the PaaS API.

## Deploy

From the repo root, apply the Week6 kustomization (including postgres):

```bash
kubectl apply -k Week6_Monitoring/k8s/
```

Or apply only the postgres resources:

```bash
kubectl apply -k Week6_Monitoring/k8s/ --selector  # not selective; use:
kubectl apply -f Week6_Monitoring/k8s/postgres/
```

If using Argo CD with the same path (`Week6_Monitoring/k8s`), the postgres resources are included in the sync.

## Connection (API)

The API (in `default` namespace) connects with:

- **Host:** `postgres.paas-logs.svc.cluster.local`
- **Port:** `5432`
- **Database:** `paaslogs`
- **User:** `paaslogs`
- **Password:** from Secret `postgres-credentials` key `POSTGRES_PASSWORD`

URL form: `postgres://paaslogs:<password>@postgres.paas-logs.svc.cluster.local:5432/paaslogs`

## Schema

- `audit_logs`: user actions (create/update/delete instance, cache get/set). Columns: id, tenant_user, instance_id, action, details (JSONB), created_at.
- `service_logs`: async events (e.g. status changes). Columns: id, tenant_user, instance_id, event_type, message, metadata (JSONB), created_at.

The Job `postgres-schema-init` runs the schema (idempotent); ensure Postgres is ready before the Job runs (Kustomize apply order is namespace → secret → PVC → deployment → service → configmap → job).

## Production

1. Override the default password: replace the Secret or use Sealed Secrets / External Secrets.
2. Optionally set `storageClassName` in `pvc.yaml` for your cluster (e.g. SKE: `premium-perf1-stackit`).
