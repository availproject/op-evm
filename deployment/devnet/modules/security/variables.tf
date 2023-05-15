variable "deployment_name" {
  description = "The unique name for this particular deployment"
  type        = string
}

variable "nodes_secrets_ssm_parameter_path" {
  description = "AWS System manager parameter path for creating the path to store the secrets"
  type        = string
}

variable "github_token_ssm_parameter_path" {
  description = "AWS System manager parameter path accessing the github token"
  type        = string
}

variable "s3_bucket_genesis_name" {
  description = "Genesis bucket name"
  type = string
}

variable "genesis_init_lambda_name" {
  description = "The name of the lambda function to initialize genesis.json"
  type = string
}
