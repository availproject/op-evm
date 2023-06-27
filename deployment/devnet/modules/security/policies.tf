resource "aws_iam_policy" "ssm_github_secrets_read" {
  name   = "ssm_github_secrets_read-${var.deployment_name}"
  policy = data.aws_iam_policy_document.ssm_github_secrets_read.json
}

resource "aws_iam_policy" "ssm_node_secrets_read" {
  name   = "ssm_node_secrets_read-${var.deployment_name}"
  policy = data.aws_iam_policy_document.ssm_node_secrets_read.json
}

resource "aws_iam_policy" "ssm_node_secrets_write" {
  name   = "ssm_node_secrets_write-${var.deployment_name}"
  policy = data.aws_iam_policy_document.ssm_node_secrets_write.json
}

resource "aws_iam_policy" "session_manager" {
  name        = "ec2_metrics_${var.deployment_name}"
  path        = "/"
  description = "Policy to provide permission to EC2"
  policy      = data.aws_iam_policy_document.session_manager.json
  tags = {
    Name        = "devnet_role"
    Provisioner = data.aws_caller_identity.provisioner.account_id
  }
}
