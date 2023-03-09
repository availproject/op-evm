# TODO add load balancers and make ec2 instances private
resource "aws_lb" "avail_settlement_nodes" {
  name               = "avail-settlement-lb-${var.deployment_name}"
  load_balancer_type = "network"
  internal           = false
  subnets            = [for subnet in aws_subnet.devnet_public : subnet.id]
}
#resource "aws_lb_target_group" "avail" {
#  name        = "avail-${var.deployment_name}"
#  protocol    = "TCP"
#  target_type = "instance"
#  vpc_id      = aws_vpc.devnet.id
#  port        = "9944"
#}
#resource "aws_lb_target_group" "bootnode" {
#  count       = length(aws_instance.bootnode)
#  name        = format("bootnode-%s-%02d", var.deployment_name, count.index + 1)
#  protocol    = "TCP"
#  target_type = "instance"
#  vpc_id      = aws_vpc.devnet.id
#  port        = element(aws_instance.bootnode, count.index).tags.Port
#}
#resource "aws_lb_target_group" "node" {
#  count       = length(aws_instance.node)
#  name        = format("node-%s-%02d", var.deployment_name, count.index + 1)
#  protocol    = "TCP"
#  target_type = "instance"
#  vpc_id      = aws_vpc.devnet.id
#  port        = element(aws_instance.node, count.index).tags.Port
#}
#resource "aws_lb_target_group" "watchtower" {
#  count       = length(aws_instance.watchtower)
#  name        = format("watchtower-%s-%02d", var.deployment_name, count.index + 1)
#  protocol    = "TCP"
#  target_type = "instance"
#  vpc_id      = aws_vpc.devnet.id
#  port        = element(aws_instance.watchtower, count.index).tags.Port
#}
#resource "aws_lb_target_group_attachment" "avail" {
#  target_group_arn = aws_lb_target_group.avail.arn
#  target_id        = aws_instance.avail.id
#  port             = "9944"
#}
#resource "aws_lb_target_group_attachment" "bootnode" {
#  count            = length(aws_instance.bootnode)
#  target_group_arn = element(aws_lb_target_group.bootnode, count.index).arn
#  target_id        = element(aws_instance.bootnode, count.index).id
#  port             = element(aws_instance.bootnode, count.index).tags.Port
#}
#resource "aws_lb_target_group_attachment" "node" {
#  count            = length(aws_instance.node)
#  target_group_arn = element(aws_lb_target_group.node, count.index).arn
#  target_id        = element(aws_instance.node, count.index).id
#  port             = element(aws_instance.node, count.index).tags.Port
#}
#resource "aws_lb_target_group_attachment" "watchtower" {
#  count            = length(aws_instance.watchtower)
#  target_group_arn = element(aws_lb_target_group.watchtower, count.index).arn
#  target_id        = element(aws_instance.watchtower, count.index).id
#  port             = element(aws_instance.watchtower, count.index).tags.Port
#}
#resource "aws_lb_listener" "avail" {
#  load_balancer_arn = aws_lb.avail_settlement_nodes.arn
#  port              = "9944"
#  protocol          = "TCP"
#
#  default_action {
#    type             = "forward"
#    target_group_arn = aws_lb_target_group.avail.arn
#  }
#}
#resource "aws_lb_listener" "bootnode" {
#  count             = length(aws_instance.bootnode)
#  load_balancer_arn = aws_lb.avail_settlement_nodes.arn
#  port              = element(aws_instance.bootnode, count.index).tags.Port
#  protocol          = "TCP"
#
#  default_action {
#    type             = "forward"
#    target_group_arn = element(aws_lb_target_group.bootnode, count.index).arn
#  }
#}
#
#resource "aws_lb_listener" "node" {
#  count             = length(aws_instance.node)
#  load_balancer_arn = aws_lb.avail_settlement_nodes.arn
#  port              = element(aws_instance.node, count.index).tags.Port
#  protocol          = "TCP"
#
#  default_action {
#    type             = "forward"
#    target_group_arn = element(aws_lb_target_group.node, count.index).arn
#  }
#}
#resource "aws_lb_listener" "watchtower" {
#  count             = length(aws_instance.watchtower)
#  load_balancer_arn = aws_lb.avail_settlement_nodes.arn
#  port              = element(aws_instance.watchtower, count.index).tags.Port
#  protocol          = "TCP"
#
#  default_action {
#    type             = "forward"
#    target_group_arn = element(aws_lb_target_group.watchtower, count.index).arn
#  }
#}
