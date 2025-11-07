output "zone_id" {
  description = "Route53 hosted zone ID"
  value       = aws_route53_zone.main.zone_id
}

output "zone_arn" {
  description = "Route53 hosted zone ARN"
  value       = aws_route53_zone.main.arn
}

output "name_servers" {
  description = "Route53 hosted zone name servers - Configure these in your domain registrar"
  value       = aws_route53_zone.main.name_servers
}

output "domain_name" {
  description = "Domain name of the hosted zone"
  value       = aws_route53_zone.main.name
}

output "health_check_id" {
  description = "Route53 health check ID (if enabled)"
  value       = var.enable_health_check ? aws_route53_health_check.main[0].id : null
}
