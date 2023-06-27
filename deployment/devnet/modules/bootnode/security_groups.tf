resource "aws_security_group" "bootnode_allow_inbound" {
  name        = "bootnode-allow-inbound-${var.deployment_name}"
  description = "Allow GRPC and JSON-RPC traffic"
  vpc_id      = var.vpc_id
}

resource "aws_security_group_rule" "bootnode_allow_inbound_grpc" {
  type              = "ingress"
  from_port         = var.grpc_port
  to_port           = var.grpc_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.bootnode_allow_inbound.id
}

resource "aws_security_group_rule" "bootnode_allow_inbound_jsonrpc" {
  type              = "ingress"
  from_port         = var.jsonrpc_port
  to_port           = var.jsonrpc_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.bootnode_allow_inbound.id
}

resource "aws_security_group" "bootnode_allow_inbound_p2p" {
  name        = "bootnode-allow-inbound-p2p-${var.deployment_name}"
  description = "Allow P2P traffic"
  vpc_id      = var.vpc_id
}

resource "aws_security_group_rule" "bootnode_allow_inbound_p2p" {
  type              = "ingress"
  from_port         = var.p2p_port
  to_port           = var.p2p_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.bootnode_allow_inbound_p2p.id
}

# Outbound

resource "aws_security_group" "bootnode_allow_outbound" {
  name        = "bootnode_allow-everything-out-${var.deployment_name}"
  description = "Allow all outgoing traffic"
  vpc_id      = var.vpc_id
}

resource "aws_security_group_rule" "bootnode_allow_outbound_all" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = -1
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.bootnode_allow_outbound.id
}

# Attachments
resource "aws_network_interface_sg_attachment" "bootnode_inbound" {
  security_group_id    = aws_security_group.bootnode_allow_inbound.id
  network_interface_id = aws_instance.node.primary_network_interface_id
}

resource "aws_network_interface_sg_attachment" "bootnode_outbound" {
  security_group_id    = aws_security_group.bootnode_allow_outbound.id
  network_interface_id = aws_instance.node.primary_network_interface_id
}

resource "aws_network_interface_sg_attachment" "bootnode_inbound_p2p" {
  security_group_id    = aws_security_group.bootnode_allow_inbound_p2p.id
  network_interface_id = aws_instance.node.primary_network_interface_id
}
