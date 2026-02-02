resource "stackit_network" "SKE_network" {
	project_id			=	var.project_id
	name				=	"SKE_network"
	description			=	"Network for SKE cluster"
	nameservers			= ["8.8.8.8", "1.1.1.1"]
	ipv4_prefix_length	= 24
}

resource "stackit_network_interface" "SKE_network_interface" {
	project_id	=	var.project_id
	network_id	=	stackit_network.SKE_network.id
	security_group_ids = [stackit_security_group.SKE_security_group.id]
	depends_on	= [stackit_network.SKE_network, stackit_security_group.SKE_security_group]
}

resource "stackit_public_ip" "SKE_public_ip" {
	project_id	=	var.project_id
	network_interface_id = stackit_network_interface.SKE_network_interface.id
	depends_on	= [stackit_network_interface.SKE_network_interface]
}
