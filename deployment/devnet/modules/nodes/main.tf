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
  node_tags = [
    for i in range(var.node_count) : {
      Name    = "${var.node_type}-${i + 1}-${var.deployment_name}"
      P2PPort = format("%d%03d", var.p2p_port_prefix, i + 1)
    }
  ]
}

resource "aws_instance" "node" {
  count = var.node_count

  ami                         = var.base_ami
  instance_type               = var.base_instance_type
  key_name                    = var.key_name
  iam_instance_profile        = var.iam_profile_id
  subnet_id                   = var.subnets_by_zone[element(var.zones, count.index)]
  availability_zone           = element(var.zones, count.index)
  user_data_replace_on_change = true
  ebs_optimized               = true

  user_data_base64 = base64gzip(data.cloudinit_config.cloud_init[count.index].rendered)
  tags             = merge({
    NodeType    = var.node_type
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }, local.node_tags[count.index])
  depends_on = [
    aws_ssm_parameter.network_key,
    aws_ssm_parameter.validator_bls_key,
    aws_ssm_parameter.validator_key,
  ]
}

resource "aws_ebs_volume" "node_ebs" {
  count             = var.node_count
  availability_zone = element(var.zones, count.index)
  size              = 30
  tags              = {
    Name = "${var.node_type}-${count.index + 1}-${var.deployment_name}"
  }
}

resource "aws_volume_attachment" "node_ebs_attach" {
  count                          = length(aws_instance.node)
  device_name                    = "/dev/sdh"
  volume_id                      = aws_ebs_volume.node_ebs[count.index].id
  instance_id                    = aws_instance.node[count.index].id
  force_detach                   = true
  stop_instance_before_detaching = true
}

resource "polygonedge_secrets" "secrets" {
  count = var.node_count
}

resource "aws_ssm_parameter" "validator_key" {
  count = var.node_count
  name  = "${var.nodes_secrets_ssm_parameter_path}/${local.node_tags[count.index].Name}/validator-key"
  type  = "SecureString"
  value = polygonedge_secrets.secrets[count.index].validator_key_encoded
}

resource "aws_ssm_parameter" "validator_bls_key" {
  count = var.node_count
  name  = "${var.nodes_secrets_ssm_parameter_path}/${local.node_tags[count.index].Name}/validator-bls-key"
  type  = "SecureString"
  value = polygonedge_secrets.secrets[count.index].validator_bls_key_encoded
}

resource "aws_ssm_parameter" "network_key" {
  count = var.node_count
  name  = "${var.nodes_secrets_ssm_parameter_path}/${local.node_tags[count.index].Name}/network-key"
  type  = "SecureString"
  value = polygonedge_secrets.secrets[count.index].network_key_encoded
}

resource "aws_lambda_invocation" "genesis_init" {
  count = var.node_count

  function_name = var.genesis_init_lambda_name
  input         = jsonencode({
    node_name = local.node_tags[count.index].Name,
    node_port = local.node_tags[count.index].P2PPort,
    node_dns  = var.lb_dns_name,
    node_type = var.node_type
  })

  depends_on = [
    aws_ssm_parameter.network_key,
    aws_ssm_parameter.validator_bls_key,
    aws_ssm_parameter.validator_key,
  ]
}
