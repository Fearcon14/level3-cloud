provider "stackit" {
	default_region         = "eu01"
	service_account_key_path = var.service_account_key_path
}

# Kubernetes provider configured via the generated kubeconfig from the SKE cluster.
provider "kubernetes" {
  config_path = local_sensitive_file.kubeconfig.filename
}

# Helm provider using the same kubeconfig (for installing Argo CD).
provider "helm" {
  kubernetes {
    config_path = local_sensitive_file.kubeconfig.filename
  }
}
