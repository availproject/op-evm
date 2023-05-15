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
    github = {
      source  = "integrations/github"
      version = "~> 5.0"
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

data "aws_region" "current" {}

data "github_release" "get_release" {
  repository  = var.github_repository
  owner       = var.github_owner
  retrieve_by = "tag"
  release_tag = var.release
}

resource "aws_ssm_parameter" "github_token" {
  name        = local.github_token_ssm_parameter_path
  description = "Github token (needed for downloading private artifacts from github)"
  type        = "SecureString"
  value       = var.github_token
}

locals {
  artifact_url                     = {for v in data.github_release.get_release.assets : v.name => v.url}
  github_token_ssm_parameter_path  = "/${var.deployment_name}/github_token"
  nodes_secrets_ssm_parameter_path = "/${var.deployment_name}/${var.nodes_secrets_ssm_parameter_id}"

  zones     = [for zone_name in var.zone_names : "${data.aws_region.current.name}${zone_name}"]
  all_nodes = flatten([for v in module.nodes : v.instances])
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

module "networking" {
  source = "./modules/networking"

  deployment_name       = var.deployment_name
  devnet_private_subnet = var.devnet_private_subnet
  devnet_public_subnet  = var.devnet_public_subnet
  devnet_vpc_block      = var.devnet_vpc_block
  zones                 = local.zones
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
  public_subnets_id = values(module.networking.public_subnets_by_zone)
  vpc_id            = module.networking.vpc_id
  nodes             = local.all_nodes
  grpc_port         = var.grpc_port
  jsonrpc_port      = var.jsonrpc_port
}

module "nodes" {
  source = "./modules/nodes"

  for_each = {
    "bootstrap-sequencer" = { node_count = 1, port_prefix = 31 }
    "sequencer"           = { node_count = var.node_count, port_prefix = 32 }
    "watchtower"          = { node_count = var.watchtower_count, port_prefix = 33 }
  }
  node_type                        = each.key
  node_count                       = each.value.node_count
  p2p_port_prefix                  = each.value.port_prefix
  deployment_name                  = var.deployment_name
  accounts_artifact_url            = local.artifact_url[var.accounts_artifact_name]
  avail_settlement_artifact_url    = local.artifact_url[var.avail_settlement_artifact_name]
  base_ami                         = var.base_ami
  base_instance_type               = var.base_instance_type
  github_token_ssm_parameter_path  = local.github_token_ssm_parameter_path
  grpc_port                        = var.grpc_port
  jsonrpc_port                     = var.jsonrpc_port
  nodes_secrets_ssm_parameter_path = local.nodes_secrets_ssm_parameter_path
  polygon_edge_artifact_url        = var.polygon_edge_artifact_url
  subnets_by_zone                  = module.networking.private_subnets_by_zone
  avail_addr                       = aws_eip.avail.public_dns
  s3_bucket_genesis_name           = module.lambda.s3_bucket_genesis_name
  genesis_init_lambda_name         = module.lambda.genesis_init_lambda_name
  iam_profile_id                   = module.security.iam_node_profile_id
  lb_dns_name                      = module.alb.dns_name
  zones                            = local.zones
  key_name                         = aws_key_pair.devnet.key_name
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
