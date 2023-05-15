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
