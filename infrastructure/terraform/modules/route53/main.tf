# Route53 Hosted Zone
resource "aws_route53_zone" "main" {
  name    = var.domain_name
  comment = "Managed by Terraform - ${var.project_name} ${var.environment}"

  tags = merge(
    var.tags,
    {
      Name        = "${var.project_name}-${var.environment}-zone"
      Domain      = var.domain_name
      Environment = var.environment
    }
  )
}

# Data source para buscar o Load Balancer do Nginx Ingress
# Commented out for destroy - LB was manually deleted
# data "aws_lb" "nginx_ingress" {
#   tags = {
#     "elbv2.k8s.aws/cluster" = "${var.project_name}-${var.environment}"
#   }
# }

# Record A ALIAS apontando para o Load Balancer do Nginx Ingress
# Commented out for destroy - depends on deleted LB
# resource "aws_route53_record" "main" {
#   zone_id = aws_route53_zone.main.zone_id
#   name    = var.domain_name
#   type    = "A"
#
#   alias {
#     name                   = data.aws_lb.nginx_ingress.dns_name
#     zone_id                = data.aws_lb.nginx_ingress.zone_id
#     evaluate_target_health = true
#   }
# }

# Health Check para o domínio principal (opcional mas recomendado)
resource "aws_route53_health_check" "main" {
  count             = var.enable_health_check ? 1 : 0
  fqdn              = var.domain_name
  port              = 443
  type              = "HTTPS"
  resource_path     = var.health_check_path
  failure_threshold = "3"
  request_interval  = "30"

  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-health-check"
    }
  )
}
