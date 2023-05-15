output "iam_role_lambda_arn" {
  value = aws_iam_role.lambda.arn
}

output "iam_node_profile_id" {
  value = aws_iam_instance_profile.node.id
}
