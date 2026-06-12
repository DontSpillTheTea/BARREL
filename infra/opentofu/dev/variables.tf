variable "resource_group_name" {
  type    = string
  default = "barrel-ai-rg"
}

variable "resource_prefix" {
  type    = string
  default = "barrel"
}

variable "env_suffix" {
  type    = string
  default = "dev"
}

variable "openai_account_name" {
  type    = string
  default = "barrel-openai-sweden"
}

variable "openai_deployment_name" {
  type    = string
  default = "barrel-ai-native-parser"
}

variable "azure_openai_api_version" {
  type    = string
  default = "2025-01-01-preview"
}

variable "container_image_api" {
  type    = string
  default = "mcr.microsoft.com/azuredocs/containerapps-helloworld:latest"
}

variable "container_image_web" {
  type    = string
  default = "mcr.microsoft.com/azuredocs/containerapps-helloworld:latest"
}

variable "vision_endpoint" {
  type    = string
  default = ""
}

variable "vision_key" {
  type      = string
  sensitive = true
  default   = ""
}

variable "demo_username" {
  type    = string
  default = "evaluator"
}

variable "demo_password" {
  type      = string
  sensitive = true
}

variable "review_token" {
  type      = string
  sensitive = true
}
