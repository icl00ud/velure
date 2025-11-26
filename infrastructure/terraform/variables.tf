variable "aws_region" {
  description = "AWS region for all resources"
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "velure"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "availability_zone" {
  description = "Single AZ to use for cost optimization"
  type        = string
  default     = "us-east-1a"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidr" {
  description = "CIDR block for primary public subnet"
  type        = string
  default     = "10.0.1.0/24"
}

variable "public_subnet_secondary_cidr" {
  description = "CIDR block for secondary public subnet (required for ALB multi-AZ)"
  type        = string
  default     = "10.0.11.0/24"
}

variable "private_subnet_cidr" {
  description = "CIDR block for private subnet"
  type        = string
  default     = "10.0.10.0/24"
}

variable "private_subnet_secondary_cidr" {
  description = "CIDR block for second private subnet (required for RDS subnet group)"
  type        = string
  default     = "10.0.12.0/24"
}

variable "availability_zone_secondary" {
  description = "Second AZ for multi-AZ requirements (ALB, RDS)"
  type        = string
  default     = "us-east-1b"
}

variable "eks_cluster_version" {
  description = "Kubernetes version for EKS cluster"
  type        = string
  default     = "1.29"
}

variable "node_instance_type" {
  description = "EC2 instance type for EKS nodes (t3.micro for cost optimization)"
  type        = string
  default     = "t3.small" # t3.micro não é suportado pelo EKS, t3.small é o mínimo
}

variable "node_desired_size" {
  description = "Desired number of nodes"
  type        = number
  default     = 2
}

variable "node_min_size" {
  description = "Minimum number of nodes"
  type        = number
  default     = 1
}

variable "node_max_size" {
  description = "Maximum number of nodes"
  type        = number
  default     = 2
}

variable "node_disk_size" {
  description = "Disk size in GB for EKS nodes"
  type        = number
  default     = 20
}

variable "rds_instance_class" {
  description = "RDS instance class (Free Tier: db.t4g.micro or db.t3.micro)"
  type        = string
  default     = "db.t4g.micro"
}

variable "rds_allocated_storage" {
  description = "RDS allocated storage in GB (Free Tier: up to 20GB)"
  type        = number
  default     = 20
}

variable "rds_engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "16.3"
}

variable "rds_auth_db_name" {
  description = "Database name for auth service"
  type        = string
  default     = "velure_auth"
}

variable "rds_auth_username" {
  description = "Master username for auth RDS instance"
  type        = string
  sensitive   = true
  default     = "velure_admin"
}

variable "rds_auth_password" {
  description = "Master password for auth RDS instance"
  type        = string
  sensitive   = true
}

variable "rds_orders_db_name" {
  description = "Database name for orders services (shared by publish and process)"
  type        = string
  default     = "velure_orders"
}

variable "rds_orders_username" {
  description = "Master username for orders RDS instance"
  type        = string
  sensitive   = true
  default     = "velure_admin"
}

variable "rds_orders_password" {
  description = "Master password for orders RDS instance"
  type        = string
  sensitive   = true
}

# Amazon MQ (RabbitMQ) Variables
variable "amazonmq_instance_type" {
  description = "Amazon MQ instance type (mq.t3.micro for cost optimization)"
  type        = string
  default     = "mq.t3.micro"
}

variable "amazonmq_deployment_mode" {
  description = "Amazon MQ deployment mode (SINGLE_INSTANCE or ACTIVE_STANDBY_MULTI_AZ)"
  type        = string
  default     = "SINGLE_INSTANCE"
}

variable "rabbitmq_admin_username" {
  description = "RabbitMQ admin username"
  type        = string
  sensitive   = true
  default     = "admin"
}

variable "rabbitmq_admin_password" {
  description = "RabbitMQ admin password"
  type        = string
  sensitive   = true
}

# Route53 Variables
variable "domain_name" {
  description = "Domain name for the application (e.g., velure.app.br)"
  type        = string
  default     = "velure.app.br"
}

variable "enable_route53_health_check" {
  description = "Enable Route53 health check for the domain"
  type        = bool
  default     = false
}

variable "route53_health_check_path" {
  description = "Path for Route53 health check endpoint"
  type        = string
  default     = "/"
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default = {
    Project     = "velure"
    ManagedBy   = "terraform"
    CostCenter  = "personal-project"
    Environment = "prod"
  }
}

# JWT Secrets
variable "jwt_secret" {
  description = "JWT signing secret for auth service"
  type        = string
  sensitive   = true
}

variable "jwt_refresh_secret" {
  description = "JWT refresh token secret for auth service"
  type        = string
  sensitive   = true
}

# MongoDB Atlas
variable "mongodb_connection_string" {
  description = "MongoDB Atlas connection string"
  type        = string
  sensitive   = true
}

# Redis (optional - for ElastiCache or in-cluster)
variable "redis_host" {
  description = "Redis host"
  type        = string
  default     = ""
}

variable "redis_port" {
  description = "Redis port"
  type        = number
  default     = 6379
}

variable "redis_password" {
  description = "Redis password"
  type        = string
  sensitive   = true
  default     = ""
}
