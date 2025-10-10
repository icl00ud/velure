output "eks_node_sg_id" {
  description = "Security group ID for EKS nodes"
  value       = aws_security_group.eks_node.id
}

output "rds_sg_id" {
  description = "Security group ID for RDS instances"
  value       = aws_security_group.rds.id
}

output "alb_sg_id" {
  description = "Security group ID for Application Load Balancer"
  value       = aws_security_group.alb.id
}
