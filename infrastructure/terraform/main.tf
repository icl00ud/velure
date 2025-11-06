provider "aws" {
  region = var.aws_region

  default_tags {
    tags = var.tags
  }
}

# Data source para obter account ID
data "aws_caller_identity" "current" {}

# Data source para obter partition (aws, aws-cn, aws-us-gov)
data "aws_partition" "current" {}

# VPC Module
module "vpc" {
  source = "./modules/vpc"

  project_name                  = var.project_name
  environment                   = var.environment
  vpc_cidr                      = var.vpc_cidr
  availability_zone             = var.availability_zone
  availability_zone_secondary   = var.availability_zone_secondary
  public_subnet_cidr            = var.public_subnet_cidr
  private_subnet_cidr           = var.private_subnet_cidr
  private_subnet_secondary_cidr = var.private_subnet_secondary_cidr
  tags                          = var.tags
}

# Security Groups Module
module "security_groups" {
  source = "./modules/security-groups"

  project_name = var.project_name
  environment  = var.environment
  vpc_id       = module.vpc.vpc_id
  vpc_cidr     = var.vpc_cidr
  tags         = var.tags
}

# EKS Module
module "eks" {
  source = "./modules/eks"

  project_name       = var.project_name
  environment        = var.environment
  cluster_version    = var.eks_cluster_version
  vpc_id             = module.vpc.vpc_id
  private_subnet_ids = module.vpc.private_subnet_ids
  public_subnet_ids  = [module.vpc.public_subnet_id]
  node_instance_type = var.node_instance_type
  node_desired_size  = var.node_desired_size
  node_min_size      = var.node_min_size
  node_max_size      = var.node_max_size
  node_disk_size     = var.node_disk_size
  tags               = var.tags

  depends_on = [
    module.vpc,
    module.security_groups
  ]
}

# RDS for Auth Service
module "rds_auth" {
  source = "./modules/rds"

  project_name      = var.project_name
  environment       = var.environment
  identifier        = "${var.project_name}-${var.environment}-auth"
  database_name     = var.rds_auth_db_name
  master_username   = var.rds_auth_username
  master_password   = var.rds_auth_password
  instance_class    = var.rds_instance_class
  allocated_storage = var.rds_allocated_storage
  engine_version    = var.rds_engine_version
  vpc_id            = module.vpc.vpc_id
  subnet_ids        = module.vpc.private_subnet_ids
  availability_zone = var.availability_zone
  security_group_id = module.security_groups.rds_sg_id
  tags              = merge(var.tags, { Service = "auth" })

  depends_on = [
    module.vpc,
    module.security_groups
  ]
}

# RDS for Orders Services (shared by publish-order and process-order)
module "rds_orders" {
  source = "./modules/rds"

  project_name      = var.project_name
  environment       = var.environment
  identifier        = "${var.project_name}-${var.environment}-orders"
  database_name     = var.rds_orders_db_name
  master_username   = var.rds_orders_username
  master_password   = var.rds_orders_password
  instance_class    = var.rds_instance_class
  allocated_storage = var.rds_allocated_storage
  engine_version    = var.rds_engine_version
  vpc_id            = module.vpc.vpc_id
  subnet_ids        = module.vpc.private_subnet_ids
  availability_zone = var.availability_zone
  security_group_id = module.security_groups.rds_sg_id
  tags              = merge(var.tags, { Service = "orders" })

  depends_on = [
    module.vpc,
    module.security_groups
  ]
}

# Amazon MQ (RabbitMQ) for Message Queue
module "amazonmq" {
  source = "./modules/amazonmq"

  project_name            = var.project_name
  environment             = var.environment
  host_instance_type      = var.amazonmq_instance_type
  deployment_mode         = var.amazonmq_deployment_mode
  rabbitmq_admin_username = var.rabbitmq_admin_username
  rabbitmq_admin_password = var.rabbitmq_admin_password
  private_subnet_ids      = module.vpc.private_subnet_ids
  security_group_id       = module.security_groups.amazonmq_sg_id
  tags                    = merge(var.tags, { Service = "messaging" })

  depends_on = [
    module.vpc,
    module.security_groups
  ]
}
