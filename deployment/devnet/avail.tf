# This file is created out of necessity to provide a working deployment, normally we will deploy avail separately
# TODO remove this file

resource "aws_instance" "avail" {
  ami                  = var.base_ami
  instance_type        = var.base_instance_type
  subnet_id            = module.networking.private_subnets_by_zone[local.zones[0]]
  user_data            = file("${path.module}/cloud-init-avail.sh")
  iam_instance_profile = module.security.iam_node_profile_id
  key_name             = aws_key_pair.devnet.key_name

  root_block_device {
    delete_on_termination = true
    volume_size           = 10
    volume_type           = "gp2"
  }

  tags = {
    Name        = "avl-${var.deployment_name}"
    Hostname    = "avl-${var.deployment_name}"
    NodeType    = "avail"
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}

resource "aws_eip" "avail" {
  instance   = aws_instance.avail.id
  vpc        = true
  depends_on = [
    module.networking.igw_id
  ]
}

resource "aws_security_group" "avail" {
  name        = "allow-avl-all-${var.deployment_name}"
  description = "Allow all rpc and ws traffic"
  vpc_id      = module.networking.vpc_id
}

resource "aws_security_group_rule" "allow_rpc_avail" {
  type              = "ingress"
  from_port         = 9933
  to_port           = 9933
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.avail.id
}

resource "aws_security_group_rule" "allow_ws_avail" {
  type              = "ingress"
  from_port         = 9944
  to_port           = 9944
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.avail.id
}

resource "aws_security_group_rule" "allow_outbound_all" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = -1
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.avail.id
}

resource "aws_network_interface_sg_attachment" "sg_avail_attachment_rpc" {
  security_group_id    = aws_security_group.avail.id
  network_interface_id = aws_instance.avail.primary_network_interface_id
}
