data "aws_partition" "current" {}
data "aws_caller_identity" "provisioner" {}
data "aws_region" "current" {}

locals {
  user      = "ubuntu"
  workspace = "/home/${local.user}/workspace"
  node_tags = {
    Name    = "${var.node_type}-${var.deployment_name}"
    P2PPort = var.p2p_port
  }
}

resource "aws_launch_template" "node_lt" {
  name          = "${var.node_type}-node-launch-template"
  image_id      = var.base_ami
  instance_type = var.base_instance_type
  key_name      = var.key_name
  iam_instance_profile {
    name = var.iam_profile_name
  }
  user_data = base64gzip(data.cloudinit_config.cloud_init.rendered)

  block_device_mappings {
    device_name = "/dev/sdh"
    ebs {
      volume_size = 30
    }
  }

  vpc_security_group_ids = [
    aws_security_group.node_allow_inbound_sg.id,
    aws_security_group.node_allow_outbound_sg.id,
    aws_security_group.node_allow_inbound_p2p_sg.id
  ]
}

resource "aws_autoscaling_group" "node_asg" {
  name                      = "${var.node_type}-node-asg"
  min_size                  = var.node_count
  max_size                  = var.node_count
  desired_capacity          = var.node_count
  vpc_zone_identifier       = toset(values(var.subnets_by_zone))
  termination_policies      = ["OldestInstance"]
  wait_for_capacity_timeout = "10m"
  launch_template {
    id      = aws_launch_template.node_lt.id
    version = "$Latest"
  }


  tag {
    key                 = "NodeType"
    value               = var.node_type
    propagate_at_launch = true
  }

  tag {
    key                 = "Provisioner"
    value               = data.aws_caller_identity.provisioner.account_id
    propagate_at_launch = true
  }

  lifecycle {
    create_before_destroy = true
  }
}
