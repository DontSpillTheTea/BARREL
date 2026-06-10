terraform {
  source = "../../opentofu/dev"
}

inputs = {
  location            = "eastus"
  resource_prefix     = "barrel"
  container_image_api = "mcr.microsoft.com/azuredocs/containerapps-helloworld:latest"
  container_image_web = "mcr.microsoft.com/azuredocs/containerapps-helloworld:latest"
  review_token        = get_env("BARREL_REVIEW_TOKEN", "fallback-review-token-123")
  demo_password       = get_env("BARREL_DEMO_PASSWORD", "fallback-demo-password-123")
  
  ai_second_read_enabled      = get_env("AI_SECOND_READ_ENABLED", "false") == "true"
  ai_second_read_auto_on_fail = get_env("AI_SECOND_READ_AUTO_ON_FAIL", "false") == "true"
  azure_openai_endpoint       = get_env("AZURE_OPENAI_ENDPOINT", "")
  azure_openai_api_key        = get_env("AZURE_OPENAI_API_KEY", "")
  azure_openai_deployment     = get_env("AZURE_OPENAI_DEPLOYMENT", "")
  azure_openai_api_version    = get_env("AZURE_OPENAI_API_VERSION", "2023-12-01-preview")
}
