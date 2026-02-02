# Generate a new SSH key pair when running Terraform
resource "tls_private_key" "ske" {
	algorithm = "RSA"
	rsa_bits  = 4096
}

# Register the public key with STACKIT (for SKE / instances)
resource "stackit_key_pair" "ske_key_pair" {
	project_id = var.project_id
	name       = "SKE_key_pair"
	public_key = tls_private_key.ske.public_key_openssh
}

# Write the private key to a file so you can use it for SSH
resource "local_sensitive_file" "private_key" {
	content         = tls_private_key.ske.private_key_pem
	filename        = "${path.module}/id_rsa"
	file_permission = "0600"
}
