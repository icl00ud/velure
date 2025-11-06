output "broker_id" {
  description = "Amazon MQ Broker ID"
  value       = aws_mq_broker.rabbitmq.id
}

output "broker_arn" {
  description = "Amazon MQ Broker ARN"
  value       = aws_mq_broker.rabbitmq.arn
}

output "amqp_endpoint" {
  description = "Amazon MQ AMQP endpoint (full amqps:// URL)"
  value       = aws_mq_broker.rabbitmq.instances[0].endpoints[0]
}

output "amqp_ssl_endpoint" {
  description = "Amazon MQ AMQP SSL endpoint (amqps://)"
  value       = aws_mq_broker.rabbitmq.instances[0].endpoints[0]
}

output "console_url" {
  description = "Amazon MQ Management Console URL"
  value       = aws_mq_broker.rabbitmq.instances[0].console_url
}

output "broker_name" {
  description = "Amazon MQ Broker name"
  value       = aws_mq_broker.rabbitmq.broker_name
}
