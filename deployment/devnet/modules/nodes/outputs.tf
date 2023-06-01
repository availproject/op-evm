output "asg_name" {
  value       = aws_autoscaling_group.node_asg.name
  description = "Name of the auto-scaling group"
}
