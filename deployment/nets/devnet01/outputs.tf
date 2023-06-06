output "dns_name" {
  value = module.devnet.dns_name
}

output "ssh_pk" {
  value     = module.devnet.ssh_pk
  sensitive = true
}
