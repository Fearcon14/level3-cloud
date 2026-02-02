resource "stackit_security_group" "SKE_security_group_kevin" {
	project_id	=	var.project_id
	name		=	"SKE_security_group_kevin"
	description	=	"Security group for SKE cluster"
}

resource "stackit_security_group_rule" "SKE_SSH" {
	project_id         = var.project_id
	security_group_id  = stackit_security_group.SKE_security_group_kevin.security_group_id
	direction          = "ingress"
	protocol = {
		name = "tcp"
	}
	port_range = {
		min = 22
		max = 22
	}
	depends_on         = [stackit_security_group.SKE_security_group_kevin]
}

resource "stackit_security_group_rule" "SKE_ICMP" {
	project_id         = var.project_id
	security_group_id  = stackit_security_group.SKE_security_group_kevin.security_group_id
	direction          = "ingress"
	protocol = {
		name = "icmp"
	}
	depends_on         = [stackit_security_group.SKE_security_group_kevin]
}
