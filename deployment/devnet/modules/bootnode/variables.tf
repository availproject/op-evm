variable "availability_zone" {
  description = "The availability zone for this particular deployment"
  type        = string
}

variable "vpc_id" {
  description = "VPC id"
  type        = string
}

variable "deployment_name" {
  description = "The unique name for this particular deployment"
  type        = string
}

variable "grpc_port" {
  description = "GRPC port for the bootstrap sequencer and sequencer to listen on"
  type        = number
}

variable "jsonrpc_port" {
  description = "JSON RPC port for the bootstrap sequencer to listen on"
  type        = number
}

variable "p2p_port" {
  description = "P2P port for the bootstrap sequencer to listen on"
  type        = number
}

variable "nodes_secrets_ssm_parameter_path" {
  description = "AWS System manager parameter path for creating the path to store the secrets"
  type        = string
}

variable "github_token_ssm_parameter_path" {
  description = "AWS System manager parameter path accessing the github token"
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

variable "op_evm_artifact_url" {
  description = "The artifact url for `op-evm` binary"
  type        = string
}

variable "avail_addr" {
  description = "Avail address"
  type = string
}

variable "iam_profile_id" {
  description = "IAM profile id"
  type = string
}

variable  "lb_dns_name" {
  description = "Load balancer DNS name"
  type = string
}

variable "subnets_by_zone" {
  description = "A mapping of zone and it's corresponding subnet"
  type        = map(string)
}

variable "key_name" {
  description = "AWS ssh public key name"
  type = string
}

variable "name" {
  description = "The name of the bootstrap sequencer node"
  type        = string
}

variable "genesis_json" {
  description = "genesis.json configuration file contents"
  type = string
}
