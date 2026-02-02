resource "stackit_network" "SKE_network_kevin" {
	project_id          = var.project_id
	name                = "SKE_network_kevin"
	ipv4_nameservers    = ["8.8.8.8", "1.1.1.1"]
	ipv4_prefix_length  = 24
}

resource "stackit_network_interface" "SKE_network_interface_kevin" {
	project_id  = var.project_id
	network_id  = stackit_network.SKE_network_kevin.network_id
	security_group_ids = [stackit_security_group.SKE_security_group_kevin.security_group_id]
	depends_on	= [stackit_network.SKE_network_kevin, stackit_security_group.SKE_security_group_kevin]
}

resource "stackit_public_ip" "SKE_public_ip_kevin" {
	project_id	=	var.project_id
	network_interface_id = stackit_network_interface.SKE_network_interface_kevin.network_interface_id
	depends_on	= [stackit_network_interface.SKE_network_interface_kevin]
}
