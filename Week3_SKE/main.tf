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
