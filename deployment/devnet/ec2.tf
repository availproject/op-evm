resource "aws_instance" "avail" {
  ami                         = var.base_ami
  instance_type               = var.base_instance_type
  key_name                    = aws_key_pair.devnet.key_name
  subnet_id                   = element(aws_subnet.devnet_public, 0).id
  associate_public_ip_address = true

  root_block_device {
    delete_on_termination = true
    volume_size           = 30
    volume_type           = "gp2"
  }

  tags = {
    Name        = "avail-${var.deployment_name}"
    Hostname    = "avail-${var.deployment_name}"
    NodeType    = "avail"
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}

resource "aws_instance" "bootnode" {
  ami                         = var.base_ami
  instance_type               = var.base_instance_type
  key_name                    = aws_key_pair.devnet.key_name
  subnet_id                   = aws_subnet.devnet_public[0].id
  associate_public_ip_address = true
  root_block_device {
    delete_on_termination = true
    volume_size           = 30
    volume_type           = "gp2"
  }

  tags = {
    Name        = format("bootnode-%s", var.deployment_name)
    Hostname    = format("bootnode-%s", var.deployment_name)
    GRPCPort    = "30001"
    JsonRPCPort = "31001"
    P2PPort     = "32001"
    NodeType    = "bootstrap-sequencer"
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}

resource "aws_instance" "node" {
  ami                         = var.base_ami
  instance_type               = var.base_instance_type
  count                       = var.node_count
  key_name                    = aws_key_pair.devnet.key_name
  subnet_id                   = element(aws_subnet.devnet_public, count.index).id
  associate_public_ip_address = true

  root_block_device {
    delete_on_termination = true
    volume_size           = 30
    volume_type           = "gp2"
  }

  tags = {
    Name        = format("node-%s-%02d", var.deployment_name, count.index + 1)
    Hostname    = format("node-%s-%02d", var.deployment_name, count.index + 1)
    GRPCPort    = format("40%03d", count.index + 1)
    JsonRPCPort = format("41%03d", count.index + 1)
    P2PPort     = format("42%03d", count.index + 1)
    NodeType    = "sequencer"
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}

resource "aws_instance" "watchtower" {
  ami                         = var.base_ami
  instance_type               = var.base_instance_type
  count                       = var.watchtower_count
  key_name                    = aws_key_pair.devnet.key_name
  subnet_id                   = element(aws_subnet.devnet_public, count.index).id
  associate_public_ip_address = true

  root_block_device {
    delete_on_termination = true
    volume_size           = 30
    volume_type           = "gp2"
  }

  tags = {
    Name        = format("watchtower-%s-%02d", var.deployment_name, count.index + 1)
    Hostname    = format("watchtower-%s-%02d", var.deployment_name, count.index + 1)
    GRPCPort    = format("50%03d", count.index + 1)
    JsonRPCPort = format("51%03d", count.index + 1)
    P2PPort     = format("52%03d", count.index + 1)
    NodeType    = "watchtower"
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}
