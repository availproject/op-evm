variable "deployment_name" {
  description = "The unique name for this particular deployment"
  type        = string
}

variable "grpc_port" {
  description = "GRPC port for the bootstrap sequencer and sequencer to listen on"
  type        = number
}

variable "jsonrpc_port" {
  description = "JSON RPC port for the bootstrap sequencer and sequencer to listen on"
  type        = number
}

variable "s3_bucket_genesis_name" {
  description = "Genesis bucket name"
  type = string
}

variable "nodes_secrets_ssm_parameter_path" {
  description = "AWS System manager parameter path for creating the path to store the secrets"
  type        = string
}

variable "github_token_ssm_parameter_path" {
  description = "AWS System manager parameter path accessing the github token"
  type        = string
}

variable "node_count" {
  description = "The number of sequencer nodes that we're going to deploy"
  type        = number
}

variable "node_type" {
  description = "The node types, can be watchtower, sequencer or bootstrap-sequencer"
  type        = string
}

variable "base_instance_type" {
  description = "The type of instance that we're going to use"
  type        = string
}

variable "base_ami" {
  description = "Value of the base AMI that we're using"
  type        = string
}

variable "avail_settlement_artifact_url" {
  description = "The artifact url for `avail-settlement` binary"
  type        = string
}

variable "accounts_artifact_url" {
  description = "The artifact url for `accounts` binary"
  type        = string
}

variable "avail_addr" {
  description = "Avail address"
  type = string
}

variable "genesis_init_lambda_name" {
  description = "The name of the lambda function to initialize genesis.json"
  type = string
}

variable "iam_profile_id" {
  description = "IAM profile id"
  type = string
}

variable "lb_dns_name" {
  description = "Load balancer DNS name"
  type = string
}

variable "zones" {
  description = "The zones for deployment"
  type        = list(string)
}

variable "subnets_by_zone" {
  description = "A mapping of zone and it's corresponding subnet"
  type        = map(string)
}

variable "p2p_port_prefix" {
  description = "P2P port prefix, must be unique across all the instances that use the same load balancer"
  type = number
}

variable "key_name" {
  description = "AWS ssh public key name"
  type = string
}
