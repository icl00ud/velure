# Amazon MQ (RabbitMQ) Broker
resource "aws_mq_broker" "rabbitmq" {
  broker_name        = "${var.project_name}-${var.environment}-rabbitmq"
  engine_type        = "RabbitMQ"
  engine_version     = "3.13"
  host_instance_type = var.host_instance_type
  deployment_mode    = var.deployment_mode
  auto_minor_version_upgrade = true

  user {
    username = var.rabbitmq_admin_username
    password = var.rabbitmq_admin_password
  }

  subnet_ids          = var.deployment_mode == "SINGLE_INSTANCE" ? [var.private_subnet_ids[0]] : var.private_subnet_ids
  security_groups     = [var.security_group_id]
  publicly_accessible = false

  logs {
    general = true
  }

  tags = merge(
    var.tags,
    {
      Name = "${var.project_name}-${var.environment}-rabbitmq"
    }
  )
}
