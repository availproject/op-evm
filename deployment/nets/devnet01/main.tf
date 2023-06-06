variable "github_token" {}

terraform {
  backend "s3" {
    bucket = "availsl-tf-states"
    key    = "state/avail-settlement/devnet01"
    region = "eu-central-1"
  }
}

module "devnet" {
  source          = "../../devnet"
  deployment_name = "devnet01"
  region          = "eu-central-1"
  base_ami        = "ami-0329d3839379bfd15"
  avail_hostname  = "internal-rpc.testnetsl.avail.private"
  release         = "v0.0.0-test3"
  avail_peer      = {
    route53_zone_private_id = "Z033343881FSG8RVC6NX"
    route_table_private_ids = [
      "rtb-0c8611e988d00b211",
      "rtb-01175cc25b265f508",
      "rtb-076838c55a357fe88",
    ]
    vpc_id = "vpc-0923ac6bb2ee99521"
  }
  github_token = var.github_token
}
