output "ssh_private_key_path" {
	description = "Path to the generated private key file for SSH"
	value       = local_sensitive_file.private_key.filename
}

output "key_pair_name" {
	description = "Name of the key pair registered in STACKIT"
	value       = stackit_key_pair.ske_key_pair.name
}
