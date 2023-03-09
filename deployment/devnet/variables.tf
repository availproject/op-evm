variable "deployment_name" {
  description = "The unique name for this particular deployment"
  type        = string
  default     = "test1"
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
  default     = "2023-02-21-avail-settlement-devnet"
}

variable "zones" {
  description = "The zones for deployment"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

variable "devnet_vpc_block" {
  description = "The cidr block for our VPC"
  type        = string
  default     = "10.0.0.0/16"
}

#variable "devnet_private_subnet" {
#  description = "The cidr block for the private subnet in our VPC"
#  type        = list(string)
#  default     = ["10.0.128.0/23", "10.0.130.0/23", "10.0.132.0/23"]
#}

variable "devnet_public_subnet" {
  description = "The cidr block for the public subnet in our VPC"
  type        = list(string)
  default     = ["10.0.2.0/23", "10.0.4.0/23", "10.0.6.0/23"]
}

variable "bootnode_count" {
  description = "The number of bootstrap sequencer nodes that we're going to deploy"
  type        = number
  default     = 1
}

variable "node_count" {
  description = "The number of sequencer nodes that we're going to deploy"
  type        = number
  default     = 1
}

variable "watchtower_count" {
  description = "The number of watchtower nodes that we're going to deploy"
  type        = number
  default     = 1
}