.PHONY: help init plan apply destroy ansible-ping ansible-deploy all clean

# Default target
help:
	@echo "Kubernetes Cluster Automation"
	@echo ""
	@echo "Available targets:"
	@echo "  make init          - Initialize Terraform"
	@echo "  make plan          - Show Terraform execution plan"
	@echo "  make apply         - Create infrastructure with Terraform"
	@echo "  make destroy       - Destroy infrastructure"
	@echo "  make ansible-ping  - Test Ansible connectivity"
	@echo "  make ansible-deploy - Run Ansible playbook"
	@echo "  make all           - Full workflow: init, apply, ansible-deploy"
	@echo "  make clean         - Destroy infrastructure and remove local files"

# Terraform targets
init:
	@echo "=== Initializing Terraform ==="
	terraform init

plan: init
	@echo "=== Terraform Plan ==="
	terraform plan

apply: init
	@echo "=== Applying Terraform ==="
	terraform apply -auto-approve
	@echo "=== Waiting for instances to be ready (30s) ==="
	@sleep 30
	@echo "=== Infrastructure created! ==="

destroy:
	@echo "=== Destroying Infrastructure ==="
	terraform destroy -auto-approve

# Ansible targets
ansible-ping:
	@echo "=== Testing Ansible Connectivity ==="
	cd ansible && ansible all -m ping

ansible-deploy: ansible-ping
	@echo "=== Running Ansible Playbook ==="
	cd ansible && ansible-playbook playbook.yml

# Full workflow
all: apply ansible-deploy
	@echo "=== Complete! ==="

# Cleanup
clean: destroy
	@echo "=== Cleaning up local files ==="
	rm -f terraform.tfstate terraform.tfstate.backup
	rm -f k8s_key k8s_key.pub
	rm -f ansible/ssh.cfg
	@echo "Cleanup complete. Run 'make init' to start fresh."
