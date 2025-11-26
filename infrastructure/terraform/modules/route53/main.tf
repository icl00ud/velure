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

# Data source para buscar o Load Balancer criado pelo AWS LB Controller
data "aws_lb" "ui_service" {
  count = var.create_dns_record ? 1 : 0

  tags = {
    "elbv2.k8s.aws/cluster" = "${var.project_name}-${var.environment}"
  }
}

# Record A ALIAS apontando para o Load Balancer do UI Service
resource "aws_route53_record" "main" {
  count   = var.create_dns_record ? 1 : 0
  zone_id = aws_route53_zone.main.zone_id
  name    = var.domain_name
  type    = "A"

  alias {
    name                   = data.aws_lb.ui_service[0].dns_name
    zone_id                = data.aws_lb.ui_service[0].zone_id
    evaluate_target_health = true
  }
}

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
