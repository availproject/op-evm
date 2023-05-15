output "dns_name" {
  value       = aws_lb.avail_settlement_nodes.dns_name
  description = "The load balancer dns name"
}
