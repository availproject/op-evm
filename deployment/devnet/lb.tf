# TODO add load balancers and make ec2 instances private
resource "aws_lb" "avail_settlement_nodes" {
  name               = "avail-settlement-lb-${var.deployment_name}"
  load_balancer_type = "network"
  internal           = false
  subnets            = [for subnet in aws_subnet.devnet_public : subnet.id]
}
