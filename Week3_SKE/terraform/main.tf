# SKE (STACKIT Kubernetes Engine) cluster
resource "stackit_ske_cluster" "cluster" {
	project_id = var.project_id
	name       = var.cluster_name

	node_pools = [
		{
			name               = "np-1"
			machine_type       = var.machine_type
			os_name            = "ubuntu"
			os_version_min     = "2204.20250728.0"
			minimum            = 1
			maximum            = 2
			availability_zones = ["eu01-1"]
		}
	]
}

# Kubeconfig for kubectl access (created after cluster is ready).
# Set refresh = true so Terraform gets a new kubeconfig when it expires (default 1h).
resource "stackit_ske_kubeconfig" "cluster" {
	project_id   = var.project_id
	cluster_name = stackit_ske_cluster.cluster.name
	refresh      = true
}

# Write kubeconfig to a file so you can use kubectl
resource "local_sensitive_file" "kubeconfig" {
	content         = stackit_ske_kubeconfig.cluster.kube_config
	filename        = "${path.module}/kubeconfig-${stackit_ske_cluster.cluster.name}"
	file_permission = "0600"
}

# Install Argo CD into the SKE cluster using Helm.
resource "helm_release" "argocd" {
  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  namespace        = "argocd"
  create_namespace = true

  depends_on = [
    local_sensitive_file.kubeconfig,
  ]
}

# Argo CD Application that points to this repo's Redis operator manifests.
resource "kubernetes_manifest" "argocd_redis_operator_app" {
  depends_on = [
    helm_release.argocd,
  ]

  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "redis-operator"
      namespace = "argocd"
    }
    spec = {
      project = "default"
      source = {
        repoURL        = "https://github.com/Fearcon14/level3-cloud.git"
        targetRevision = "main"
        path           = "Week3_SKE/kubernetes/spotahome_redis_operator"
      }
      destination = {
        server    = "https://kubernetes.default.svc"
        namespace = "default"
      }
      syncPolicy = {
        automated = {
          prune    = true
          selfHeal = true
        }
      }
    }
  }
}
