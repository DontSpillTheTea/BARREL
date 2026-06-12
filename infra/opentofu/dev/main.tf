data "azurerm_resource_group" "main" {
  name = var.resource_group_name
}

data "azurerm_cognitive_account" "openai" {
  name                = var.openai_account_name
  resource_group_name = data.azurerm_resource_group.main.name
}

data "azurerm_client_config" "current" {}

resource "azurerm_log_analytics_workspace" "main" {
  name                = "${var.resource_prefix}-law"
  location            = data.azurerm_resource_group.main.location
  resource_group_name = data.azurerm_resource_group.main.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
}

resource "azurerm_storage_account" "main" {
  name                     = "${var.resource_prefix}sa${var.env_suffix}"
  resource_group_name      = data.azurerm_resource_group.main.name
  location                 = data.azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
}

resource "azurerm_storage_container" "reviews" {
  name                  = "jobs"
  storage_account_name  = azurerm_storage_account.main.name
  container_access_type = "private"
}

resource "azurerm_storage_table" "reviews" {
  name                 = "reviews"
  storage_account_name = azurerm_storage_account.main.name
}

resource "azurerm_container_registry" "main" {
  name                = "${var.resource_prefix}acr${var.env_suffix}"
  resource_group_name = data.azurerm_resource_group.main.name
  location            = data.azurerm_resource_group.main.location
  sku                 = "Basic"
  admin_enabled       = true
}

resource "azurerm_container_app_environment" "main" {
  name                       = "${var.resource_prefix}-cae"
  location                   = data.azurerm_resource_group.main.location
  resource_group_name        = data.azurerm_resource_group.main.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
}

resource "azurerm_user_assigned_identity" "api" {
  name                = "${var.resource_prefix}-api-id"
  location            = data.azurerm_resource_group.main.location
  resource_group_name = data.azurerm_resource_group.main.name
}

resource "azurerm_container_app" "api" {
  name                         = "${var.resource_prefix}-api"
  container_app_environment_id = azurerm_container_app_environment.main.id
  resource_group_name          = data.azurerm_resource_group.main.name
  revision_mode                = "Single"

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.api.id]
  }

  secret {
    name  = "acr-password"
    value = azurerm_container_registry.main.admin_password
  }
  secret {
    name  = "storage-conn-string"
    value = azurerm_storage_account.main.primary_connection_string
  }
  secret {
    name  = "azure-openai-key"
    value = data.azurerm_cognitive_account.openai.primary_access_key
  }
  secret {
    name  = "vision-key"
    value = var.vision_key
  }
  secret {
    name  = "demo-password"
    value = var.demo_password
  }
  secret {
    name  = "review-token"
    value = var.review_token
  }

  registry {
    server               = azurerm_container_registry.main.login_server
    username             = azurerm_container_registry.main.admin_username
    password_secret_name = "acr-password"
  }

  template {
    min_replicas = 1
    max_replicas = 3

    container {
      name   = "api"
      image  = var.container_image_api
      cpu    = 0.5
      memory = "1Gi"

      env {
        name  = "PORT"
        value = "8080"
      }
      env {
        name  = "RULESET_PATH"
        value = "/rules/ttb"
      }
      env {
        name  = "MAX_UPLOAD_MB"
        value = "25"
      }
      env {
        name  = "AI_NATIVE_ENABLED"
        value = "true"
      }
      env {
        name  = "BARREL_ANALYSIS_PROVIDER"
        value = "tiered"
      }
      env {
        name  = "AZURE_VISION_ENDPOINT"
        value = var.vision_endpoint
      }
      env {
        name        = "AZURE_VISION_KEY"
        secret_name = "vision-key"
      }
      env {
        name  = "AZURE_VISION_API_VERSION"
        value = "2024-02-01"
      }
      env {
        name  = "AZURE_OPENAI_ENDPOINT"
        value = data.azurerm_cognitive_account.openai.endpoint
      }
      env {
        name        = "AZURE_OPENAI_API_KEY"
        secret_name = "azure-openai-key"
      }
      env {
        name  = "AZURE_OPENAI_DEPLOYMENT"
        value = var.openai_deployment_name
      }
      env {
        name  = "AZURE_OPENAI_API_VERSION"
        value = var.azure_openai_api_version
      }
      env {
        name  = "STORAGE_PROVIDER"
        value = "azure_blob"
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
        name  = "AZURE_STORAGE_TABLE"
        value = azurerm_storage_table.reviews.name
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
  resource_group_name          = data.azurerm_resource_group.main.name
  revision_mode                = "Single"

  secret {
    name  = "acr-password"
    value = azurerm_container_registry.main.admin_password
  }

  registry {
    server               = azurerm_container_registry.main.login_server
    username             = azurerm_container_registry.main.admin_username
    password_secret_name = "acr-password"
  }

  template {
    min_replicas = 1
    max_replicas = 2

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
    target_port      = 5173
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }
}

resource "azurerm_key_vault" "main" {
  name                     = "${var.resource_prefix}-kv-${var.env_suffix}"
  location                 = data.azurerm_resource_group.main.location
  resource_group_name      = data.azurerm_resource_group.main.name
  tenant_id                = data.azurerm_client_config.current.tenant_id
  sku_name                 = "standard"
  purge_protection_enabled = false

  access_policy {
    tenant_id          = data.azurerm_client_config.current.tenant_id
    object_id          = data.azurerm_client_config.current.object_id
    secret_permissions = ["Set", "Get", "Delete", "Purge", "Recover", "List"]
  }

  access_policy {
    tenant_id          = data.azurerm_client_config.current.tenant_id
    object_id          = azurerm_user_assigned_identity.api.principal_id
    secret_permissions = ["Get"]
  }
}

output "web_url" {
  value = "https://${azurerm_container_app.web.ingress[0].fqdn}"
}

output "api_url" {
  value = "https://${azurerm_container_app.api.ingress[0].fqdn}"
}

output "acr_server" {
  value = azurerm_container_registry.main.login_server
}

output "storage_account" {
  value = azurerm_storage_account.main.name
}

output "storage_table" {
  value = azurerm_storage_table.reviews.name
}

output "key_vault_name" {
  value = azurerm_key_vault.main.name
}
