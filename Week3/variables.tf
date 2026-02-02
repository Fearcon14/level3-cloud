variable "service_account_key" {
	description = "StackIT service account key (Key Flow auth)"
	type        = string
	sensitive   = true
	default     = ""
}

variable "private_key" {
	description = "StackIT private key (Key Flow auth)"
	type        = string
	sensitive   = true
	default     = ""
}
