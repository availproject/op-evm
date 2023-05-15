# Lambda

resource "aws_iam_role" "lambda" {
  name               = "lambda-${var.deployment_name}"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

resource "aws_iam_role_policy_attachment" "lambda_ssm_node_secrets_read" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.ssm_node_secrets_read.arn
}

resource "aws_iam_role_policy_attachment" "lambda_s3_read" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.s3_read.arn
}

resource "aws_iam_role_policy_attachment" "lambda_s3_write" {
  role       = aws_iam_role.lambda.name
  policy_arn = aws_iam_policy.s3_write.arn
}

# EC2

resource "aws_iam_instance_profile" "node" {
  name = "instances_profile-${var.deployment_name}"
  role = aws_iam_role.ec2.name
}

resource "aws_iam_role" "ec2" {
  name_prefix = "ec2-${var.deployment_name}"

  assume_role_policy = data.aws_iam_policy_document.ec2_assume_role.json
}

resource "aws_iam_role_policy_attachment" "ec2_ssm_github_secrets_read" {
  role       = aws_iam_role.ec2.name
  policy_arn = aws_iam_policy.ssm_github_secrets_read.arn
}

resource "aws_iam_role_policy_attachment" "ssm_node_secrets_read" {
  role       = aws_iam_role.ec2.name
  policy_arn = aws_iam_policy.ssm_node_secrets_read.arn
}

resource "aws_iam_role_policy_attachment" "ssm_node_secrets_write" {
  role       = aws_iam_role.ec2.name
  policy_arn = aws_iam_policy.ssm_node_secrets_write.arn
}

resource "aws_iam_role_policy_attachment" "s3_read" {
  role       = aws_iam_role.ec2.name
  policy_arn = aws_iam_policy.s3_read.arn
}

resource "aws_iam_role_policy_attachment" "ec2_lambda_invoke" {
  role       = aws_iam_role.ec2.name
  policy_arn = aws_iam_policy.lambda_invoke.arn
}

resource "aws_iam_role_policy_attachment" "ec2_metrics" {
  role       = aws_iam_role.ec2.name
  policy_arn = aws_iam_policy.session_manager.arn
}
