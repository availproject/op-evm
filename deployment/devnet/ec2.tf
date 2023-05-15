resource "aws_instance" "avail" {
  ami                         = var.base_ami
  instance_type               = var.base_instance_type
  key_name                    = aws_key_pair.devnet.key_name
  iam_instance_profile        = module.security.iam_node_profile_id
  subnet_id                   = aws_subnet.devnet_public[0].id
  availability_zone           = aws_subnet.devnet_public[0].availability_zone
  associate_public_ip_address = false
  user_data                   = file("${path.module}/ebs-mount.sh")

  root_block_device {
    delete_on_termination = true
    volume_size           = 10
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
  iam_instance_profile        = module.security.iam_node_profile_id
  subnet_id                   = aws_subnet.devnet_public[0].id
  availability_zone           = aws_subnet.devnet_public[0].availability_zone
  associate_public_ip_address = false
  user_data                   = file("${path.module}/ebs-mount.sh")

  root_block_device {
    delete_on_termination = true
    volume_size           = 10
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
  iam_instance_profile        = module.security.iam_node_profile_id
  subnet_id                   = element(aws_subnet.devnet_public, count.index).id
  availability_zone           = element(aws_subnet.devnet_public, count.index).availability_zone
  associate_public_ip_address = false
  user_data                   = file("${path.module}/ebs-mount.sh")

  root_block_device {
    delete_on_termination = true
    volume_size           = 10
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
  iam_instance_profile        = module.security.iam_node_profile_id
  subnet_id                   = element(aws_subnet.devnet_public, count.index).id
  availability_zone           = element(aws_subnet.devnet_public, count.index).availability_zone
  associate_public_ip_address = false
  user_data                   = file("${path.module}/ebs-mount.sh")

  root_block_device {
    delete_on_termination = true
    volume_size           = 10
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

resource "aws_ebs_volume" "avail" {
  availability_zone = aws_subnet.devnet_public[0].availability_zone
  size              = 30
}

resource "aws_ebs_volume" "bootnode" {
  availability_zone = aws_subnet.devnet_public[0].availability_zone
  size              = 30
}

resource "aws_ebs_volume" "node" {
  count             = var.node_count
  availability_zone = element(aws_subnet.devnet_public, count.index).availability_zone
  size              = 30
}

resource "aws_ebs_volume" "watchtower" {
  count             = var.watchtower_count
  availability_zone = element(aws_subnet.devnet_public, count.index).availability_zone
  size              = 30
}

resource "aws_volume_attachment" "avail" {
  device_name = "/dev/sdh"
  volume_id   = aws_ebs_volume.avail.id
  instance_id = aws_instance.avail.id
}

resource "aws_volume_attachment" "bootnode" {
  device_name = "/dev/sdh"
  volume_id   = aws_ebs_volume.bootnode.id
  instance_id = aws_instance.bootnode.id
}

resource "aws_volume_attachment" "node" {
  count       = var.node_count
  device_name = "/dev/sdh"
  volume_id   = element(aws_ebs_volume.node, count.index).id
  instance_id = element(aws_instance.node, count.index).id
}

resource "aws_volume_attachment" "watchtower" {
  count       = var.watchtower_count
  device_name = "/dev/sdh"
  volume_id   = element(aws_ebs_volume.watchtower, count.index).id
  instance_id = element(aws_instance.watchtower, count.index).id
}

