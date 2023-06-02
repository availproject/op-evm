variable "github_token" {}

terraform {
  backend "s3" {
    bucket = "availsl-tf-states"
    key    = "state/avail-settlement/devnet01"
    region = "us-east-1"
  }
}

module "devnet" {
  source          = "../../devnet"
  deployment_name = "devnet01"
  region          = "us-east-1"
  avail_hostname  = "internal-rpc.testnetsl.avail.private"
  avail_peer      = {
    route53_zone_private_id = "Z00910032XDCLG3QODES"
    route_table_private_ids = [
      "rtb-0504970d4f1352288",
      "rtb-0d42b7d1091ec6647",
      "rtb-019b5975af56a1a51",
    ]
    vpc_id = "vpc-0129e5f061ed6e3b5"
  }
  github_token = var.github_token
}
