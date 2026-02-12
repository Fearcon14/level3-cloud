provider "stackit" {
  default_region           = "eu01"
  service_account_key_path = var.service_account_key_path
}

# 1. Parse the kubeconfig YAML directly from the resource output
locals {
  # Decode the YAML string into a Terraform object
  kube_config_parsed = yamldecode(stackit_ske_kubeconfig.cluster.kube_config)

  # Extract cluster connection details
  cluster_config = local.kube_config_parsed.clusters[0].cluster
  user_config    = local.kube_config_parsed.users[0].user
}

# 2. Configure Kubernetes provider using the parsed values
provider "kubernetes" {
  host                   = local.cluster_config.server
  cluster_ca_certificate = base64decode(local.cluster_config.certificate-authority-data)

  # SKE usually uses a token or client certs; we use `try` to handle whichever is present
  token                  = try(local.user_config.token, null)
  client_certificate     = try(base64decode(local.user_config.client-certificate-data), null)
  client_key             = try(base64decode(local.user_config.client-key-data), null)
}

# 3. Configure Helm provider using the same parsed values
provider "helm" {
  kubernetes {
    host                   = local.cluster_config.server
    cluster_ca_certificate = base64decode(local.cluster_config.certificate-authority-data)

    token                  = try(local.user_config.token, null)
    client_certificate     = try(base64decode(local.user_config.client-certificate-data), null)
    client_key             = try(base64decode(local.user_config.client-key-data), null)
  }
}
