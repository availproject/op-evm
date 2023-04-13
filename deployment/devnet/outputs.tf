output "all_instances" {
  value = local.all_instances
}
output "ssh_pk" {
  value     = tls_private_key.pk.private_key_pem
  sensitive = true
}
