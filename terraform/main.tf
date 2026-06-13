
resource "random_string" "suffix" {
  length  = 6
  special = false
  upper   = false  
}

# # 1. IAM Role para la Lambda
resource "aws_iam_role" "lambda_role" {
  name = "go-hexagonal-lambda-role-${random_string.suffix.result}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action    = "sts:AssumeRole"
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
    }]
  })
}

# Adjuntar permisos para escribir Logs en CloudWatch
resource "aws_iam_role_policy_attachment" "lambda_logs" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# # 2. La Función Lambda (Recibe el archivo zip compilado en Go)
resource "aws_lambda_function" "api_lambda" {
  filename         = "../bootstrap.zip" # Creado por GitHub Actions
  function_name    = "go-hexagonal-api-${random_string.suffix.result}" 
  role             = aws_iam_role.lambda_role.arn
  handler          = "bootstrap"        # Obligatorio para lambdas de Go en formato al2023
  runtime          = "provided.al2023"  # Entorno Linux optimizado de AWS
  timeout          = 15

  environment {
    variables = {
      DATABASE_URL = var.database_url
      JWT_SECRET   = var.jwt_secret
    }
  }
}

# Crear el grupo de Logs en CloudWatch de forma explícita
resource "aws_cloudwatch_log_group" "lambda_log_group" {
  name              = "/aws/lambda/${aws_lambda_function.api_lambda.function_name}"
  retention_in_days = 7
}

# # 3. API Gateway (Para exponer la Lambda al mundo/Flutter)
resource "aws_apigatewayv2_api" "http_api" {
  name          = "go-hexagonal-gateway-${random_string.suffix.result}" # 🔥 Nombre dinámico automático
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "lambda_integration" {
  api_id           = aws_apigatewayv2_api.http_api.id
  integration_type = "AWS_PROXY"
  integration_uri  = aws_lambda_function.api_lambda.arn
}

# Ruta comodín para que todas las peticiones (login, register, upload) vayan a la Lambda
resource "aws_apigatewayv2_route" "any_route" {
  api_id    = aws_apigatewayv2_api.http_api.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration.id}"
}

resource "aws_apigatewayv2_stage" "default_stage" {
  api_id      = aws_apigatewayv2_api.http_api.id
  name        = "$default"
  auto_deploy = true
}

# Permiso para que API Gateway pueda ejecutar la Lambda
resource "aws_lambda_permission" "api_gw_permission" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.api_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.http_api.execution_arn}/*/*"
}

# =========================================================================
# 1. CONFIGURACIÓN DE SNS (Simple Notification Service)
# =========================================================================
resource "aws_sns_topic" "user_notifications" {
  name = "user-notifications-topic-${random_string.suffix.result}"
}

# =========================================================================
# 2. CONFIGURACIÓN DE SQS (Simple Queue Service)
# =========================================================================
resource "aws_sqs_queue" "notification_queue" {
  name                      = "notification-processing-queue-${random_string.suffix.result}"
  delay_seconds             = 0
  max_message_size          = 262144
  message_retention_seconds = 86400 # 1 día de retención
  receive_wait_time_seconds = 10    # Long polling
}

# Política para permitir que SNS publique mensajes dentro de la cola SQS
resource "aws_sqs_queue_policy" "sns_to_sqs_policy" {
  queue_url = aws_sqs_queue.notification_queue.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect    = "Allow"
        Principal = "*"
        Action    = "sqs:SendMessage"
        Resource  = aws_sqs_queue.notification_queue.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.user_notifications.arn
          }
        }
      }
    ]
  })
}

# =========================================================================
# 3. SUSCRIPCIÓN SNS ➔ SQS
# =========================================================================
resource "aws_sns_topic_subscription" "sns_to_sqs" {
  topic_arn = aws_sns_topic.user_notifications.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.notification_queue.arn
}

# =========================================================================
# 4. NUEVA LAMBDA DE NOTIFICACIONES (notification-lambda)
# =========================================================================
resource "aws_lambda_function" "notification_lambda" {
  filename         = "../notification_bootstrap.zip" # Este ZIP lo creará GitHub Actions
  function_name    = "notification-lambda-${random_string.suffix.result}"
  role             = aws_iam_role.lambda_exec_role.arn # Reutilizamos el rol existente
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architectures    = ["x86_64"]
  timeout          = 30
}

# Grupo de logs en CloudWatch para la nueva Lambda
resource "aws_cloudwatch_log_group" "notification_log_group" {
  name              = "/aws/lambda/notification-lambda-${random_string.suffix.result}"
  retention_in_days = 7
}

# Permisos para que SQS pueda activar a la Lambda de Notificaciones
resource "aws_iam_role_policy_attachment" "lambda_sqs_execution" {
  role       = aws_iam_role.lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaSQSQueueExecutionRole"
}

# =========================================================================
# 5. EVENT SOURCE MAPPING (SQS ➔ Lambda)
# =========================================================================
resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn = aws_sqs_queue.notification_queue.arn
  function_name    = aws_lambda_function.notification_lambda.arn
  batch_size       = 10 # Procesa hasta 10 mensajes juntos
}

# Enviar el ARN de SNS como variable de entorno a la Lambda Principal 
# Busca tu recurso actual "aws_lambda_function" (la de tu backend de Go) 
# y asegúrate de agregar esta variable dentro del bloque environment:
# environment {
#   variables = {
#     DATABASE_URL = var.database_url
#     JWT_SECRET   = var.jwt_secret
#     SNS_TOPIC_ARN = aws_sns_topic.user_notifications.arn  <-- ¡ESTA LÍNEA!
#   }
# }