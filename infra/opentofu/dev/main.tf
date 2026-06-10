resource "random_string" "suffix" {
  length  = 6
  special = false
  upper   = false
}

resource "azurerm_resource_group" "main" {
  name     = "${var.resource_prefix}-rg-${random_string.suffix.result}"
  location = var.location
}

resource "azurerm_log_analytics_workspace" "main" {
  name                = "${var.resource_prefix}-law-${random_string.suffix.result}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_cognitive_account" "vision" {
  name                = "${var.resource_prefix}-vision-${random_string.suffix.result}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  kind                = "ComputerVision"
  sku_name            = var.azure_vision_sku
}

resource "azurerm_storage_account" "main" {
  name                     = "${var.resource_prefix}sa${random_string.suffix.result}"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_container" "reviews" {
  name                  = "barrel-review"
  storage_account_name  = azurerm_storage_account.main.name
  container_access_type = "private"
}

resource "azurerm_container_registry" "main" {
  name                = "${var.resource_prefix}acr${random_string.suffix.result}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = "Basic"
  admin_enabled       = true
}

resource "azurerm_container_app_environment" "main" {
  name                       = "${var.resource_prefix}-cae-${random_string.suffix.result}"
  location                   = azurerm_resource_group.main.location
  resource_group_name        = azurerm_resource_group.main.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
}

resource "azurerm_container_app" "api" {
  name                         = "${var.resource_prefix}-api"
  container_app_environment_id = azurerm_container_app_environment.main.id
  resource_group_name          = azurerm_resource_group.main.name
  revision_mode                = "Single"

  template {
    container {
      name   = "api"
      image  = var.container_image_api
      cpu    = 0.5
      memory = "1Gi"

      env {
        name  = "STORAGE_PROVIDER"
        value = "azure_blob"
      }
      env {
        name  = "OCR_PROVIDER"
        value = "azure_vision"
      }
      env {
        name  = "OCR_FALLBACK_PROVIDER"
        value = "paddleocr_worker"
      }
      env {
        name  = "OCR_ALLOW_FAST_FALLBACK"
        value = "false"
      }
      env {
        name  = "AI_PROVIDER"
        value = "none"
      }
      env {
        name  = "AZURE_VISION_ENDPOINT"
        value = azurerm_cognitive_account.vision.endpoint
      }
      env {
        name        = "AZURE_VISION_KEY"
        secret_name = "vision-key"
      }
      env {
        name  = "AZURE_VISION_API_VERSION"
        value = "2024-02-01-preview" # Using standard vision API version
      }
      env {
        name  = "AZURE_STORAGE_ACCOUNT"
        value = azurerm_storage_account.main.name
      }
      env {
        name        = "AZURE_STORAGE_CONNECTION_STRING"
        secret_name = "storage-conn-string"
      }
      env {
        name  = "AZURE_STORAGE_CONTAINER"
        value = azurerm_storage_container.reviews.name
      }
      env {
        name  = "BARREL_DEMO_USERNAME"
        value = var.demo_username
      }
      env {
        name        = "BARREL_DEMO_PASSWORD"
        secret_name = "demo-password"
      }
      env {
        name        = "BARREL_REVIEW_TOKEN"
        secret_name = "review-token"
      }
    }
  }

  secret {
    name  = "storage-conn-string"
    value = azurerm_storage_account.main.primary_connection_string
  }
  secret {
    name  = "vision-key"
    value = azurerm_cognitive_account.vision.primary_access_key
  }
  secret {
    name  = "demo-password"
    value = var.demo_password
  }
  secret {
    name  = "review-token"
    value = var.review_token
  }

  ingress {
    external_enabled = true
    target_port      = 8080
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }
}

resource "azurerm_container_app" "web" {
  name                         = "${var.resource_prefix}-web"
  container_app_environment_id = azurerm_container_app_environment.main.id
  resource_group_name          = azurerm_resource_group.main.name
  revision_mode                = "Single"

  template {
    container {
      name   = "web"
      image  = var.container_image_web
      cpu    = 0.25
      memory = "0.5Gi"
      env {
        name  = "VITE_API_BASE_URL"
        value = "https://${azurerm_container_app.api.ingress[0].fqdn}"
      }
    }
  }

  ingress {
    external_enabled = true
    target_port      = 80
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }
}

resource "azurerm_key_vault" "main" {
  name                = "${var.resource_prefix}-kv-${random_string.suffix.result}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"
}

data "azurerm_client_config" "current" {}
