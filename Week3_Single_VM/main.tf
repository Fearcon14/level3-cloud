resource "stackit_server" "SKE_Kevin_server" {
	project_id         = var.project_id
	name               = "SKE-Kevin-server"
	boot_volume = {
		source_type = "volume"
		source_id   = stackit_volume.SKE_volume_kevin.volume_id
	}
	availability_zone   = "eu01-1"
	machine_type       = var.machine_type
	keypair_name       = stackit_key_pair.ske_key_pair.name
	network_interfaces = [stackit_network_interface.SKE_network_interface_kevin.network_interface_id]
}
