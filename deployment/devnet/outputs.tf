output "ssh_pk" {
  value     = tls_private_key.pk.private_key_pem
  sensitive = true
}

output "dns_name" {
  value = module.alb.dns_name
}
