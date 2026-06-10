terraform {
  source = "../../opentofu/dev"
}

inputs = {
  location            = "westus2"
  resource_prefix     = "barrel"
  container_image_api = "ghcr.io/dontspillthetea/barrel-api:latest"
  container_image_web = "ghcr.io/dontspillthetea/barrel-web:latest"
  review_token        = get_env("BARREL_REVIEW_TOKEN", "fallback-review-token-123")
  demo_password       = get_env("BARREL_DEMO_PASSWORD", "fallback-demo-password-123")
}
