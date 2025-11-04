# VPC Outputs
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "vpc_cidr" {
  description = "VPC CIDR block"
  value       = module.vpc.vpc_cidr
}

output "public_subnet_id" {
  description = "Public subnet ID"
  value       = module.vpc.public_subnet_id
}

output "private_subnet_id" {
  description = "Private subnet ID"
  value       = module.vpc.private_subnet_id
}

output "nat_gateway_id" {
  description = "NAT Gateway ID"
  value       = module.vpc.nat_gateway_id
}

# EKS Outputs
output "eks_cluster_id" {
  description = "EKS cluster ID"
  value       = module.eks.cluster_id
}

output "eks_cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = module.eks.cluster_endpoint
}

output "eks_cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}

output "eks_cluster_certificate_authority_data" {
  description = "EKS cluster certificate authority data"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "eks_cluster_oidc_issuer_url" {
  description = "OIDC issuer URL for the EKS cluster"
  value       = module.eks.cluster_oidc_issuer_url
}

output "eks_oidc_provider_arn" {
  description = "ARN of the OIDC Provider for EKS"
  value       = module.eks.oidc_provider_arn
}

output "eks_node_group_id" {
  description = "EKS node group ID"
  value       = module.eks.node_group_id
}

output "eks_node_role_arn" {
  description = "ARN of the EKS node IAM role"
  value       = module.eks.node_role_arn
}

# RDS Outputs - Auth
output "rds_auth_endpoint" {
  description = "RDS endpoint for auth service"
  value       = module.rds_auth.db_endpoint
}

output "rds_auth_address" {
  description = "RDS address for auth service"
  value       = module.rds_auth.db_address
}

output "rds_auth_port" {
  description = "RDS port for auth service"
  value       = module.rds_auth.db_port
}

output "rds_auth_database_name" {
  description = "Database name for auth service"
  value       = module.rds_auth.db_name
}

output "rds_auth_connection_string" {
  description = "Connection string for auth service (without password)"
  value       = "postgresql://${var.rds_auth_username}@${module.rds_auth.db_address}:${module.rds_auth.db_port}/${module.rds_auth.db_name}"
  sensitive   = true
}

output "rds_auth_password" {
  description = "Password for auth RDS (for Kubernetes secrets)"
  value       = var.rds_auth_password
  sensitive   = true
}

# RDS Outputs - Orders
output "rds_orders_endpoint" {
  description = "RDS endpoint for orders services"
  value       = module.rds_orders.db_endpoint
}

output "rds_orders_address" {
  description = "RDS address for orders services"
  value       = module.rds_orders.db_address
}

output "rds_orders_port" {
  description = "RDS port for orders services"
  value       = module.rds_orders.db_port
}

output "rds_orders_database_name" {
  description = "Database name for orders services"
  value       = module.rds_orders.db_name
}

output "rds_orders_connection_string" {
  description = "Connection string for orders services (without password)"
  value       = "postgresql://${var.rds_orders_username}@${module.rds_orders.db_address}:${module.rds_orders.db_port}/${module.rds_orders.db_name}"
  sensitive   = true
}

output "rds_orders_password" {
  description = "Password for orders RDS (for Kubernetes secrets)"
  value       = var.rds_orders_password
  sensitive   = true
}

# Security Groups
output "eks_node_security_group_id" {
  description = "Security group ID for EKS nodes"
  value       = module.security_groups.eks_node_sg_id
}

output "rds_security_group_id" {
  description = "Security group ID for RDS instances"
  value       = module.security_groups.rds_sg_id
}

output "alb_security_group_id" {
  description = "Security group ID for Application Load Balancer"
  value       = module.security_groups.alb_sg_id
}

# ALB Controller IAM Role
output "alb_controller_role_arn" {
  description = "ARN of IAM role for AWS Load Balancer Controller"
  value       = module.eks.aws_load_balancer_controller_role_arn
}

output "ebs_csi_driver_role_arn" {
  description = "ARN of IAM role for EBS CSI Driver"
  value       = module.eks.ebs_csi_driver_role_arn
}

# Kubeconfig command
output "kubeconfig_command" {
  description = "Command to update kubeconfig for EKS cluster"
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${module.eks.cluster_name}"
}

# Quick setup commands
output "setup_commands" {
  description = "Commands to set up kubectl and verify cluster"
  value       = <<-EOT
    # Update kubeconfig
    aws eks update-kubeconfig --region ${var.aws_region} --name ${module.eks.cluster_name}
    
    # Verify cluster
    kubectl get nodes
    kubectl get pods -A
    
    # Install AWS Load Balancer Controller (after configuring OIDC)
    # See docs/eks-load-balancer-controller-setup.md
  EOT
}
