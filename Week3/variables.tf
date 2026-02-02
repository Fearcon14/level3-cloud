variable "service_account_key_path" {
	description = "StackIT service account key (Key Flow auth)"
	type        = string
	sensitive   = true
	default     = "service_account_key.json"
}

