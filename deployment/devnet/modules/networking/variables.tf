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

variable "avail_peer" {
  description = "Avail peering configuration, for peering to work properly we need the peer VPC id, the Route53 zone and a list of peer route tables that we need to configure to point to out VPC"
  type = object({
    vpc_id = string
    route53_zone_private_id = string
    route_table_private_ids = list(string)
  })
}
