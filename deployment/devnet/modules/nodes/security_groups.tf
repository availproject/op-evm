resource "aws_security_group" "node_allow_inbound_sg" {
  name        = "allow-inbound-${var.deployment_name}-${var.node_type}"
  description = "Allow grpc and jsonrpc traffic"
  vpc_id      = var.vpc_id
}

resource "aws_security_group_rule" "node_allow_inbound_grpc_sgr" {
  type              = "ingress"
  from_port         = var.grpc_port
  to_port           = var.grpc_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.node_allow_inbound_sg.id
}

resource "aws_security_group_rule" "node_allow_inbound_jsonrpc_sgr" {
  type              = "ingress"
  from_port         = var.jsonrpc_port
  to_port           = var.jsonrpc_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.node_allow_inbound_sg.id
}

resource "aws_security_group" "node_allow_inbound_p2p_sg" {
  name        = "allow-inbound-p2p-${var.deployment_name}-${var.node_type}"
  description = "Allow p2p traffic"
  vpc_id      = var.vpc_id
}

resource "aws_security_group_rule" "node_allow_inbound_p2p_sgr" {
  type              = "ingress"
  from_port         = var.p2p_port
  to_port           = var.p2p_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.node_allow_inbound_p2p_sg.id
}

# Outbound

resource "aws_security_group" "node_allow_outbound_sg" {
  name        = "allow-everything-out-${var.deployment_name}-${var.node_type}"
  description = "Allow all outgoing traffic"
  vpc_id      = var.vpc_id
}

resource "aws_security_group_rule" "node_allow_outbound_all_sgr" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = -1
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.node_allow_outbound_sg.id
}
