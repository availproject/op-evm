data "aws_caller_identity" "provisioner" {}
data "aws_region" "current" {}

data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "ssm_node_secrets_read" {
  version = "2012-10-17"
  statement {
    actions = [
      "ssm:GetParameter",
      "ssm:GetParameters",
      "ssm:GetParametersByPath",
    ]
    resources = [
      "arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.provisioner.account_id}:parameter${var.nodes_secrets_ssm_parameter_path}/*"
    ]
  }
}

data "aws_iam_policy_document" "ssm_node_secrets_write" {
  version = "2012-10-17"
  statement {
    actions = [
      "ssm:PutParameter",
      "ssm:DeleteParameter",
    ]
    resources = [
      "arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.provisioner.account_id}:parameter${var.nodes_secrets_ssm_parameter_path}/*"
    ]
  }
}

data "aws_iam_policy_document" "ssm_github_secrets_read" {
  version = "2012-10-17"
  statement {
    actions = [
      "ssm:GetParameter",
      "ssm:GetParameters",
      "ssm:GetParametersByPath"
    ]
    resources = [
      "arn:aws:ssm:${data.aws_region.current.name}:${data.aws_caller_identity.provisioner.account_id}:parameter${var.github_token_ssm_parameter_path}"
    ]
  }
}

data "aws_iam_policy_document" "s3_read" {
  version = "2012-10-17"
  statement {
    actions = [
      "s3:GetObject",
      "s3:ListBucket"
    ]
    resources = [
      "arn:aws:s3:::${var.s3_bucket_genesis_name}",
      "arn:aws:s3:::${var.s3_bucket_genesis_name}/*"
    ]
  }
}

data "aws_iam_policy_document" "s3_write" {
  version = "2012-10-17"
  statement {
    actions = [
      "s3:PutObject",
      "s3:DeleteObject"
    ]
    resources = [
      "arn:aws:s3:::${var.s3_bucket_genesis_name}",
      "arn:aws:s3:::${var.s3_bucket_genesis_name}/*"
    ]
  }
}

data "aws_iam_policy_document" "lambda_invoke" {
  version = "2012-10-17"
  statement {
    actions = [
      "lambda:InvokeFunction"
    ]
    resources = [
      "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.provisioner.account_id}:function:${var.genesis_init_lambda_name}"
    ]
  }
}
