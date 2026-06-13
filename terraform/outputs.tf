output "api_gateway_url" {
  value       = aws_apigatewayv2_stage.default_stage.invoke_url
  description = "Pega esta URL en el api_service.dart de Flutter"
}
output "sns_topic_arn" {
  value       = aws_sns_topic.user_notifications.arn
  description = "ARN del Tópico SNS para notificaciones"
}