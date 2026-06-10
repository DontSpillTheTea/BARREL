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
