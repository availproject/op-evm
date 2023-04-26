output "all_instances" {
  value = local.all_instances
}

output "all_eips" {
  value = concat([aws_eip.avail], [aws_eip.bootnode], aws_eip.node, aws_eip.watchtower)
}

output "ssh_pk" {
  value     = tls_private_key.pk.private_key_pem
  sensitive = true
}
