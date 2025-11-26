# Secret ARNs
output "rds_auth_secret_arn" {
  description = "ARN of the RDS auth secret"
  value       = aws_secretsmanager_secret.rds_auth.arn
}

output "rds_orders_secret_arn" {
  description = "ARN of the RDS orders secret"
  value       = aws_secretsmanager_secret.rds_orders.arn
}

output "rabbitmq_secret_arn" {
  description = "ARN of the RabbitMQ secret"
  value       = aws_secretsmanager_secret.rabbitmq.arn
}

output "jwt_secret_arn" {
  description = "ARN of the JWT secret"
  value       = aws_secretsmanager_secret.jwt.arn
}

output "mongodb_secret_arn" {
  description = "ARN of the MongoDB secret"
  value       = aws_secretsmanager_secret.mongodb.arn
}

output "redis_secret_arn" {
  description = "ARN of the Redis secret"
  value       = aws_secretsmanager_secret.redis.arn
}

# Secret Names (for External Secrets Operator)
output "rds_auth_secret_name" {
  description = "Name of the RDS auth secret"
  value       = aws_secretsmanager_secret.rds_auth.name
}

output "rds_orders_secret_name" {
  description = "Name of the RDS orders secret"
  value       = aws_secretsmanager_secret.rds_orders.name
}

output "rabbitmq_secret_name" {
  description = "Name of the RabbitMQ secret"
  value       = aws_secretsmanager_secret.rabbitmq.name
}

output "jwt_secret_name" {
  description = "Name of the JWT secret"
  value       = aws_secretsmanager_secret.jwt.name
}

output "mongodb_secret_name" {
  description = "Name of the MongoDB secret"
  value       = aws_secretsmanager_secret.mongodb.name
}

output "redis_secret_name" {
  description = "Name of the Redis secret"
  value       = aws_secretsmanager_secret.redis.name
}
