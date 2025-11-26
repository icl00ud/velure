variable "domain_name" {
  description = "Domain name for the hosted zone (e.g., velure.app)"
  type        = string
}

variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "environment" {
  description = "Environment (development, staging, production)"
  type        = string
}

variable "tags" {
  description = "Common tags to apply to all resources"
  type        = map(string)
  default     = {}
}

variable "enable_health_check" {
  description = "Enable Route53 health check for the domain"
  type        = bool
  default     = false
}

variable "health_check_path" {
  description = "Path for health check endpoint"
  type        = string
  default     = "/"
}

variable "create_dns_record" {
  description = "Create DNS A record pointing to the LoadBalancer (requires LB to exist)"
  type        = bool
  default     = false
}
