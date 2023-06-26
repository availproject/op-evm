variable "github_token" {}

terraform {
  backend "s3" {
    bucket = "availsl-tf-states"
    key    = "state/op-evm/devnet01"
    region = "eu-central-1"
  }
}

module "devnet" {
  source          = "../../devnet"
  deployment_name = "devnet01"
  region          = "eu-central-1"
  base_ami        = "ami-0329d3839379bfd15"
  avail_hostname  = "internal-rpc.testnetsl.avail.private"
  release         = "v0.0.0-test5"
  avail_peer      = {
    route53_zone_private_id = "Z0203299HDIO94TVIKOE"
    route_table_private_ids = [
      "rtb-09ccd6e9861b75348",
      "rtb-04322cec80f3852fe",
      "rtb-093f19237360b723b",
    ]
    vpc_id = "vpc-0104686bd6c7cd394"
  }
  github_token = var.github_token
}
