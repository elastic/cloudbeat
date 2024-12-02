output "cloudtrail_id" {
  description = "The ID of the CloudTrail"
  value       = aws_cloudtrail.main.id
}

output "s3_bucket_name" {
  description = "The name of the S3 bucket used for CloudTrail logs"
  value       = aws_s3_bucket.cloudtrail.bucket
}

output "kms_key_id" {
  description = "The ID of the KMS key used for encryption"
  value       = aws_kms_key.cloudtrail.id
}

output "kms_alias_name" {
  description = "The name of the KMS alias"
  value       = aws_kms_alias.cloudtrail.name
}
