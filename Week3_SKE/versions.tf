terraform {
	required_providers {
		stackit = {
			source  = "stackitcloud/stackit"
			version = ">= 0.20.0"
		}
		tls = {
			source  = "hashicorp/tls"
			version = "~> 4.0"
		}
		local = {
			source  = "hashicorp/local"
			version = "~> 2.0"
		}
	}
}
