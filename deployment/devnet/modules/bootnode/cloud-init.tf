data "cloudinit_config" "cloud_init" {
  gzip          = false
  base64_encode = false

  part {
    content_type = "text/x-shellscript"
    filename     = "01-mount-ebs.sh"
    content      = templatefile("${path.module}/templates/mount-ebs.sh", {
      workspace = local.workspace
    })
  }

  part {
    content_type = "text/x-shellscript"
    filename     = "02-cloud-init.sh"
    content      = templatefile("${path.module}/templates/cloud-init.sh", {
      workspace                       = local.workspace
      s3_bucket_name                  = var.s3_bucket_genesis_name
      avail_addr                      = var.avail_addr
      github_token_ssm_parameter_path = var.github_token_ssm_parameter_path
      user                            = local.user
      region                          = data.aws_region.current.name
      avail_settlement_artifact_url   = var.avail_settlement_artifact_url
      config_yaml_base64              = base64encode(templatefile("${path.module}/templates/config.yaml", {
        workspace    = local.workspace
        grpc_port    = var.grpc_port
        jsonrpc_port = var.jsonrpc_port
        p2p_port     = local.P2PPort
        public_dns   = var.lb_dns_name
      }))
      secrets_config_json_base64 = base64encode(templatefile("${path.module}/templates/secrets-config.json", {
        node_name                        = local.name
        region                           = data.aws_region.current.name
        nodes_secrets_ssm_parameter_path = var.nodes_secrets_ssm_parameter_path
      }))
      avail_settlement_service_base64 = base64encode(templatefile("${path.module}/templates/avail-settlement.service", {
        workspace  = local.workspace
        avail_addr = var.avail_addr
        user       = local.user
      }))
    })
  }
  depends_on = [
    aws_lambda_invocation.genesis_init
  ]
}
