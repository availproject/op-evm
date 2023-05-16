variable "deployment_name" {
  description = "The unique name for this particular deployment"
  type        = string
}

variable "zones" {
  description = "The zones for deployment"
  type        = list(string)
}

variable "devnet_public_subnet" {
  description = "The cidr block for the public subnet in our VPC"
  type        = list(string)
}

variable "devnet_private_subnet" {
  description = "The cidr block for the private subnet in our VPC"
  type        = list(string)
}

variable "devnet_vpc_block" {
  description = "The cidr block for our VPC"
  type        = string
}
