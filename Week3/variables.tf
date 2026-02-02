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

