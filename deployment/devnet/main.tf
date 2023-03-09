terraform {
  cloud {
    organization = "avail"

    workspaces {
      name = "avail-settlement"
    }
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.19.0"
    }
  }

  required_version = ">= 1.3.9"
}

provider "aws" {
  region = "us-east-1"
  default_tags {
    tags = {
      Environment    = "devnet"
      Network        = "avail-settlement"
      DeploymentName = var.deployment_name
    }
  }
}

resource "tls_private_key" "pk" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "devnet" {
  key_name   = "${var.devnet_key_name}-${var.deployment_name}"
  public_key = tls_private_key.pk.public_key_openssh
}

data "aws_caller_identity" "provisioner" {}

locals {
  all_instances = concat([aws_instance.avail], aws_instance.node, aws_instance.bootnode, aws_instance.watchtower)
  all_nodes     = concat(aws_instance.node, aws_instance.bootnode, aws_instance.watchtower)
  ref_bootnode  = {
    for i, node in local.all_nodes : node.id => [for j, bootnode in aws_instance.bootnode : bootnode if node.id != bootnode.id]
  }
  ref_node  = {
    for i, node in local.all_nodes : node.id => [for j, node2 in aws_instance.node : node2 if node.id != node2.id]
  }
  ref = { //TODO rename
    for i, node in local.all_nodes : node.id => length(local.ref_bootnode[node.id]) > 0 ? local.ref_bootnode[node.id][0]: local.ref_node[node.id][0]
  }
}