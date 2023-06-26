output "dns_name" {
  value       = aws_lb.op_evm_nodes.dns_name
  description = "The load balancer dns name"
}
