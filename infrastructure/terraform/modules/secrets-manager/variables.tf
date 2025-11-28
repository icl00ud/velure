variable "project_name" {
  description = "Project name for resource naming"
  type        = string
}

variable "environment" {
  description = "Environment name"
  type        = string
}

variable "tags" {
  description = "Common tags for resources"
  type        = map(string)
  default     = {}
}

# RDS Auth variables
variable "rds_auth_username" {
  description = "RDS auth database username"
  type        = string
}

variable "rds_auth_password" {
  description = "RDS auth database password"
  type        = string
  sensitive   = true
}

variable "rds_auth_endpoint" {
  description = "RDS auth database endpoint"
  type        = string
}

variable "rds_auth_db_name" {
  description = "RDS auth database name"
  type        = string
}

# RDS Orders variables
variable "rds_orders_username" {
  description = "RDS orders database username"
  type        = string
}

variable "rds_orders_password" {
  description = "RDS orders database password"
  type        = string
  sensitive   = true
}

variable "rds_orders_endpoint" {
  description = "RDS orders database endpoint"
  type        = string
}

variable "rds_orders_db_name" {
  description = "RDS orders database name"
  type        = string
}

# RabbitMQ variables
variable "rabbitmq_username" {
  description = "RabbitMQ admin username"
  type        = string
}

variable "rabbitmq_password" {
  description = "RabbitMQ admin password"
  type        = string
  sensitive   = true
}

variable "rabbitmq_endpoint" {
  description = "RabbitMQ broker endpoint (without protocol)"
  type        = string
}

# JWT variables
variable "jwt_secret" {
  description = "JWT signing secret"
  type        = string
  sensitive   = true
}

variable "jwt_refresh_secret" {
  description = "JWT refresh token secret"
  type        = string
  sensitive   = true
}

# MongoDB variables
variable "mongodb_connection_string" {
  description = "MongoDB Atlas connection string"
  type        = string
  sensitive   = true
}

# Redis variables REMOVED - Redis runs in-cluster
# Only needed if using AWS ElastiCache (managed Redis)
# variable "redis_host" {
#   description = "Redis host"
#   type        = string
#   default     = ""
# }
# variable "redis_port" {
#   description = "Redis port"
#   type        = number
#   default     = 6379
# }
# variable "redis_password" {
#   description = "Redis password"
#   type        = string
#   sensitive   = true
#   default     = ""
# }
