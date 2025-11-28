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
  public_subnet_secondary_cidr  = var.public_subnet_secondary_cidr
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

# Route53 Module for Domain Management
module "route53" {
  source = "./modules/route53"

  project_name        = var.project_name
  environment         = var.environment
  domain_name         = var.domain_name
  enable_health_check = var.enable_route53_health_check
  health_check_path   = var.route53_health_check_path
  create_dns_record   = false # Set to true AFTER ALB is created by Helm deployment
  tags                = var.tags
}

# IMPORTANTE: Route53 DNS Record Configuration
# 1. Primeiro deployment: create_dns_record = false (apenas cria Hosted Zone)
# 2. Após deploy do Helm (ALB criado): create_dns_record = true
# 3. Copie os nameservers da Hosted Zone e configure no Registro.br
#    Outputs: module.route53.name_servers

# Secrets Manager Module for centralized secrets
module "secrets_manager" {
  source = "./modules/secrets-manager"

  project_name = var.project_name
  environment  = var.environment
  tags         = var.tags

  # RDS Auth credentials
  rds_auth_username = var.rds_auth_username
  rds_auth_password = var.rds_auth_password
  rds_auth_endpoint = module.rds_auth.db_address
  rds_auth_db_name  = var.rds_auth_db_name

  # RDS Orders credentials
  rds_orders_username = var.rds_orders_username
  rds_orders_password = var.rds_orders_password
  rds_orders_endpoint = module.rds_orders.db_address
  rds_orders_db_name  = var.rds_orders_db_name

  # RabbitMQ credentials
  rabbitmq_username = var.rabbitmq_admin_username
  rabbitmq_password = var.rabbitmq_admin_password
  rabbitmq_endpoint = replace(replace(module.amazonmq.amqp_ssl_endpoint, "amqps://", ""), ":5671", "")

  # JWT secrets
  jwt_secret         = var.jwt_secret
  jwt_refresh_secret = var.jwt_refresh_secret

  # MongoDB Atlas
  mongodb_connection_string = var.mongodb_connection_string

  # Redis variables removed - Redis runs in-cluster via Helm Chart
  # If migrating to ElastiCache, uncomment variables in modules/secrets-manager/variables.tf

  depends_on = [
    module.rds_auth,
    module.rds_orders,
    module.amazonmq
  ]
}

# Null Resource to cleanup Kubernetes-created AWS resources before destroy
# This prevents "DependencyViolation" errors when destroying VPC/subnets
resource "null_resource" "cleanup_k8s_resources" {
  # Trigger recreates if cluster name changes
  triggers = {
    cluster_name = module.eks.cluster_name
    region       = var.aws_region
  }

  # Destroy-time provisioner: runs BEFORE Terraform destroys resources
  provisioner "local-exec" {
    when    = destroy
    command = <<-EOT
      set -e

      echo "=========================================="
      echo "🧹 Cleaning up Kubernetes-created AWS resources..."
      echo "=========================================="

      # Update kubeconfig to ensure we can connect
      aws eks update-kubeconfig \
        --region ${self.triggers.region} \
        --name ${self.triggers.cluster_name} \
        --kubeconfig /tmp/cleanup-kubeconfig || true

      export KUBECONFIG=/tmp/cleanup-kubeconfig

      # Delete Ingresses (creates ALBs via AWS Load Balancer Controller)
      echo "Deleting Ingresses (ALBs)..."
      kubectl delete ingress --all -A --ignore-not-found=true --timeout=300s || true

      # Delete LoadBalancer Services (creates NLBs/CLBs)
      echo "Deleting LoadBalancer Services..."
      kubectl delete svc --all -A \
        --field-selector spec.type=LoadBalancer \
        --ignore-not-found=true \
        --timeout=300s || true

      # Delete PVCs (creates EBS volumes)
      echo "Deleting PersistentVolumeClaims (EBS volumes)..."
      kubectl delete pvc --all -A --ignore-not-found=true --timeout=300s || true

      # Wait for AWS to cleanup resources
      echo "Waiting 120 seconds for AWS to cleanup ENIs and dependencies..."
      sleep 120

      echo "✅ Kubernetes resource cleanup completed!"

      # Cleanup temp kubeconfig
      rm -f /tmp/cleanup-kubeconfig
    EOT

    # Environment variables for AWS CLI
    environment = {
      AWS_DEFAULT_REGION = self.triggers.region
    }
  }

  # Ensure this runs after EKS is created but before it's destroyed
  depends_on = [
    module.eks
  ]
}
