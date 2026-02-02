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
	description = "STACKIT flavor"
	type		= string
	default		= "g2i.2"
}

variable "image_id" {
	description = "STACKIT image ID for the boot volume (e.g. Ubuntu 24.04). Get from STACKIT console or API: iaas.api.<region>.stackit.cloud/v1/projects/<project_id>/images"
	type		=	string
	default		=	"33b2ffe4-73cc-4d2b-b406-215e66661a7a"
}

