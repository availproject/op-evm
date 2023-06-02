resource "aws_vpc_peering_connection" "avail_peering" {
  vpc_id      = aws_vpc.devnet.id
  peer_vpc_id = var.avail_peer.vpc_id
  auto_accept = true

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  requester {
    allow_remote_vpc_dns_resolution = true
  }
}

data "aws_vpc_peering_connection" "avail_peering" {
  id = aws_vpc_peering_connection.avail_peering.id
}

resource "aws_route" "route_to_peer" {
  route_table_id            = aws_route_table.devnet_private.id
  destination_cidr_block    = data.aws_vpc_peering_connection.avail_peering.peer_cidr_block_set[0].cidr_block
  vpc_peering_connection_id = data.aws_vpc_peering_connection.avail_peering.id
}

resource "aws_route" "route_from_peer" {
  for_each                  = toset(var.avail_peer.route_table_private_ids)
  route_table_id            = each.key
  destination_cidr_block    = var.devnet_vpc_block
  vpc_peering_connection_id = data.aws_vpc_peering_connection.avail_peering.id
}

resource "aws_route53_zone_association" "r53z_to_peer" {
  zone_id = aws_route53_zone.private_zone.id
  vpc_id  = var.avail_peer.vpc_id
}

resource "aws_route53_zone_association" "r53z_from_peer" {
  zone_id = var.avail_peer.route53_zone_private_id
  vpc_id  = aws_vpc.devnet.id
}
