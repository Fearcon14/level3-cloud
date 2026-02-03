# Optional: SSH key pair for node access or future bastion VM.
# Not required for SKE cluster access (use kubeconfig). Remove this file for a minimal SKE-only setup.
resource "tls_private_key" "ske_kevin" {
	algorithm = "RSA"
	rsa_bits  = 4096
}

# Register the public key with STACKIT (for SKE / instances)
resource "stackit_key_pair" "ske_key_pair" {
	name       = "SKE_key_pair_kevin"
	public_key = tls_private_key.ske_kevin.public_key_openssh
}

# Write the private key to a file so you can use it for SSH
resource "local_sensitive_file" "private_key" {
	content         = tls_private_key.ske_kevin.private_key_pem
	filename        = "${path.module}/id_rsa"
	file_permission = "0600"
}
