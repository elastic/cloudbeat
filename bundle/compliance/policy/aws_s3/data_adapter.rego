package compliance.policy.aws_s3.data_adapter

is_s3 {
	input.subType == "aws-s3"
}

sse_algorithm := input.resource.sse_algorithm

bucket_policy := input.resource.bucket_policy

bucket_policy_statements := object.get(bucket_policy, "Statement", [])

bucket_versioning := input.resource.bucket_versioning
