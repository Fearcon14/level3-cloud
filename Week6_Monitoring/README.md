# Week 6 – Internal Monitoring (Prometheus + Grafana)

This directory contains the Argo CD–driven setup for internal monitoring on the SKE cluster using the [kube-prometheus-stack](https://github.com/prometheus-community/helm-charts/tree/main/charts/kube-prometheus-stack) Helm chart (Prometheus, Grafana, node-exporter, kube-state-metrics).

## Layout

- **k8s/** – Manifests synced by the Terraform-managed Argo CD Application (path `Week6_Monitoring/k8s`):
  - `namespace.yaml` – `monitoring` namespace
  - `grafana-admin-secret.yaml` – Grafana admin credentials (use a strong password in production)
  - `application.yaml` – Argo CD Application that deploys kube-prometheus-stack from the Prometheus Community Helm repo
  - `values.yaml` – Reference values for the stack (keep in sync with `application.yaml` helm values)

## Bootstrap

1. Apply Terraform (or ensure the `monitoring` Argo CD Application exists and points at `Week6_Monitoring/k8s`).
2. Argo CD syncs `Week6_Monitoring/k8s`, creating the `monitoring` namespace, the Grafana secret, and the `kube-prometheus-stack` Application.
3. **One-time: install Prometheus Operator CRDs** (avoids "metadata.annotations: Too long" when the chart syncs). Use **server-side apply** so kubectl does not add the large `last-applied-configuration` annotation:
   ```bash
   helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
   helm show crds prometheus-community/kube-prometheus-stack --version 67.0.0 | kubectl apply -f - --server-side --force-conflicts
   ```
4. Argo CD then syncs the `kube-prometheus-stack` Application, which installs the Helm chart into `monitoring`. The Application uses `skipCrds: true` and `ServerSideApply=true` so CRDs (and other large resources) do not hit the 262144-byte annotation limit.

## Grafana

- **Admin secret**: The chart uses existing secret `grafana-admin` in namespace `monitoring` with keys `admin-user` and `admin-password`. The default in `grafana-admin-secret.yaml` is a placeholder; for production, create a secret with a strong password, e.g.:
  ```bash
  kubectl create secret generic grafana-admin -n monitoring \
    --from-literal=admin-user=admin \
    --from-literal=admin-password='YOUR_STRONG_PASSWORD' \
    --dry-run=client -o yaml | kubectl apply -f -
  ```
- **Access**: Use `kubectl port-forward svc/kube-prometheus-stack-grafana -n monitoring 3000:80` and open http://localhost:3000.

## Grafana Ingress (optional)

To expose Grafana via Ingress with TLS (e.g. cert-manager):

1. In `k8s/values.yaml` (and in `application.yaml` helm values), enable and set:
   - `grafana.ingress.enabled: true`
   - `grafana.ingress.ingressClassName: nginx`
   - `grafana.ingress.hosts` and `grafana.ingress.tls` for your domain (e.g. `grafana.<your-stackit-subdomain>`).
2. Add annotations for cert-manager if needed, e.g. `cert-manager.io/cluster-issuer: letsencrypt-prod`.
3. Re-sync the `kube-prometheus-stack` Application.

## Admission webhook (TLS "bad certificate")

If the Prometheus Operator logs show `tls: bad certificate` and Prometheus/Alertmanager never get StatefulSets, the admission webhook's TLS is failing. The stack is configured with `prometheusOperator.admissionWebhooks.failurePolicy: Ignore` so the API server still admits requests when the webhook fails; after syncing, the operator should reconcile and create Prometheus and Alertmanager pods.

## Redis operator metrics

The Redis operator in `Week3_SKE/kubernetes/spotahome_redis_operator/operator.yaml` includes a ServiceMonitor and PodMonitor with label `release: kube-prometheus-stack` so Prometheus (with `serviceMonitorSelectorNilUsesHelmValues: false`) discovers and scrapes its metrics.
