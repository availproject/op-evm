resource "aws_lb" "avail_settlement_nodes" {
  name               = "avl-sl-lb-${var.deployment_name}"
  load_balancer_type = "network"
  internal           = false
  subnets            = [var.public_subnets_id[0]]
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

resource "aws_lb_listener" "p2p_listener" {
  load_balancer_arn = aws_lb.avail_settlement_nodes.arn
  port              = var.p2p_port
  protocol          = "TCP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.p2p_tg.arn
  }
}

resource "aws_lb_target_group" "p2p_tg" {
  name        = "p2p-tg-${var.deployment_name}"
  protocol    = "TCP"
  target_type = "instance"
  vpc_id      = var.vpc_id
  port        = var.p2p_port
}


resource "aws_autoscaling_attachment" "jsonrpc_tg_attach" {
  autoscaling_group_name = var.asg_name
  lb_target_group_arn = aws_lb_target_group.jsonrpc_tg.arn
}

resource "aws_autoscaling_attachment" "grpc_tg_attach" {
  autoscaling_group_name = var.asg_name
  lb_target_group_arn = aws_lb_target_group.grpc_tg.arn
}

resource "aws_lb_target_group_attachment" "bootnode_grpc_tg_attach" {
  target_group_arn = aws_lb_target_group.grpc_tg.arn
  target_id        = var.bootnode_instance_id
  port             = var.grpc_port
}

resource "aws_lb_target_group_attachment" "bootnode_jsonrpc_tg_attach" {
  target_group_arn = aws_lb_target_group.jsonrpc_tg.arn
  target_id        = var.bootnode_instance_id
  port             = var.jsonrpc_port
}

resource "aws_lb_target_group_attachment" "bootnode_p2p_tg_attach" {
  target_group_arn = aws_lb_target_group.p2p_tg.arn
  target_id        = var.bootnode_instance_id
  port             = var.p2p_port
}
