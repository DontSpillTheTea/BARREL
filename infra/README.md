# BARREL Infrastructure

This directory contains OpenTofu (Terraform-compatible) and Terragrunt configuration for deploying BARREL to Azure.

## Prerequisites

Install these tools before deploying:

| Tool | Install |
|------|---------|
| [Azure CLI](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) | `brew install azure-cli` or [docs](https://learn.microsoft.com/en-us/cli/azure/install-azure-cli) |
| [OpenTofu](https://opentofu.org/docs/intro/install/) | `brew install opentofu` |
| [Terragrunt](https://terragrunt.gruntwork.io/docs/getting-started/install/) | `brew install terragrunt` |
| [Docker](https://docs.docker.com/get-docker/) | Docker Desktop or `docker.io` |
| [Task](https://taskfile.dev/installation/) | `brew install go-task` |

## Pre-existing Azure Resources

The IaC references two Azure AI resources that must exist **before** running `apply`. These are created separately because AI model deployments have region/quota constraints that may require manual selection.

| Resource | Type | Purpose |
|----------|------|---------|
| `barrel-openai-sweden` | Azure OpenAI | AI vision + text parser (gpt-4.1-mini) |
| `barrel-vision-dev` | Computer Vision | OCR text extraction |

Create them if they don't exist:
```bash
# Azure OpenAI (choose a region with gpt-4.1-mini quota)
az cognitiveservices account create \
  --name barrel-openai-sweden \
  --resource-group barrel-ai-rg \
  --kind OpenAI --sku S0 \
  --location swedencentral --yes

az cognitiveservices account deployment create \
  --name barrel-openai-sweden \
  --resource-group barrel-ai-rg \
  --deployment-name barrel-ai-native-parser \
  --model-name gpt-4.1-mini --model-version "2025-04-14" \
  --model-format OpenAI --sku-capacity 50 --sku-name GlobalStandard

# Azure Vision OCR
az cognitiveservices account create \
  --name barrel-vision-dev \
  --resource-group barrel-ai-rg \
  --kind ComputerVision --sku F0 \
  --location swedencentral --yes
```

## Azure Provider Registration

Register these providers on your subscription (one-time):
```bash
az provider register -n Microsoft.App
az provider register -n Microsoft.ContainerRegistry
az provider register -n Microsoft.Storage
az provider register -n Microsoft.KeyVault
az provider register -n Microsoft.OperationalInsights
```

## Environment Variables

Set these before deploying (or rely on defaults in `terragrunt.hcl`):

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `BARREL_DEMO_PASSWORD` | No | `fallback-demo-password-123` | Evaluator login password |
| `BARREL_REVIEW_TOKEN` | No | `fallback-review-token-123` | API auth token |
| `AZURE_VISION_KEY` | Yes | — | Vision OCR API key (sensitive) |
| `AZURE_OPENAI_API_VERSION` | No | `2025-01-01-preview` | OpenAI API version |

Get the Vision key:
```bash
export AZURE_VISION_KEY=$(az cognitiveservices account keys list \
  --name barrel-vision-dev --resource-group barrel-ai-rg \
  --query "key1" -o tsv)
```

## First Deploy

```bash
# 1. Login to Azure
az login

# 2. Initialize Terragrunt
task azure:infra:init

# 3. Preview what will be created
task azure:infra:plan

# 4. Create all Azure resources
task azure:infra:apply

# 5. Build Docker images
task azure:build

# 6. Push images and deploy to Container Apps
task azure:deploy

# 7. Get deployed URLs
task azure:outputs
```

## What Gets Created

| Resource | Name Pattern | Purpose |
|----------|-------------|---------|
| Container App (API) | `barrel-api` | Go backend on port 8080 |
| Container App (Web) | `barrel-web` | React frontend on port 5173 |
| Container Registry | `barrelacrdev` | Docker image storage |
| Storage Account | `barrelsadev` | Blob + Table storage |
| Blob Container | `jobs` | Label images, results, decisions |
| Table Storage | `reviews` | Review metadata for fast listing |
| Key Vault | `barrel-kv-dev` | Secret storage |
| Log Analytics | `barrel-law` | Container logs |
| Managed Identity | `barrel-api-id` | API → Key Vault access |

## Updating

After code changes:
```bash
task azure:build    # Rebuild images
task azure:deploy   # Push and update Container Apps
```

After infrastructure changes (main.tf/variables.tf):
```bash
task azure:infra:plan    # Preview changes
task azure:infra:apply   # Apply changes
```

## Tearing Down

```bash
task azure:infra:destroy
```

This destroys all IaC-managed resources. The pre-existing OpenAI and Vision resources are NOT destroyed (they're referenced as data sources, not managed resources).

## Troubleshooting

**"Provider not registered"**: Run `az provider register -n Microsoft.App --wait`

**"Insufficient quota"**: Check model availability in your region. gpt-4.1-mini may not be available in all regions. Try `swedencentral`, `eastus2`, or `westus`.

**"State lock"**: Another Terragrunt process is running. Wait for it to finish or remove the lock file in `.terragrunt-cache/`.

**Container App not starting**: Check logs with `az containerapp logs show -n barrel-api -g barrel-ai-rg --tail 20`

**Build fails with Go version mismatch**: The API Dockerfile uses `golang:1.25-alpine`. Ensure your go.mod matches.

## File Structure

```
infra/
├── opentofu/dev/
│   ├── main.tf          # All Azure resource definitions
│   ├── variables.tf     # Input variables with defaults
│   ├── outputs.tf       # Output values (URLs, names)
│   └── providers.tf     # Azure provider config
└── terragrunt/dev/
    └── terragrunt.hcl   # Terragrunt wrapper with env var inputs
```
