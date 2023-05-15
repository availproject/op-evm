# Default Security Group of VPC
resource "aws_default_security_group" "default" {
  vpc_id     = aws_vpc.devnet.id
  depends_on = [
    aws_vpc.devnet
  ]

  ingress {
    from_port = "0"
    to_port   = "0"
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port = "0"
    to_port   = "0"
    protocol  = "-1"
    self      = true
  }
}

# Avail

resource "aws_security_group" "avail" {
  name        = "allow-avail-all-${var.deployment_name}"
  description = "Allow all rpc, ws traffic"
  vpc_id      = aws_vpc.devnet.id
  lifecycle {
    create_before_destroy = true
  }
}
resource "aws_security_group_rule" "allow_rpc_avail" {
  type              = "ingress"
  from_port         = "9933"
  to_port           = "9933"
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.avail.id
}
resource "aws_security_group_rule" "allow_ws_avail" {
  type              = "ingress"
  from_port         = "9944"
  to_port           = "9944"
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.avail.id
}

resource "aws_network_interface_sg_attachment" "sg_avail_attachment_rpc" {
  security_group_id    = aws_security_group.avail.id
  network_interface_id = aws_instance.avail.primary_network_interface_id
}

# Watchtower

resource "aws_security_group" "watchtower" {
  count       = length(aws_instance.watchtower)
  name        = format("allow-p2p-watchtower-%s-%02d", var.deployment_name, count.index + 1)
  description = "Allow all p2p, grpc, jsonrpc traffic"
  vpc_id      = aws_vpc.devnet.id
  lifecycle {
    create_before_destroy = true
  }
}
resource "aws_security_group_rule" "watchtower_grpc" {
  count             = length(aws_instance.watchtower)
  type              = "ingress"
  from_port         = element(aws_instance.watchtower, count.index).tags.GRPCPort
  to_port           = element(aws_instance.watchtower, count.index).tags.GRPCPort
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = element(aws_security_group.watchtower, count.index).id
}
resource "aws_security_group_rule" "watchtower_jsonrpc" {
  count             = length(aws_instance.watchtower)
  type              = "ingress"
  from_port         = element(aws_instance.watchtower, count.index).tags.JsonRPCPort
  to_port           = element(aws_instance.watchtower, count.index).tags.JsonRPCPort
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = element(aws_security_group.watchtower, count.index).id
}
resource "aws_security_group_rule" "watchtower_p2p" {
  count             = length(aws_instance.watchtower)
  type              = "ingress"
  from_port         = element(aws_instance.watchtower, count.index).tags.P2PPort
  to_port           = element(aws_instance.watchtower, count.index).tags.P2PPort
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = element(aws_security_group.watchtower, count.index).id
}

resource "aws_network_interface_sg_attachment" "sg_watchtower_attachment_p2p" {
  count                = length(aws_instance.watchtower)
  security_group_id    = element(aws_security_group.watchtower, count.index).id
  network_interface_id = element(aws_instance.watchtower, count.index).primary_network_interface_id
}

# Boot node

resource "aws_security_group" "bootnode" {
  name        = format("allow-p2p-bootnode-%s", var.deployment_name)
  description = "Allow all p2p, grpc, jsonrpc traffic"
  vpc_id      = aws_vpc.devnet.id
  lifecycle {
    create_before_destroy = true
  }
}
resource "aws_security_group_rule" "bootnode_grpc" {
  type              = "ingress"
  from_port         = aws_instance.bootnode.tags.GRPCPort
  to_port           = aws_instance.bootnode.tags.GRPCPort
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.bootnode.id
}
resource "aws_security_group_rule" "bootnode_jsonrpc" {
  type              = "ingress"
  from_port         = aws_instance.bootnode.tags.JsonRPCPort
  to_port           = aws_instance.bootnode.tags.JsonRPCPort
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.bootnode.id
}
resource "aws_security_group_rule" "bootnode_p2p" {
  type              = "ingress"
  from_port         = aws_instance.bootnode.tags.P2PPort
  to_port           = aws_instance.bootnode.tags.P2PPort
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.bootnode.id
}

resource "aws_network_interface_sg_attachment" "sg_bootnode_attachment_p2p" {
  security_group_id    = aws_security_group.bootnode.id
  network_interface_id = aws_instance.bootnode.primary_network_interface_id
}

# Node

resource "aws_security_group" "node" {
  count       = length(aws_instance.node)
  name        = format("allow-p2p-node-%s-%02d", var.deployment_name, count.index + 1)
  description = "Allow all p2p, gpc, jsonrpc traffic"
  vpc_id      = aws_vpc.devnet.id
  lifecycle {
    create_before_destroy = true
  }
}
resource "aws_security_group_rule" "node_grpc" {
  count             = length(aws_instance.node)
  type              = "ingress"
  from_port         = element(aws_instance.node, count.index).tags.GRPCPort
  to_port           = element(aws_instance.node, count.index).tags.GRPCPort
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = element(aws_security_group.node, count.index).id
}
resource "aws_security_group_rule" "node_jsonrpc" {
  count             = length(aws_instance.node)
  type              = "ingress"
  from_port         = element(aws_instance.node, count.index).tags.JsonRPCPort
  to_port           = element(aws_instance.node, count.index).tags.JsonRPCPort
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = element(aws_security_group.node, count.index).id
}
resource "aws_security_group_rule" "node_p2p" {
  count             = length(aws_instance.node)
  type              = "ingress"
  from_port         = element(aws_instance.node, count.index).tags.P2PPort
  to_port           = element(aws_instance.node, count.index).tags.P2PPort
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = element(aws_security_group.node, count.index).id
}

resource "aws_network_interface_sg_attachment" "node_attachment" {
  count                = length(aws_instance.node)
  security_group_id    = element(aws_security_group.node, count.index).id
  network_interface_id = element(aws_instance.node, count.index).primary_network_interface_id
}

# Outbound

resource "aws_security_group" "allow_outbound_everywhere" {
  name        = "allow-everything-out-${var.deployment_name}"
  description = "Allow all outgoing traffic"
  vpc_id      = aws_vpc.devnet.id
  lifecycle {
    create_before_destroy = true
  }
}
resource "aws_security_group_rule" "allow_outbound_everywhere" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = -1
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.allow_outbound_everywhere.id
}
resource "aws_network_interface_sg_attachment" "allow_outbound_everywhere" {
  count                = length(local.all_instances)
  security_group_id    = aws_security_group.allow_outbound_everywhere.id
  network_interface_id = element(local.all_instances, count.index).primary_network_interface_id
}
