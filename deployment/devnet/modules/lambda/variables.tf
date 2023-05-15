variable "deployment_name" {
  description = "The unique name for this particular deployment"
  type        = string
}

variable "iam_role_arn" {
  description = "IAM role name"
  type = string
}

variable "nodes_secrets_ssm_parameter_path" {
  description = "AWS System manager parameter path for creating the path to store the secrets"
  type        = string
}

variable "ssm_namespace" {
  description = "AWS System manager namespace for storing the secrets"
  type        = string
  default     = "admin"
}

variable "assm_artifact_url" {
  description = "The artifact url for `assm` binary"
  type        = string
}

variable "github_token" {
  description = "The github token needed for downloading the private artifacts"
  type        = string
  sensitive   = true
}

variable "genesis_bucket_prefix" {
  description = "The prefix for the bucket to store the genesis.json file"
  type        = string
}

variable "total_nodes" {
  description = "The number of nodes to wait for"
  type = number
}
