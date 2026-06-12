terraform {
  source = "../../opentofu/dev"
}

inputs = {
  resource_group_name    = "barrel-ai-rg"
  resource_prefix        = "barrel"
  env_suffix             = "dev"
  openai_account_name    = "barrel-openai-sweden"
  openai_deployment_name = "barrel-ai-native-parser"

  azure_openai_api_version = get_env("AZURE_OPENAI_API_VERSION", "2025-01-01-preview")

  vision_endpoint = get_env("AZURE_VISION_ENDPOINT", "https://barrel-vision-dev-008ac.cognitiveservices.azure.com/")
  vision_key      = get_env("AZURE_VISION_KEY", "")

  demo_username = "evaluator"
  demo_password = get_env("BARREL_DEMO_PASSWORD", "fallback-demo-password-123")
  review_token  = get_env("BARREL_REVIEW_TOKEN", "fallback-review-token-123")
}
