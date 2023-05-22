# Default Security Group of VPC
resource "aws_default_security_group" "default" {
  vpc_id = module.networking.vpc_id

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

resource "aws_security_group" "allow_inbound_sg" {
  name        = "allow-inbound-${var.deployment_name}"
  description = "Allow grpc and jsonrpc traffic"
  vpc_id      = module.networking.vpc_id
}

resource "aws_security_group_rule" "allow_inbound_grpc_sgr" {
  type              = "ingress"
  from_port         = var.grpc_port
  to_port           = var.grpc_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.allow_inbound_sg.id
}

resource "aws_security_group_rule" "allow_inbound_jsonrpc_sgr" {
  type              = "ingress"
  from_port         = var.jsonrpc_port
  to_port           = var.jsonrpc_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.allow_inbound_sg.id
}

resource "aws_security_group" "allow_inbound_p2p_sg" {
  count       = length(local.all_nodes)
  name        = "allow-inbound-p2p-${count.index + 1}-${var.deployment_name}"
  description = "Allow p2p traffic"
  vpc_id      = module.networking.vpc_id
}

resource "aws_security_group_rule" "allow_inbound_p2p_sgr" {
  count             = length(local.all_nodes)
  type              = "ingress"
  from_port         = local.all_nodes[count.index].p2p_port
  to_port           = local.all_nodes[count.index].p2p_port
  protocol          = "tcp"
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.allow_inbound_p2p_sg[count.index].id
}

# Outbound

resource "aws_security_group" "allow_outbound_sg" {
  name        = "allow-everything-out-${var.deployment_name}"
  description = "Allow all outgoing traffic"
  vpc_id      = module.networking.vpc_id
}

resource "aws_security_group_rule" "allow_outbound_all_sgr" {
  type              = "egress"
  from_port         = 0
  to_port           = 0
  protocol          = -1
  cidr_blocks       = ["0.0.0.0/0"]
  security_group_id = aws_security_group.allow_outbound_sg.id
}

# Attachments

resource "aws_network_interface_sg_attachment" "inbound_node_sg_attach" {
  count                = length(local.all_nodes)
  security_group_id    = aws_security_group.allow_inbound_sg.id
  network_interface_id = local.all_nodes[count.index].primary_network_interface_id
}

resource "aws_network_interface_sg_attachment" "outbound_node_sg_attach" {
  count                = length(local.all_nodes)
  security_group_id    = aws_security_group.allow_outbound_sg.id
  network_interface_id = local.all_nodes[count.index].primary_network_interface_id
}

resource "aws_network_interface_sg_attachment" "inbound_node_p2p_sg_attach" {
  count                = length(local.all_nodes)
  security_group_id    = aws_security_group.allow_inbound_p2p_sg[count.index].id
  network_interface_id = local.all_nodes[count.index].primary_network_interface_id
}
