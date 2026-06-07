variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "database_url" {
  type      = string
  sensitive = true
}

variable "jwt_secret" {
  type      = string
  sensitive = true
}