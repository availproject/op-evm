resource "aws_lb" "avail_settlement_nodes" {
  name               = "avail-settlement-lb-${var.deployment_name}"
  load_balancer_type = "network"
  internal           = false
  subnets            = var.public_subnets_id
}

resource "aws_lb_listener" "p2p_listener" {
  count             = length(var.nodes)
  load_balancer_arn = aws_lb.avail_settlement_nodes.arn
  port              = var.nodes[count.index].p2p_port
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.p2p_tg[count.index].arn
  }
}

resource "aws_lb_target_group" "p2p_tg" {
  count       = length(var.nodes)
  name        = "p2p-tg-${count.index + 1}-${var.deployment_name}"
  protocol    = "TCP"
  target_type = "instance"
  vpc_id      = var.vpc_id
  port        = var.nodes[count.index].p2p_port
}

resource "aws_lb_target_group_attachment" "p2p_tg_attach" {
  count            = length(var.nodes)
  target_group_arn = aws_lb_target_group.p2p_tg[count.index].arn
  target_id        = var.nodes[count.index].id
  port             = var.nodes[count.index].p2p_port
}

locals {
  sequencers = {
    for index, node in var.nodes : index => node.id if node.node_type == "sequencer" || node.node_type == "bootstrap-sequencer"
  }
}

resource "aws_lb_listener" "grpc_listener" {
  load_balancer_arn = aws_lb.avail_settlement_nodes.arn
  port              = var.grpc_port
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.grpc_tg.arn
  }
}

resource "aws_lb_target_group" "grpc_tg" {
  name        = "grpc-tg-${var.deployment_name}"
  protocol    = "TCP"
  target_type = "instance"
  vpc_id      = var.vpc_id
  port        = var.grpc_port
}

resource "aws_lb_target_group_attachment" "grpc_tg_attach" {
  for_each         = local.sequencers
  target_group_arn = aws_lb_target_group.grpc_tg.arn
  target_id        = each.value
  port             = var.grpc_port
}

resource "aws_lb_listener" "jsonrpc_listener" {
  load_balancer_arn = aws_lb.avail_settlement_nodes.arn
  port              = var.jsonrpc_port
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.jsonrpc_tg.arn
  }
}

resource "aws_lb_target_group" "jsonrpc_tg" {
  name        = "jsonrpc-tg-${var.deployment_name}"
  protocol    = "TCP"
  target_type = "instance"
  vpc_id      = var.vpc_id
  port        = var.jsonrpc_port
}

resource "aws_lb_target_group_attachment" "jsonrpc_tg_attach" {
  for_each         = local.sequencers
  target_group_arn = aws_lb_target_group.jsonrpc_tg.arn
  target_id        = each.value
  port             = var.jsonrpc_port
}
