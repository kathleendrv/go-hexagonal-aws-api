# 🎲 Generador de sufijo aleatorio para evitar conflictos de nombres duplicados (409)
resource "random_string" "suffix" {
  length  = 6
  special = false
  upper   = false  # Todo en minúsculas para que sea compatible con nombres de AWS
}

# # 1. IAM Role para la Lambda
resource "aws_iam_role" "lambda_role" {
  # 🤖 El nombre se generará dinámicamente: ej. go-hexagonal-lambda-role-abc123
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
  function_name    = "go-hexagonal-api-${random_string.suffix.result}" # 🔥 Nombre dinámico automático
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
# 👁️ NOTA: Al añadir este bloque aquí, AWS ya tendrá el Log Group creado y verás los logs de inmediato
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