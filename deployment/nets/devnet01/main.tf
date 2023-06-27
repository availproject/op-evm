variable "github_token" {}

terraform {
  backend "s3" {
    bucket = "op-evm-tf-states"
    key    = "state/op-evm/devnet01"
    region = "eu-central-1"
  }
}

module "devnet" {
  source          = "../../devnet"
  deployment_name = "devnet01"
  region          = "eu-central-1"
  base_ami        = "ami-0329d3839379bfd15"
  avail_hostname  = "internal-rpc.testnetopevm.avail.private"
  release         = "v0.1.0"
  avail_peer      = {
    route53_zone_private_id = "Z0862640EX4LNUFXIE04"
    route_table_private_ids = [
      "rtb-08cf5bd8611e433b2",
      "rtb-0fea1adae5499b243",
      "rtb-0cfce2eb657f6eb28",
    ]
    vpc_id = "vpc-03a641408742d0e21"
  }
  github_token = var.github_token
}
