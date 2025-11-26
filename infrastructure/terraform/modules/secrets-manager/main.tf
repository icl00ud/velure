# AWS Secrets Manager for Velure secrets

# RDS Auth Database Secret
resource "aws_secretsmanager_secret" "rds_auth" {
  name        = "${var.project_name}/${var.environment}/rds-auth"
  description = "RDS credentials for auth service"

  tags = merge(
    var.tags,
    {
      Name    = "${var.project_name}-${var.environment}-rds-auth"
      Service = "auth"
    }
  )
}

resource "aws_secretsmanager_secret_version" "rds_auth" {
  secret_id = aws_secretsmanager_secret.rds_auth.id
  secret_string = jsonencode({
    username = var.rds_auth_username
    password = var.rds_auth_password
    host     = var.rds_auth_endpoint
    port     = 5432
    dbname   = var.rds_auth_db_name
    url      = "postgresql://${var.rds_auth_username}:${var.rds_auth_password}@${var.rds_auth_endpoint}:5432/${var.rds_auth_db_name}?sslmode=require"
  })
}

# RDS Orders Database Secret
resource "aws_secretsmanager_secret" "rds_orders" {
  name        = "${var.project_name}/${var.environment}/rds-orders"
  description = "RDS credentials for orders services"

  tags = merge(
    var.tags,
    {
      Name    = "${var.project_name}-${var.environment}-rds-orders"
      Service = "orders"
    }
  )
}

resource "aws_secretsmanager_secret_version" "rds_orders" {
  secret_id = aws_secretsmanager_secret.rds_orders.id
  secret_string = jsonencode({
    username = var.rds_orders_username
    password = var.rds_orders_password
    host     = var.rds_orders_endpoint
    port     = 5432
    dbname   = var.rds_orders_db_name
    url      = "postgresql://${var.rds_orders_username}:${var.rds_orders_password}@${var.rds_orders_endpoint}:5432/${var.rds_orders_db_name}?sslmode=require"
  })
}

# RabbitMQ Secret
resource "aws_secretsmanager_secret" "rabbitmq" {
  name        = "${var.project_name}/${var.environment}/rabbitmq"
  description = "RabbitMQ credentials for messaging"

  tags = merge(
    var.tags,
    {
      Name    = "${var.project_name}-${var.environment}-rabbitmq"
      Service = "messaging"
    }
  )
}

resource "aws_secretsmanager_secret_version" "rabbitmq" {
  secret_id = aws_secretsmanager_secret.rabbitmq.id
  secret_string = jsonencode({
    username = var.rabbitmq_username
    password = var.rabbitmq_password
    host     = var.rabbitmq_endpoint
    port     = 5671
    url      = "amqps://${var.rabbitmq_username}:${var.rabbitmq_password}@${var.rabbitmq_endpoint}:5671"
  })
}

# JWT Secrets
resource "aws_secretsmanager_secret" "jwt" {
  name        = "${var.project_name}/${var.environment}/jwt"
  description = "JWT secrets for authentication"

  tags = merge(
    var.tags,
    {
      Name    = "${var.project_name}-${var.environment}-jwt"
      Service = "auth"
    }
  )
}

resource "aws_secretsmanager_secret_version" "jwt" {
  secret_id = aws_secretsmanager_secret.jwt.id
  secret_string = jsonencode({
    secret        = var.jwt_secret
    refreshSecret = var.jwt_refresh_secret
  })
}

# MongoDB Atlas Secret
resource "aws_secretsmanager_secret" "mongodb" {
  name        = "${var.project_name}/${var.environment}/mongodb"
  description = "MongoDB Atlas connection string"

  tags = merge(
    var.tags,
    {
      Name    = "${var.project_name}-${var.environment}-mongodb"
      Service = "product"
    }
  )
}

resource "aws_secretsmanager_secret_version" "mongodb" {
  secret_id = aws_secretsmanager_secret.mongodb.id
  secret_string = jsonencode({
    url = var.mongodb_connection_string
  })
}

# Redis Secret (for future use or if using ElastiCache)
resource "aws_secretsmanager_secret" "redis" {
  name        = "${var.project_name}/${var.environment}/redis"
  description = "Redis connection details"

  tags = merge(
    var.tags,
    {
      Name    = "${var.project_name}-${var.environment}-redis"
      Service = "cache"
    }
  )
}

resource "aws_secretsmanager_secret_version" "redis" {
  secret_id = aws_secretsmanager_secret.redis.id
  secret_string = jsonencode({
    host     = var.redis_host
    port     = var.redis_port
    password = var.redis_password
  })
}
