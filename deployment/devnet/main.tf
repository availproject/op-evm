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
  all_instances = concat([aws_instance.avail], [aws_instance.bootnode], aws_instance.node, aws_instance.watchtower)
}