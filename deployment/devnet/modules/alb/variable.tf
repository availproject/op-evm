variable "deployment_name" {
  description = "The unique name for this particular deployment"
  type        = string
}

variable "nodes" {
  description = "List of nodes info"
  type        = list(object({
    id = string
    node_type = string
    p2p_port = number
  }))
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
