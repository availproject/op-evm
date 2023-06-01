variable "deployment_name" {
  description = "The unique name for this particular deployment"
  type        = string
}

variable "asg_name" {
  description = "Name of the ASG"
  type        = string
}

variable "bootnode_instance_id" {
  description = "Instance ID of the bootnode"
  type        = string
}

variable "public_subnets_id" {
  description = "Public subnets id"
  type        = list(string)
}

variable "vpc_id" {
  description = "VPC id"
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

variable "p2p_port" {
  description = "P2P port for the bootstrap sequencer to listen on"
  type        = number
}
