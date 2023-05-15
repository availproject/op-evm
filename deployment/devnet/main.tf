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

provider "github" {
  token = var.github_token
}

data "github_release" "get_release" {
  repository  = var.github_repository
  owner       = var.github_owner
  retrieve_by = "tag"
  release_tag = var.release
}

locals {
  artifact_url                     = {for v in data.github_release.get_release.assets : v.name => v.url}
  github_token_ssm_parameter_path  = "/${var.deployment_name}/github_token"
  nodes_secrets_ssm_parameter_path = "/${var.deployment_name}/${var.nodes_secrets_ssm_parameter_id}"
}

module "lambda" {
  source = "./modules/lambda"

  deployment_name                  = var.deployment_name
  assm_artifact_url                = local.artifact_url[var.assm_artifact_name]
  genesis_bucket_prefix            = var.genesis_bucket_prefix
  github_token                     = var.github_token
  iam_role_arn                     = module.security.iam_role_lambda_arn
  nodes_secrets_ssm_parameter_path = local.nodes_secrets_ssm_parameter_path
  total_nodes                      = var.node_count + var.watchtower_count + 1
}

module "security" {
  source = "./modules/security"

  deployment_name                  = var.deployment_name
  genesis_init_lambda_name         = module.lambda.genesis_init_lambda_name
  s3_bucket_genesis_name           = module.lambda.s3_bucket_genesis_name
  github_token_ssm_parameter_path  = local.github_token_ssm_parameter_path
  nodes_secrets_ssm_parameter_path = local.nodes_secrets_ssm_parameter_path
}

module "alb" {
  source = "./modules/alb"

  deployment_name   = var.deployment_name
  public_subnets_id = [for subnet in aws_subnet.devnet_public : subnet.id]
  vpc_id            = aws_vpc.devnet.id
  nodes             = [
    for node in concat([aws_instance.bootnode], aws_instance.node, aws_instance.watchtower) :
    { id : node.id, p2p_port : node.tags.P2PPort, node_type : node.tags.NodeType }
  ]
  grpc_port         = var.grpc_port
  jsonrpc_port      = var.jsonrpc_port
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