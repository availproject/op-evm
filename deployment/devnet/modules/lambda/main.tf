locals {
  assm_artifact_name = "assm_artifact.zip"
}

resource "null_resource" "download_release" {
  triggers = {
    always_run = timestamp()
  }
  provisioner "local-exec" {
    command = format("curl -H 'Authorization: token %s' -H 'Accept:application/octet-stream' -L -o %s %s",
      var.github_token,
      local.assm_artifact_name,
      var.assm_artifact_url)
  }
}

data "local_file" "lambda_zip_file" {
  filename   = local.assm_artifact_name
  depends_on = [null_resource.download_release]
}

resource "aws_lambda_function" "genesis_init" {
  function_name = "genesis_init-${var.deployment_name}"
  role          = var.iam_role_arn
  runtime       = "go1.x"
  handler       = "assm"
  timeout       = 20

  filename         = local.assm_artifact_name
  source_code_hash = sha256(data.local_file.lambda_zip_file.content_base64)

  environment {
    variables = {
      SSM_PARAM_PATH = var.nodes_secrets_ssm_parameter_path
      SSM_NAMESPACE  = var.ssm_namespace
      S3_BUCKET_NAME = aws_s3_bucket.genesis.bucket
      TOTAL_NODES    = var.total_nodes
    }
  }
}

resource "aws_s3_bucket" "genesis" {
  bucket_prefix = "${var.genesis_bucket_prefix}-${var.deployment_name}"
  force_destroy = true
}
