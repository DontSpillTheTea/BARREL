variable "location" {
  type    = string
  default = "westus2"
}

variable "resource_prefix" {
  type    = string
  default = "barrel"
}

variable "container_image_api" {
  type    = string
  default = "ghcr.io/dontspillthetea/barrel-api:latest"
}

variable "container_image_web" {
  type    = string
  default = "ghcr.io/dontspillthetea/barrel-web:latest"
}

variable "review_token" {
  type      = string
  sensitive = true
}

variable "demo_username" {
  type    = string
  default = "evaluator"
}

variable "demo_password" {
  type      = string
  sensitive = true
}

variable "azure_vision_sku" {
  type    = string
  default = "F0"
}

variable "storage_container_name" {
  type    = string
  default = "barrel-review"
}

variable "ai_second_read_enabled" {
  type    = bool
  default = false
}

variable "ai_second_read_auto_on_fail" {
  type    = bool
  default = false
}

variable "azure_openai_endpoint" {
  type    = string
  default = ""
}

variable "azure_openai_api_key" {
  type      = string
  sensitive = true
  default   = ""
}

variable "azure_openai_deployment" {
  type    = string
  default = ""
}

variable "azure_openai_api_version" {
  type    = string
  default = ""
}
