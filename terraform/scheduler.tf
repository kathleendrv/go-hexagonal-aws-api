# 1. Crear el Rol de IAM para EventBridge
resource "aws_iam_role" "scheduler_role" {
  name = "utesa-eventbridge-sns-scheduler-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = { Service = "scheduler.amazonaws.com" }
        Action    = "sts:AssumeRole"
      }
    ]
  })
}

# 2. Política de IAM apuntando a tu SNS de tu main.tf (user_notifications)
resource "aws_iam_policy" "scheduler_sns_policy" {
  name        = "utesa-eventbridge-scheduler-sns-policy"
  description = "Permite a EventBridge Scheduler publicar en el topico SNS"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = "sns:Publish"
        Resource = aws_sns_topic.user_notifications.arn # ➔ Cambiado para coincidir con tu main.tf
      }
    ]
  })
}

# 3. Adjuntar la política al rol
resource "aws_iam_role_policy_attachment" "scheduler_attach" {
  role       = aws_iam_role.scheduler_role.name
  policy_arn = aws_iam_policy.scheduler_sns_policy.arn
}

# 4. Configurar el Amazon EventBridge Scheduler (Ejecución cada 5 minutos)
resource "aws_scheduler_schedule" "five_minute_schedule" {
  name        = "utesa-notification-cron-every-5-minutes"
  group_name  = "default"

  schedule_expression = "rate(5 minutes)"

  flexible_time_window {
    mode = "OFF"
  }

  target {
    arn      = aws_sns_topic.user_notifications.arn # ➔ Cambiado para coincidir con tu main.tf
    role_arn = aws_iam_role.scheduler_role.arn

    input = jsonencode({
      email   = "scheduler@utesa.edu",
      subject = "Ejecucion Automatica EventBridge",
      message = "Mensaje programado automaticamente cada 5 minutos por EventBridge Scheduler."
    })
  }
}