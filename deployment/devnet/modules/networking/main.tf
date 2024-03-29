data "aws_caller_identity" "provisioner" {}

resource "aws_vpc" "devnet" {
  cidr_block           = var.devnet_vpc_block
  instance_tenancy     = "default"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "devnet-${var.deployment_name}"
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.devnet.id

  tags = {
    Name        = "igw-${var.deployment_name}"
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}

resource "aws_eip" "nat_eip" {
  vpc        = true
  depends_on = [aws_internet_gateway.igw]
}

resource "aws_nat_gateway" "nat" {
  subnet_id     = aws_subnet.devnet_public[0].id
  allocation_id = aws_eip.nat_eip.id
}

resource "aws_subnet" "devnet_public" {
  count = length(var.zones)

  vpc_id                  = aws_vpc.devnet.id
  availability_zone       = var.zones[count.index]
  cidr_block              = var.devnet_public_subnet[count.index]
  map_public_ip_on_launch = false

  depends_on = [aws_internet_gateway.igw]

  tags = {
    Name        = "public-subnet-${var.zones[count.index]}-${var.deployment_name}"
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}

resource "aws_route_table" "devnet_public" {
  vpc_id = aws_vpc.devnet.id
}

resource "aws_route" "public_internet_gateway" {
  route_table_id         = aws_route_table.devnet_public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.igw.id
}

resource "aws_route_table_association" "public" {
  count = length(var.zones)

  subnet_id      = aws_subnet.devnet_public[count.index].id
  route_table_id = aws_route_table.devnet_public.id
}

resource "aws_subnet" "devnet_private" {
  count = length(var.zones)

  vpc_id            = aws_vpc.devnet.id
  availability_zone = var.zones[count.index]
  cidr_block        = var.devnet_private_subnet[count.index]
  tags              = {
    Name        = "private-subnet-${var.zones[count.index]}-${var.deployment_name}"
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}

resource "aws_route_table" "devnet_private" {
  vpc_id = aws_vpc.devnet.id
}

resource "aws_route" "private_nat_gateway" {
  route_table_id         = aws_route_table.devnet_private.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_nat_gateway.nat.id
}

resource "aws_route_table_association" "private" {
  count = length(var.zones)

  subnet_id      = aws_subnet.devnet_private[count.index].id
  route_table_id = aws_route_table.devnet_private.id
}

resource "aws_route53_zone" "private_zone" {
  name          = "${var.deployment_name}.op-evm.private"
  force_destroy = true
  vpc {
    vpc_id = aws_vpc.devnet.id
  }
  lifecycle {
    ignore_changes = [vpc]
  }
}
