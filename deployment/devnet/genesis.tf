resource "polygonedge_secrets" "secrets" {}

resource "aws_ssm_parameter" "validator_key" {
  name  = "${local.nodes_secrets_ssm_parameter_path}/${local.bootnode_name}/validator-key"
  type  = "SecureString"
  value = polygonedge_secrets.secrets.validator_key_encoded
}

resource "aws_ssm_parameter" "validator_bls_key" {
  name  = "${local.nodes_secrets_ssm_parameter_path}/${local.bootnode_name}/validator-bls-key"
  type  = "SecureString"
  value = polygonedge_secrets.secrets.validator_bls_key_encoded
}

resource "aws_ssm_parameter" "network_key" {
  name  = "${local.nodes_secrets_ssm_parameter_path}/${local.bootnode_name}/network-key"
  type  = "SecureString"
  value = polygonedge_secrets.secrets.network_key_encoded
}

locals {
  genesis_json = templatefile("${path.module}/templates/genesis.json", {
    validator_address = polygonedge_secrets.secrets.address
    node_dns          = module.alb.dns_name
    node_port         = var.p2p_port
    node_id           = polygonedge_secrets.secrets.node_id
  })
}
