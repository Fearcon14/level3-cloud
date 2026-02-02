variable "service_account_key_path" {
	description = "STACKIT service account key (Key Flow auth)"
	type        = string
	sensitive   = true
	default     = "service_account_key.json"
}

variable "project_id" {
	description = "STACKIT project ID"
	type        = string
	default     = "605325c6-d565-481c-9733-88ff5f3bac1c"
}

variable "machine_type" {
	description = "STACKIT machine type for SKE node pool (e.g. g1a.2d)"
	type        = string
	default     = "g1a.2d"
}

variable "cluster_name" {
	description = "Name of the SKE Kubernetes cluster"
	type        = string
	default     = "ske-kevin"
}

variable "kubernetes_version" {
	description = "Kubernetes version for the SKE cluster (e.g. 1.31). Omit to use default."
	type        = string
	default     = null
}

