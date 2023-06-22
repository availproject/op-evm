data "aws_partition" "current" {}
data "aws_caller_identity" "provisioner" {}
data "aws_region" "current" {}

locals {
  user      = "ubuntu"
  workspace = "/home/${local.user}/workspace"
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
    Name        = var.name
    Provisioner = data.aws_caller_identity.provisioner.account_id
    P2PPort     = var.p2p_port
  }
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
