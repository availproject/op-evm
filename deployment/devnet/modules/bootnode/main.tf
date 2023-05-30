terraform {
  required_providers {
    polygonedge = {
      source = "danielvladco/polygonedge"
    }
  }
}

data "aws_partition" "current" {}
data "aws_caller_identity" "provisioner" {}
data "aws_region" "current" {}

locals {
  user      = "ubuntu"
  workspace = "/home/${local.user}/workspace"
  Name    = "bootstrap-sequencer-1-${var.deployment_name}"
  P2PPort = var.p2p_port
}

resource "aws_instance" "node" {
  ami                         = var.base_ami
  instance_type               = var.base_instance_type
  key_name                    = var.key_name
  iam_instance_profile        = var.iam_profile_id
  subnet_id                   = var.subnets_by_zone[var.availability_zone]
  availability_zone           = var.availability_zone
  user_data_replace_on_change = true
  ebs_optimized               = true

  user_data_base64 = base64gzip(data.cloudinit_config.cloud_init.rendered)

  tags             = {
    Name        = local.Name
    NodeType    = "bootstrap-sequencer"
    Provisioner = data.aws_caller_identity.provisioner.account_id
    P2PPort     = local.P2PPort
  }

  depends_on = [
    aws_ssm_parameter.network_key,
    aws_ssm_parameter.validator_bls_key,
    aws_ssm_parameter.validator_key,
  ]
}

resource "aws_ebs_volume" "node_ebs" {
  availability_zone = var.availability_zone
  size              = 30
  tags              = {
    Name = "bootstrap-sequencer-1-${var.deployment_name}"
  }
}

resource "aws_volume_attachment" "node_ebs_attach" {
  device_name                    = "/dev/sdh"
  volume_id                      = aws_ebs_volume.node_ebs.id
  instance_id                    = aws_instance.node.id
  force_detach                   = true
  stop_instance_before_detaching = true
}

resource "polygonedge_secrets" "secrets" {}

resource "aws_ssm_parameter" "validator_key" {
  name  = "${var.nodes_secrets_ssm_parameter_path}/${local.Name}/validator-key"
  type  = "SecureString"
  value = polygonedge_secrets.secrets.validator_key_encoded
}

resource "aws_ssm_parameter" "validator_bls_key" {
  name  = "${var.nodes_secrets_ssm_parameter_path}/${local.Name}/validator-bls-key"
  type  = "SecureString"
  value = polygonedge_secrets.secrets.validator_bls_key_encoded
}

resource "aws_ssm_parameter" "network_key" {
  name  = "${var.nodes_secrets_ssm_parameter_path}/${local.Name}/network-key"
  type  = "SecureString"
  value = polygonedge_secrets.secrets.network_key_encoded
}

resource "aws_lambda_invocation" "genesis_init" {
  function_name = var.genesis_init_lambda_name
  input         = jsonencode({
    node_name = local.Name,
    node_port = tostring(local.P2PPort),
    node_dns  = var.lb_dns_name,
    node_type = "bootstrap-sequencer"
  })

  depends_on = [
    aws_ssm_parameter.network_key,
    aws_ssm_parameter.validator_bls_key,
    aws_ssm_parameter.validator_key,
  ]
}
