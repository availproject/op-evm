resource "aws_iam_policy" "ssm_node_secrets_read" {
  name   = "ssm_node_secrets_read-${var.deployment_name}"
  policy = data.aws_iam_policy_document.ssm_node_secrets_read.json
}

resource "aws_iam_policy" "s3_read" {
  name   = "s3_read-${var.deployment_name}"
  policy = data.aws_iam_policy_document.s3_read.json
}

resource "aws_iam_policy" "s3_write" {
  name   = "s3_write-${var.deployment_name}"
  policy = data.aws_iam_policy_document.s3_write.json
}
