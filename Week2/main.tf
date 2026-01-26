# Define the Provider (Who we are talking to)

terraform {
  required_version = ">=0.14.0"
  required_providers {
	openstack = {
		source = "terraform-provider-openstack/openstack"
		version = "~> 1.53.0"
	}
  }
}

# Configure the OpenStack Provider
# We leave this empty because we will use the environment variables to authenticate
# from 'source openrc'

provider "openstack" {}

# Create Security Group

resource "openstack_networking_secgroup_v2" "terraform-ssh" {
	name = "terraform-ssh"
	description = "Allow SSH access to the instance"
}

# Add SSH Rule to Security Group

resource "openstack_networking_secgroup_rule_v2" "terraform-ssh-rule" {
	direction = "ingress"
	protocol = "tcp"
	ethertype = "IPv4"
	port_range_min = 22
	port_range_max = 22
	remote_ip_prefix = "0.0.0.0/0"
	security_group_id = openstack_networking_secgroup_v2.terraform-ssh.id
}

# Define the Resource (What we are creating)

resource "openstack_compute_instance_v2" "terraform-instance" {
  name = "terraform-vm-01"
  image_name = "cirros-0.6.3-x86_64-disk"
  flavor_name = "m1.nano"
  key_pair = "stack-key"
  security_groups = ["default", openstack_networking_secgroup_v2.terraform-ssh.name]

  network {
	name = "Network-Test"
  }
}
