variable "grpc_port" {
  description = "GRPC port for the bootstrap sequencer and sequencer to listen on"
  type        = number
  default     = 20001
}

variable "jsonrpc_port" {
  description = "JSON RPC port for the bootstrap sequencer and sequencer to listen on"
  type        = number
  default     = 20002
}

variable "p2p_port" {
  description = "P2P port for the bootstrap sequencer and sequencer to listen on"
  type        = number
  default     = 20021
}

variable "nodes_secrets_ssm_parameter_id" {
  description = "AWS System manager parameter id for creating the path to store the secrets"
  type        = string
  default     = "nodes_secrets"
}

variable "github_owner" {
  description = "Github repository owner or organisation to download the artifacts from"
  type        = string
  default     = "availproject"
}

variable "github_repository" {
  description = "Github repository name to download the artifacts from"
  type        = string
  default     = "op-evm"
}

variable "release" {
  description = "The avail settlement release (will match a tag from the github.com/availproject/op-evm repository)"
  type        = string
}

variable "github_token" {
  description = "The github token needed for downloading the private artifacts"
  type        = string
  sensitive   = true
}

variable "avail_settlement_artifact_name" {
  description = "The artifact name for `op-evm` binary"
  type        = string
  default     = "op-evm-linux-arm64.zip"
}

variable "deployment_name" {
  description = "The unique name for this particular deployment"
  type        = string
}

variable "base_instance_type" {
  description = "The type of instance that we're going to use"
  type        = string
  default     = "t4g.micro"
}

# TODO use aws_ami instead of referencing a default existing ami
variable "base_ami" {
  description = "Value of the base AMI that we're using"
  type        = string
  default     = "ami-0f9bd9098aca2d42b" # Ubuntu 22.04 LTS
}

variable "devnet_key_name" {
  description = "The name that we want to use for the ssh key pair"
  type        = string
  default     = "2023-02-21-op-evm-devnet"
}

variable "region" {
  description = "The AWS region"
  type        = string
  default     = "us-east-1"
}

variable "zone_names" {
  description = "The zones for deployment"
  type        = list(string)
  default     = ["a", "b", "c"]
}

variable "devnet_vpc_block" {
  description = "The cidr block for our VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "devnet_public_subnet" {
  description = "The cidr block for the public subnet in our VPC"
  type        = list(string)
  default     = ["10.0.2.0/23", "10.0.4.0/23", "10.0.6.0/23"]
}

variable "devnet_private_subnet" {
  description = "The cidr block for the private subnet in our VPC"
  type        = list(string)
  default     = ["10.0.128.0/23", "10.0.130.0/23", "10.0.132.0/23"]
}

variable "node_count" {
  description = "The number of sequencer nodes that we're going to deploy"
  type        = number
  default     = 1
}

variable "avail_hostname" {
  description = "Avail hostname is usually a dns name. (if avail is not exposed publicly make sure to configure vpc peering)"
  type = string
}

variable "avail_ws_port" {
  description = "Avail port number"
  type = number
  default = 8546
}

variable "avail_peer" {
  description = "Avail peering configuration, for peering to work properly we need the peer VPC id, the Route53 zone and a list of peer route tables that we need to configure to point to out VPC. Peering is only supported in the same aws account and region."
  type = object({
    vpc_id = string
    route53_zone_private_id = string
    route_table_private_ids = list(string)
  })
  default = null
}

variable "watchtower_count" {
  description = "The number of watchtower nodes that we're going to deploy"
  type        = number
  default     = 1
}
