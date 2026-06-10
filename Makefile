.PHONY: default help install-tools setup

define TASK_MISSING_HINT
@echo "BARREL uses go-task as the supported command interface."
@echo "Make is only provided here for initial setup."
@echo "Run: task help"
@exit 1
endef

default: help

help:
	@echo "BARREL Initial Setup Makefile"
	@echo "Usage:"
	@echo "  make install-tools  - Installs required tools (task, opentofu, terragrunt, azure-cli)"
	@echo "  make setup          - Runs install-tools and provides next steps"

install-tools:
	@echo "Installing go-task..."
	@curl -sL https://taskfile.dev/install.sh | sh -s -- -b /usr/local/bin || sudo sh -c 'curl -sL https://taskfile.dev/install.sh | sh -s -- -b /usr/local/bin'
	@echo "Installing OpenTofu..."
	@curl --proto '=https' --tlsv1.2 -fsSL https://get.opentofu.org/install-opentofu.sh -o install-opentofu.sh
	@chmod +x install-opentofu.sh && ./install-opentofu.sh --install-method standalone && rm install-opentofu.sh
	@echo "Installing Terragrunt..."
	@wget -q -O terragrunt https://github.com/gruntwork-io/terragrunt/releases/download/v0.58.12/terragrunt_linux_amd64
	@chmod +x terragrunt && sudo mv terragrunt /usr/local/bin/terragrunt || mv terragrunt ~/.local/bin/terragrunt
	@echo "Checking Azure CLI..."
	@az version >/dev/null 2>&1 || echo "Azure CLI not found. Please install via: curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash"
	@echo "Tools installed successfully. You can now use 'task'."

setup: install-tools
	@echo ""
	@echo "Setup complete! Next steps:"
	@echo "1. Run 'az login' to authenticate with Azure."
	@echo "2. Run 'task azure:infra:init' to initialize infrastructure."
	@echo "3. Run 'task azure:infra:apply' to provision Azure resources."
	@echo "4. Run 'task azure:build' and 'task azure:deploy' to deploy code."

%:
	$(TASK_MISSING_HINT)
