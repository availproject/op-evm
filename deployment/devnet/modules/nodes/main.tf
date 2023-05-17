data "aws_partition" "current" {}
data "aws_caller_identity" "provisioner" {}
data "aws_region" "current" {}

resource "aws_instance" "node" {
  count = var.node_count

  ami                         = var.base_ami
  instance_type               = var.base_instance_type
  key_name                    = var.key_name
  iam_instance_profile        = var.iam_profile_id
  subnet_id                   = var.subnets_by_zone[var.zones[count.index]]
  availability_zone           = var.zones[count.index]
  user_data_replace_on_change = true
  ebs_optimized               = true

  tags = {
    Name        = "${var.node_type}-${count.index + 1}-${var.deployment_name}"
    P2PPort     = format("%d%03d", var.p2p_port_prefix, count.index + 1)
    NodeType    = var.node_type
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}

resource "aws_ebs_volume" "node_ebs" {
  count             = var.node_count
  availability_zone = var.zones[count.index]
  size              = 30
}

resource "aws_volume_attachment" "node_ebs_attach" {
  count                          = length(aws_instance.node)
  device_name                    = "/dev/sdh"
  volume_id                      = aws_ebs_volume.node_ebs[count.index].id
  instance_id                    = aws_instance.node[count.index].id
  force_detach                   = true
  stop_instance_before_detaching = true
}
