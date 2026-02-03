output "cluster_name" {
	description = "Name of the SKE Kubernetes cluster"
	value       = stackit_ske_cluster.cluster.name
}

output "kubeconfig_path" {
	description = "Path to the kubeconfig file for kubectl access"
	value       = local_sensitive_file.kubeconfig.filename
}

output "ssh_private_key_path" {
	description = "Path to the generated private key file for SSH (for node access if needed)"
	value       = local_sensitive_file.private_key.filename
}

output "key_pair_name" {
	description = "Name of the key pair registered in STACKIT"
	value       = stackit_key_pair.ske_key_pair.name
}
