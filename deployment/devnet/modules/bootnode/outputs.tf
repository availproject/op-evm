output "instance" {
  value = {
    id                           = aws_instance.node.id
    primary_network_interface_id = aws_instance.node.primary_network_interface_id
    p2p_port                     = aws_instance.node.tags.P2PPort
    node_type                    = aws_instance.node.tags.NodeType
  }
}
