output "all_instances" {
  value = local.all_instances
}

output "dns_name" {
  value       = module.alb.dns_name
  description = "The load balancer dns name"
}

output "avail_addr" {
  value       = aws_eip.avail.public_dns
  description = "Avail address"
}

output "ssh_pk" {
  value     = tls_private_key.pk.private_key_pem
  sensitive = true
}
