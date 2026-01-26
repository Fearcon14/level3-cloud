# 1. Provider Setup

terraform {
	required_version = ">=0.14.0"
	required_providers {
	  openstack = {
		source = "terraform-provider-openstack/openstack"
		version = "~-> 1.53.0"
	  }
	}
}

# Use credentials from environment variables

provider "openstack" {}

# 2. Security Groups

resource "openstack_networking_secgroup_v2" "k8s_sg" {
	name = "k8s-cluster-sg"
	description = "Security group for the k8s cluster"
}

# Allow SSH (Port 22) - So you can log in

resource "openstack_networking_secgroup_rule_v2" "sg_rule_ssh" {
	direction = "ingress"
	ethertype = "IPv4"
	protocol = "tcp"
	port_range_min = 22
	port_range_max = 22
	remote_ip_prefix = "0.0.0.0/0"
	security_group_id = openstack_networking_secgroup_v2.k8s_sg.id
}

# Allow Kubernetes API (Port 6443) - The "Front Door" for the cluster

resource "openstack_networking_secgroup_rule_v2" "sg_rule_k8s_api" {
	direction = "ingress"
	ethertype = "IPv4"
	protocol = "tcp"
	port_range_min = 6443
	port_range_max = 6443
	remote_ip_prefix = "0.0.0.0/0"
	security_group_id = openstack_networking_secgroup_v2.k8s_sg.id
}

# Allow All Internal Traffic - So the nodes can talk to each other

resource "openstack_networking_secgroup_rule_v2" "sg_rule_internal" {
	direction = "ingress"
	ethertype = "IPv4"
	protocol = "tcp"
	remote_group_id = openstack_networking_secgroup_v2.k8s_sg.id
	security_group_id = openstack_networking_secgroup_v2.k8s_sg.id
}

# 3. Master Node

resource "openstack_compute_instance_v2" "k8s_master" {
	name = "k8s-master"
	image_name = "Ubuntu-22.04" # Must match Glance Upload
	flavor_name = "m1.master"	# Must match Flavor Creation
	key_pair = "stack-key"
	security_groups = ["default", openstack_networking_secgroup_v2.k8s_sg.name]

	network {
		name = "Network-Test"	# Must match Network Creation
	}
}

# 4. Worker Nodes

resource "openstack_compute_instance_v2" "k8s_worker" {
	count = 2 # Create 2 worker nodes
	name = "k8s-worker-${count.index}"
	image_name = "Ubuntu-22.04" # Must match Glance Upload
	flavor_name = "m1.worker"	# Must match Flavor Creation
	key_pair = "stack-key"
	security_groups = ["default", openstack_networking_secgroup_v2.k8s_sg.name]

	network {
		name = "Network-Test"	# Must match master node network
	}
}

# 5. Outputs

output "master_ip" {
	value = openstack_compute_instance_v2.k8s_master.access_ip_v4
}

output "worker_ips" {
	value = openstack_compute_instance_v2.k8s_worker[*].access_ip_v4
}
