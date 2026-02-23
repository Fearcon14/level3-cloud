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
3. Argo CD then syncs the `kube-prometheus-stack` Application, which installs the Helm chart into `monitoring`.

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

## Redis operator metrics

The Redis operator in `Week3_SKE/kubernetes/spotahome_redis_operator/operator.yaml` includes a ServiceMonitor and PodMonitor with label `release: kube-prometheus-stack` so Prometheus (with `serviceMonitorSelectorNilUsesHelmValues: false`) discovers and scrapes its metrics.
