output "api_url" {
  value = "https://${azurerm_container_app.api.ingress[0].fqdn}"
}

output "web_url" {
  value = "https://${azurerm_container_app.web.ingress[0].fqdn}"
}

output "azure_vision_endpoint" {
  value = azurerm_cognitive_account.vision.endpoint
}

output "storage_account_name" {
  value = azurerm_storage_account.main.name
}

output "key_vault_name" {
  value = azurerm_key_vault.main.name
}

output "review_container_name" {
  value = azurerm_storage_container.reviews.name
}
