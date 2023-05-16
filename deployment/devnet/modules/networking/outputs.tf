output "vpc_id" {
  value = aws_vpc.devnet.id
}

output "private_subnets_by_zone" {
  value = {for v in aws_subnet.devnet_private : v.availability_zone => v.id}
}

output "public_subnets_by_zone" {
  value = {for v in aws_subnet.devnet_public : v.availability_zone => v.id}
}

output "igw_id" {
  value = aws_internet_gateway.igw.id
}
