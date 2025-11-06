variable "project_name" {
  description = "Project name"
  type        = string
}

variable "environment" {
  description = "Environment (e.g., production, staging)"
  type        = string
}

variable "host_instance_type" {
  description = "Instance type for Amazon MQ broker"
  type        = string
  default     = "mq.t3.micro"
}

variable "deployment_mode" {
  description = "Deployment mode (SINGLE_INSTANCE or ACTIVE_STANDBY_MULTI_AZ)"
  type        = string
  default     = "SINGLE_INSTANCE"
  validation {
    condition     = contains(["SINGLE_INSTANCE", "ACTIVE_STANDBY_MULTI_AZ"], var.deployment_mode)
    error_message = "deployment_mode must be either SINGLE_INSTANCE or ACTIVE_STANDBY_MULTI_AZ"
  }
}

variable "rabbitmq_admin_username" {
  description = "RabbitMQ admin username"
  type        = string
  sensitive   = true
}

variable "rabbitmq_admin_password" {
  description = "RabbitMQ admin password"
  type        = string
  sensitive   = true
}

variable "private_subnet_ids" {
  description = "Private subnet IDs for Amazon MQ broker"
  type        = list(string)
}

variable "security_group_id" {
  description = "Security group ID for Amazon MQ broker"
  type        = string
}

variable "tags" {
  description = "Tags to apply to resources"
  type        = map(string)
  default     = {}
}
