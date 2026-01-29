.PHONY: help init plan apply destroy wait-for-ssh ansible-ping ansible-deploy all clean

# Use bash so 'source' works in recipes
SHELL := /bin/bash

# OpenStack credentials (source before Terraform operations)
OPENRC = source ~/devstack/openrc demo demo

# Default target
help:
	@echo "Kubernetes Cluster Automation"
	@echo ""
	@echo "Available targets:"
	@echo "  make init          - Initialize Terraform"
	@echo "  make plan          - Show Terraform execution plan"
	@echo "  make apply         - Create infrastructure with Terraform"
	@echo "  make destroy       - Destroy infrastructure"
	@echo "  make wait-for-ssh  - Wait until all instances are reachable via SSH"
	@echo "  make ansible-ping  - Test Ansible connectivity"
	@echo "  make ansible-deploy - Run Ansible playbook"
	@echo "  make all           - Full workflow: init, apply, wait-for-ssh, ansible-deploy"
	@echo "  make clean         - Destroy infrastructure and remove local files"

# Terraform targets (require OpenStack credentials via openrc demo demo)
init:
	@echo "=== Initializing Terraform ==="
	$(OPENRC) && terraform init

plan: init
	@echo "=== Terraform Plan ==="
	$(OPENRC) && terraform plan

apply: init
	@echo "=== Applying Terraform ==="
	$(OPENRC) && terraform apply -auto-approve
	@echo "=== Infrastructure created! ==="

# Wait for SSH (retry Ansible ping until all hosts respond or timeout)
WAIT_MAX_ATTEMPTS = 42
WAIT_INTERVAL = 15

wait-for-ssh:
	@echo "=== Waiting for instances to be reachable via SSH ==="
	@for i in $$(seq 1 $(WAIT_MAX_ATTEMPTS)); do \
		if (cd ansible && ansible all -m ping -o >/dev/null 2>&1); then \
			echo "All hosts reachable."; exit 0; \
		fi; \
		echo "Attempt $$i/$(WAIT_MAX_ATTEMPTS): hosts not ready, retrying in $(WAIT_INTERVAL)s..."; \
		sleep $(WAIT_INTERVAL); \
	done; \
	echo "Timeout: hosts did not become reachable within $$(($(WAIT_MAX_ATTEMPTS) * $(WAIT_INTERVAL)))s."; exit 1

destroy:
	@echo "=== Destroying Infrastructure ==="
	$(OPENRC) && terraform destroy -auto-approve

# Ansible targets
ansible-ping:
	@echo "=== Testing Ansible Connectivity ==="
	cd ansible && ansible all -m ping

ansible-deploy: ansible-ping
	@echo "=== Running Ansible Playbook ==="
	cd ansible && ansible-playbook playbook.yml

# Full workflow
all: apply wait-for-ssh ansible-deploy
	@echo "=== Complete! ==="

# Cleanup
clean: destroy
	@echo "=== Cleaning up local files ==="
	rm -f terraform.tfstate terraform.tfstate.backup
	rm -f k8s_key k8s_key.pub
	rm -f ansible/ssh.cfg
	@echo "Cleanup complete. Run 'make init' to start fresh."
