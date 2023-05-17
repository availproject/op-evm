output "instances" {
  value = [
    for v in aws_instance.node : {
      id                           = v.id
      primary_network_interface_id = v.primary_network_interface_id
      p2p_port                     = v.tags.P2PPort
      node_type                    = v.tags.NodeType
    }
  ]
}
