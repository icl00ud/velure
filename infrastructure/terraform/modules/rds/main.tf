# DB Subnet Group (precisa de pelo menos 2 subnets em AZs diferentes)
# Como estamos usando apenas 1 AZ, vamos criar uma segunda subnet na mesma AZ
# Isso não é ideal para produção, mas economiza custos

resource "aws_db_subnet_group" "main" {
  name       = "${var.identifier}-subnet-group"
  subnet_ids = var.subnet_ids

  tags = merge(
    var.tags,
    {
      Name = "${var.identifier}-subnet-group"
    }
  )
}

# RDS Parameter Group
resource "aws_db_parameter_group" "main" {
  name   = "${var.identifier}-pg"
  family = "postgres17"

  # Otimizações para free tier / low resource
  # Valores em formato numérico conforme requisito do AWS RDS
  # shared_buffers e effective_cache_size: 8KB blocks
  # work_mem e maintenance_work_mem: kilobytes
  parameter {
    name         = "shared_buffers"
    value        = "16384" # 128MB = 16384 blocks de 8KB
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "max_connections"
    value        = "100"
    apply_method = "pending-reboot"
  }

  parameter {
    name         = "work_mem"
    value        = "4096" # 4MB = 4096 KB
    apply_method = "immediate"
  }

  parameter {
    name         = "maintenance_work_mem"
    value        = "65536" # 64MB = 65536 KB
    apply_method = "immediate"
  }

  parameter {
    name         = "effective_cache_size"
    value        = "65536" # 512MB = 65536 blocks de 8KB
    apply_method = "pending-reboot"
  }

  tags = var.tags
}

# RDS Instance
resource "aws_db_instance" "main" {
  identifier     = var.identifier
  engine         = "postgres"
  engine_version = var.engine_version
  instance_class = var.instance_class

  # Storage
  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.allocated_storage + 10 # Auto-scaling limitado
  storage_type          = "gp3"                      # gp3 é mais barato que gp2
  storage_encrypted     = true

  # Database
  db_name  = var.database_name
  username = var.master_username
  password = var.master_password
  port     = 5432

  # Networking
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [var.security_group_id]
  publicly_accessible    = false
  availability_zone      = var.availability_zone

  # Maintenance
  parameter_group_name       = aws_db_parameter_group.main.name
  apply_immediately          = true
  auto_minor_version_upgrade = true
  maintenance_window         = "sun:03:00-sun:04:00"

  # Backup
  backup_retention_period  = 1 # Free Tier limit
  backup_window            = "02:00-03:00"
  delete_automated_backups = true
  skip_final_snapshot      = true # CUIDADO: Em produção, sempre fazer snapshot final
  # final_snapshot_identifier = "${var.identifier}-final-snapshot-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"

  # Monitoring - CloudWatch logs disabled to avoid conflicts
  enabled_cloudwatch_logs_exports = []
  monitoring_interval             = 0 # Desabilitar enhanced monitoring para economizar
  # monitoring_role_arn           = aws_iam_role.rds_monitoring.arn

  # Performance Insights (desabilitado para economizar)
  performance_insights_enabled = false

  # Multi-AZ (desabilitado para economizar)
  multi_az = false

  # Deletion protection
  deletion_protection = false # Facilitar cleanup, mas em prod deve ser true

  tags = merge(
    var.tags,
    {
      Name = var.identifier
    }
  )

  lifecycle {
    ignore_changes = [
      password, # Ignorar mudanças de senha após criação inicial
    ]
  }
}

# CloudWatch Log Groups removed to avoid conflicts
