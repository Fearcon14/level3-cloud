# 1. Provider Setup

terraform {
	required_version = ">=0.14.0"
	required_providers {
	  openstack = {
		source = "terraform-provider-openstack/openstack"
		version = "~> 1.53.0"
	  }
	  tls = {
		source = "hashicorp/tls"
		version = "~> 4.0"
	  }
	  local = {
		source = "hashicorp/local"
		version = "~> 2.0"
	  }
	}
}

# Use credentials from environment variables

provider "openstack" {}

# Generate SSH key pair
resource "tls_private_key" "k8s_ssh_key" {
  algorithm = "ED25519"
}

# Save the private key to a file
resource "local_file" "private_key" {
  content         = tls_private_key.k8s_ssh_key.private_key_openssh
  filename        = "${path.module}/k8s_key"
  file_permission = "0600"
}

# Save the public key to a file (optional, for reference)
resource "local_file" "public_key" {
  content         = tls_private_key.k8s_ssh_key.public_key_openssh
  filename        = "${path.module}/k8s_key.pub"
  file_permission = "0644"
}

# Create OpenStack keypair using the generated public key
resource "openstack_compute_keypair_v2" "k8s_key" {
  name       = "stack-key"
  public_key = tls_private_key.k8s_ssh_key.public_key_openssh
}

# 2. Network Setup

# Find public router ID to connect to outside world

data "openstack_networking_network_v2" "public_net" {
	name = "public"
}

# Create the Router

resource "openstack_networking_router_v2" "k8s_router" {
	name = "k8s-router"
	admin_state_up = true
	external_network_id = data.openstack_networking_network_v2.public_net.id
}

# Create the Private Network

resource "openstack_networking_network_v2" "k8s_network" {
	name = "k8s_network"
	admin_state_up = true
}

# Create the Subnet

resource "openstack_networking_subnet_v2" "k8s_subnet" {
	name = "k8s-subnet"
	network_id = openstack_networking_network_v2.k8s_network.id
	cidr = "192.168.14.0/24"
	ip_version = 4
	dns_nameservers = ["8.8.8.8", "1.1.1.1"]
}

# Plug Subnet into Router

resource "openstack_networking_router_interface_v2" "k8s_interface" {
	router_id = openstack_networking_router_v2.k8s_router.id
	subnet_id = openstack_networking_subnet_v2.k8s_subnet.id
}

# 3. Security Groups

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

# 4. Master Node

resource "openstack_compute_instance_v2" "k8s_master" {
	name = "k8s-master"
	image_name = "Ubuntu-22.04" # Must match Glance Upload
	flavor_name = "m1.master"	# Must match Flavor Creation
	key_pair = openstack_compute_keypair_v2.k8s_key.name
	security_groups = ["default", openstack_networking_secgroup_v2.k8s_sg.name]

	network {
		uuid = openstack_networking_network_v2.k8s_network.id
	}

	depends_on = [openstack_networking_router_interface_v2.k8s_interface]
}

# 5. Worker Nodes

resource "openstack_compute_instance_v2" "k8s_worker" {
	count = 1 # Create 1 worker nodes
	name = "k8s-worker-${count.index}"
	image_name = "Ubuntu-22.04" # Must match Glance Upload
	flavor_name = "m1.worker"	# Must match Flavor Creation
	key_pair = openstack_compute_keypair_v2.k8s_key.name
	security_groups = ["default", openstack_networking_secgroup_v2.k8s_sg.name]

	network {
		uuid = openstack_networking_network_v2.k8s_network.id
	}

	depends_on = [openstack_networking_router_interface_v2.k8s_interface]
}

# 6. External Access (Floating IP)

# Request a Floating IP from the "public" pool

resource "openstack_networking_floatingip_v2" "k8s_master_floating_ip" {
	pool = "public"
}

# Attach to the Master Node

resource "openstack_compute_floatingip_associate_v2" "k8s_master_floating_ip_associate" {
	floating_ip = openstack_networking_floatingip_v2.k8s_master_floating_ip.address
	instance_id = openstack_compute_instance_v2.k8s_master.id
}

# 7. Outputs

output "master_public_ip" {
	value = openstack_networking_floatingip_v2.k8s_master_floating_ip.address
}

output "worker_ips" {
	value = openstack_compute_instance_v2.k8s_worker[*].access_ip_v4
}

output "private_key_path" {
	value = local_file.private_key.filename
	description = "Path to the generated private key file"
}

output "public_key_path" {
	value = local_file.public_key.filename
	description = "Path to the generated public key file"
}

# AUTOMATION: GENERATE ANSIBLE SSH CONFIG

resource "local_file" "ansible_ssh_config" {
  filename = "./ansible/ssh.cfg"
  content  = <<EOF
Host *
  StrictHostKeyChecking no
  UserKnownHostsFile /dev/null
  ServerAliveInterval 30
  ServerAliveCountMax 3

# 1. The Gateway (Master Node)
Host master-node
  Hostname ${openstack_networking_floatingip_v2.k8s_master_floating_ip.address}
  User ubuntu
  IdentityFile ${path.module}/k8s_key

# 2. Worker 0
Host worker-0
  Hostname ${openstack_compute_instance_v2.k8s_worker[0].access_ip_v4}
  User ubuntu
  IdentityFile ${path.module}/k8s_key
  ProxyJump master-node

EOF
}

# --- AUTOMATED HOST NETWORK FIX ---

# 1. Define your host internet interface (Change if not enp3s0)
variable "host_interface" {
  default = "enp3s0"
}

# 2. Configure NAT on the Host Machine
resource "null_resource" "host_nat_config" {
  triggers = {
    interface = var.host_interface
  }

  # ENABLE NAT (Runs on Apply)
  provisioner "local-exec" {
    command = <<EOT
      echo "--- Enabling Host NAT on ${var.host_interface} ---"
      sudo sysctl -w net.ipv4.ip_forward=1
      # Check if rule exists, if not, add it
      sudo iptables -t nat -C POSTROUTING -o ${var.host_interface} -j MASQUERADE 2>/dev/null || \
      sudo iptables -t nat -A POSTROUTING -o ${var.host_interface} -j MASQUERADE
    EOT
  }

  # CLEANUP NAT (Runs on Destroy)
  provisioner "local-exec" {
    when    = destroy
    command = <<EOT
      echo "--- Cleaning up Host NAT rules ---"
      sudo iptables -t nat -D POSTROUTING -o enp3s0 -j MASQUERADE || true
    EOT
  }
}
