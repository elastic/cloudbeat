package compliance.policy.aws_s3.data_adapter

is_s3 {
	input.subType == "aws-s3"
}

sse_algorithm := input.resource.SSEAlgorithm

bucket_policy := input.resource.BucketPolicy

bucket_policy_statement := bucket_policy.Statement[_]

bucket_versioning := input.resource.BucketVersioning
