output "s3_bucket_genesis_name" {
  value = aws_s3_bucket.genesis.bucket
}

output "genesis_init_lambda_name" {
  value = aws_lambda_function.genesis_init.function_name
}
